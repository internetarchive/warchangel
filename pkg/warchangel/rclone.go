package warchangel

import (
	"context"
	"fmt"
	"strings"

	"github.com/rclone/rclone/backend/internetarchive"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config/configmap"
)

func initRcloneFS(filename, item string) (fs fs.Fs, err error) {
	// rcloneConfig := internetarchive.Options{
	// 	"access_key_id":     S3AccessKey,
	// 	"secret_access_key": S3SecretKey,
	// 	"item_derive":       boolToString(intToBool(config.Derive)),
	// 	"endpoint":          "https://s3.us.archive.org",
	// 	"front_endpoint":    "https://archive.org",
	// 	"disable_checksum":  "true",
	// }
	rcloneConfig := configmap.New()

	rcloneConfig.Set("access_key_id", S3AccessKey)
	rcloneConfig.Set("secret_access_key", S3SecretKey)
	rcloneConfig.Set("item_derive", boolToString(intToBool(config.Derive)))
	rcloneConfig.Set("endpoint", "https://s3.us.archive.org")
	rcloneConfig.Set("front_endpoint", "https://archive.org")
	rcloneConfig.Set("disable_checksum", "true")
	rcloneConfig.Set("wait_archive", "0")

	// Build IA's item metadata
	var itemMetadata []string
	for key, values := range config.Metadata {
		for _, value := range values {
			itemMetadata = append(itemMetadata, fmt.Sprintf("%s=%s", key, value))
		}
	}

	for _, c := range config.Collections {
		itemMetadata = append(itemMetadata, fmt.Sprintf("collection=%s", c))
	}

	// Extract metadata from filename
	parsedFilename, err := parseFilename(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to parse filename: %w", err)
	}

	itemMetadata = append(itemMetadata, fmt.Sprintf("crawler=%s", parsedFilename.Crawler))
	itemMetadata = append(itemMetadata, fmt.Sprintf("date=%s", parsedFilename.FullTimestamp[:4]))
	itemMetadata = append(itemMetadata, fmt.Sprintf("description=%s", config.Description))
	itemMetadata = append(itemMetadata, fmt.Sprintf("operator=%s", config.Operator))
	itemMetadata = append(itemMetadata, fmt.Sprintf("title=%s", config.TitlePrefix))
	itemMetadata = append(itemMetadata, fmt.Sprintf("scanner=%s", parsedFilename.FQDN))

	rcloneConfig.Set("metadata", strings.Join(itemMetadata, ","))

	fs, err = internetarchive.NewFs(context.Background(), filename, item, rcloneConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create Internet Archive S3 client: %w", err)
	}

	fmt.Printf("root %s", fs.Root())

	return fs, nil
}
