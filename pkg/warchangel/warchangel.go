package warchangel

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/remeh/sizedwaitgroup"
)

var (
	UploadsInProgress sync.Map
	S3AccessKey       string
	S3SecretKey       string
	logger            *slog.Logger
	config            *Config
)

func NewWatcher(c *Config, l *slog.Logger, uploadThreads int, s3AccessKey, s3SecretKey string, doneChan chan struct{}) error {
	var (
		currentItem     string
		currentItemSize int64
		itemCount       = 1
		wg              = sizedwaitgroup.New(uploadThreads)
	)

	// Set global variables
	S3AccessKey = s3AccessKey
	S3SecretKey = s3SecretKey
	logger = l
	config = c

	logger.Info("starting watcher", "path", config.WARCsDir, "interval", config.ScanInterval)
	ticker := time.NewTicker(time.Duration(config.ScanInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-doneChan:
			logger.Info("stopping watcher, waiting for uploads to finish")
			wg.Wait()
			logger.Info("all uploads finished, exiting watcher")
			return nil
		case <-ticker.C:
			logger.Debug("watching", "path", config.WARCsDir)

			// Read directory
			files, err := os.ReadDir(config.WARCsDir)
			if err != nil {
				logger.Error("error reading directory", "err", err)
				continue
			}

			// Iterate over files
			for _, file := range files {
				if file.IsDir() {
					continue
				}

				name := file.Name()

				// Check if it's a WARC file
				if !(strings.HasSuffix(name, ".warc.zst") || strings.HasSuffix(name, ".warc.gz")) {
					continue
				}

				// Check if already uploading
				if _, ok := UploadsInProgress.Load(name); ok {
					continue
				}

				// Init the item name if it's the first upload
				if currentItem == "" {
					currentItem, err = getItemName(name)
					if err != nil {
						panic(err)
					}
				}

				// Get file size
				fullPath := filepath.Join(config.WARCsDir, name)
				info, err := os.Stat(fullPath)
				if err != nil {
					logger.Error("unable to stat file", "file", name, "err", err)
					continue
				}

				size := info.Size()

				// Check if adding this file exceeds the current item size limit
				if currentItemSize+size > int64(config.ItemSize*1024) {
					// finalize current item (in real logic, you might have to mark item completed)
					logger.Info("item size limit reached, starting new item")

					itemCount++
					currentItem, err = getItemName(name)
					if err != nil {
						panic(err)
					}

					currentItemSize = 0
				}

				// Add this file to current item and start uploading
				currentItemSize += size

				// Start upload
				wg.Add()
				go uploadFile(name, currentItem, &wg)
			}
		}
	}
}
