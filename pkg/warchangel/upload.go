package warchangel

import (
	"context"
	"os"
	"path"

	"github.com/remeh/sizedwaitgroup"
)

func uploadFile(filename string, item string, wg *sizedwaitgroup.SizedWaitGroup) {
	defer wg.Done()

	UploadsInProgress.Store(filename, item)

	logger.Info("uploading file", "file", filename, "item", item)

	// Init Internet Archive S3 client
	fs, err := initRcloneFS(filename, item)
	if err != nil {
		logger.Error("unable to init rclone FS", "err", err)
		panic(err)
	}

	// Open file
	file, err := os.Open(path.Join(config.WARCsDir, filename))
	if err != nil {
		logger.Error("unable to open file", "err", err)
		return
	}

	// Upload file
	object, err := fs.NewObject(context.Background(), filename)
	if err != nil {
		logger.Error("unable to create new object", "err", err)
		return
	}

	object, err = fs.Put(context.Background(), file, object, nil)
	if err != nil {
		logger.Error("unable to upload file", "err", err)
		return
	}

	// Once done, remove from inProgress
	UploadsInProgress.Delete(filename)

	logger.Info("finished uploading file", "file", filename, "item", item, "path", object.Remote())
}
