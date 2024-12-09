package warchangel

import (
	"errors"
	"fmt"
	"strings"
)

// ParsedFilename holds the extracted components from a WARC filename.
type ParsedFilename struct {
	TLA           string // Prefix
	Timestamp     string // 14-digit timestamp
	FullTimestamp string
	Serial        string // Serial number
	PID           string // Process ID (Heritrix)
	FQDN          string // Fully Qualified Domain Name
	Host          string // Hostname
	Port          string // Port number (Heritrix)
	Original      string // Original filename without extension
	FullName      string // Full original filename
	Crawler       string
}

// parseFilename extracts all parts from the filename based on the WARC naming convention.
func parseFilename(filename string) (*ParsedFilename, error) {
	fullName := filename

	// Remove extensions
	base := strings.TrimSuffix(filename, ".warc.zst")
	base = strings.TrimSuffix(base, ".warc.gz")

	parts := strings.Split(base, "-")
	if len(parts) < 4 {
		return nil, errors.New("filename does not have the expected format")
	}

	tla := parts[0]
	timestamp := parts[1]
	serial := parts[2]

	if len(timestamp) < 14 {
		return nil, errors.New("timestamp is shorter than 14 digits")
	}
	truncatedTimestamp := timestamp[:14]

	parsed := &ParsedFilename{
		TLA:           tla,
		Timestamp:     truncatedTimestamp,
		FullTimestamp: timestamp,
		Serial:        serial,
		Original:      base,
		FullName:      fullName,
	}

	switch config.WARCNaming {
	case ZenoWARCNaming:
		// Zeno Naming: {TLA}-{timestamp}-{serial}-{fqdn}.warc.gz
		parsed.FQDN = parts[3]
		parsed.Host = strings.Split(parsed.FQDN, ".")[0]
		parsed.Crawler = "Zeno"
	case HeritrixWARCNaming:
		// Heritrix Naming: {TLA}-{timestamp}-{serial}-{PID}~{fqdn}~{port}.warc.gz
		if len(parts) < 4 {
			return nil, errors.New("Heritrix filename does not have the expected format")
		}
		lastPart := parts[3]
		subparts := strings.Split(lastPart, "~")
		if len(subparts) < 3 {
			return nil, errors.New("Heritrix filename does not have the expected format")
		}
		parsed.PID = subparts[0]
		parsed.FQDN = subparts[1]
		parsed.Host = strings.Split(parsed.FQDN, ".")[0]
		parsed.Port = subparts[2]
		parsed.Crawler = "Heritrix"
	default:
		return nil, errors.New("unknown WARC naming convention")
	}

	return parsed, nil
}

func getItemName(filename string) (string, error) {
	parsed, err := parseFilename(filename)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s-%s", parsed.TLA, parsed.Timestamp, parsed.Host), nil
}
