package warchangel

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// WARCNaming represents the WARC naming convention
type WARCNaming uint8

const (
	ZenoWARCNaming     WARCNaming = 1 // {TLA}-{timestamp}-{serial}-{fqdn}.warc.gz
	HeritrixWARCNaming            = 2 // {TLA}-{timestamp}-{serial}-{PID}~{fqdn}~{port}.warc.gz
)

// ConfigFormat is an enum representing supported configuration formats.
type ConfigFormat int

const (
	FormatUnknown ConfigFormat = iota
	FormatJSON
	FormatYAML
)

// String provides a string representation for ConfigFormat values.
func (f ConfigFormat) String() string {
	switch f {
	case FormatJSON:
		return "json"
	case FormatYAML:
		return "yaml"
	default:
		return "unknown"
	}
}

type Config struct {
	// Job name
	Job string `json:"job"`
	// Directory where the WARCs are stored
	WARCsDir string `json:"warcs"`
	// ScanInterval is the number of seconds between each scan of the WARCs directory
	ScanInterval int `json:"scan_interval"`
	// Target item size in gigabytes
	ItemSize int `json:"item_size"`
	// WARC naming convention
	WARCNaming WARCNaming `json:"warc_naming"`
	// Description inserted in the item's metadata
	Description string `json:"description"`
	// Operator inserted in the item's metadata
	Operator string `json:"operator"`
	// Collections to which the item belongs
	Collections []string `json:"collections"`
	// Prefix inserted in the item's title
	TitlePrefix string `json:"title_prefix"`
	// Metadata to be inserted in the item's metadata
	Metadata map[string][]string `json:"subject"`
	// Derive flag, if set to 0 the item will not be derived
	Derive int `json:"derive"`
}

type draintaskerConfig struct {
	Crawljob       string              `yaml:"crawljob"`
	JobDir         string              `yaml:"job_dir"`
	XferDir        string              `yaml:"xfer_dir"`
	VerifyGzip     int                 `yaml:"verify_gzip"`
	Md5sum         int                 `yaml:"md5sum"`
	SleepTime      int                 `yaml:"sleep_time"`
	MaxSize        int                 `yaml:"max_size"`
	WARCNaming     int                 `yaml:"WARC_naming"`
	BlockDelay     int                 `yaml:"block_delay"`
	MaxBlockCount  int                 `yaml:"max_block_count"`
	RetryDelay     int                 `yaml:"retry_delay"`
	Description    string              `yaml:"description"`
	Operator       string              `yaml:"operator"`
	Collections    []string            `yaml:"collections"`
	TitlePrefix    string              `yaml:"title_prefix"`
	Creator        string              `yaml:"creator"`
	Sponsor        string              `yaml:"sponsor"`
	Contributor    string              `yaml:"contributor"`
	ScanningCenter string              `yaml:"scanningcenter"`
	Metadata       map[string][]string `yaml:"metadata"`
	Derive         int                 `yaml:"derive"`
	CompactNames   int                 `yaml:"compact_names"`
}

// DetectConfigFormat attempts to determine if data is JSON or YAML.
// Returns the appropriate ConfigFormat and an error if neither is detected.
func DetectConfigFormat(path string) (ConfigFormat, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FormatUnknown, err
	}

	var jsonObj interface{}
	if err := json.Unmarshal(data, &jsonObj); err == nil {
		return FormatJSON, nil
	}

	var yamlObj interface{}
	if err := yaml.Unmarshal(data, &yamlObj); err == nil {
		return FormatYAML, nil
	}

	return FormatUnknown, fmt.Errorf("data is neither valid JSON nor YAML")
}

// LoadConfig loads a warchangel configuration file (JSON)
func LoadConfig(path string) (c *Config, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadDraintaskerConfig loads a Draintasker configuration file (YAML)
func LoadDraintaskerConfig(path string) (c *Config, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var dtCfg draintaskerConfig
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&dtCfg); err != nil {
		return nil, err
	}

	// Transform Draintasker configuration into warchangel configuration
	cfg := Config{
		Job:          dtCfg.Crawljob,
		WARCsDir:     dtCfg.JobDir,
		ScanInterval: dtCfg.SleepTime,
		ItemSize:     dtCfg.MaxSize,
		WARCNaming:   WARCNaming(dtCfg.WARCNaming),
		Description:  dtCfg.Description,
		Collections:  dtCfg.Collections,
		TitlePrefix:  dtCfg.TitlePrefix,
		Metadata:     dtCfg.Metadata,
		Derive:       dtCfg.Derive,
	}

	cfg.Metadata["creator"] = []string{dtCfg.Creator}
	cfg.Metadata["sponsor"] = []string{dtCfg.Sponsor}
	cfg.Metadata["contributor"] = []string{dtCfg.Contributor}
	cfg.Metadata["scanningcenter"] = []string{dtCfg.ScanningCenter}
	cfg.Metadata["operator"] = []string{dtCfg.Operator}

	return &cfg, nil
}
