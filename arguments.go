package main

import (
	"bufio"
	"log/slog"
	"os"
	"strings"

	"github.com/akamensky/argparse"
)

var arguments struct {
	Threads     int
	S3AccessKey string
	S3SecretKey string
	S3CredsFile string
	Config      string
	Debug       bool
}

func argumentParsing(args []string) {
	// Create new parser object
	parser := argparse.NewParser("warchangel", "upload WARC files to the Internet Archive")

	threads := parser.Int("t", "threads", &argparse.Options{
		Required: false,
		Default:  4,
		Help:     "Number of parallel uploads"})

	S3AccessKey := parser.String("", "s3-access-key", &argparse.Options{
		Required: false,
		Help:     "S3 access key"})

	S3SecretKey := parser.String("", "s3-secret-key", &argparse.Options{
		Required: false,
		Help:     "S3 secret key"})

	S3CredsFile := parser.String("", "s3-creds-file", &argparse.Options{
		Required: false,
		Help:     "S3 credentials file"})

	config := parser.String("c", "config", &argparse.Options{
		Required: true,
		Default:  "",
		Help:     "Configuration file (can either be a legacy YAML draintasker configuration file or a warchangel JSON configuration)"})

	debug := parser.Flag("d", "debug", &argparse.Options{
		Required: false,
		Help:     "Enable debug mode"})

	// Parse input
	err := parser.Parse(args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		slog.Error(parser.Usage(err))
		os.Exit(0)
	}

	// Finally save the collected flags
	arguments.Threads = *threads
	arguments.S3AccessKey = *S3AccessKey
	arguments.S3SecretKey = *S3SecretKey
	arguments.S3CredsFile = *S3CredsFile
	arguments.Config = *config
	arguments.Debug = *debug

	// Load S3 credentials from file if specified
	if arguments.S3AccessKey == "" || arguments.S3SecretKey == "" {
		// Get the .ias3cfg in the $HOME directory
		arguments.S3CredsFile = os.ExpandEnv("$HOME/.ias3cfg")

		arguments.S3AccessKey, arguments.S3SecretKey, err = loadS3CredsFromFile(arguments.S3CredsFile)
		if err != nil {
			slog.Error("Failed to load S3 credentials from file", "err", err.Error())
			os.Exit(1)
		}
	}
}

func loadS3CredsFromFile(filePath string) (string, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inDefaultSection := false
	var accessKey, secretKey string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := strings.TrimSpace(line[1 : len(line)-1])
			inDefaultSection = section == "default"
			continue
		}
		if inDefaultSection {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			switch key {
			case "access_key":
				accessKey = value
			case "secret_key":
				secretKey = value
			}
			if accessKey != "" && secretKey != "" {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", err
	}

	return accessKey, secretKey, nil
}
