// upload_test.go
package warchangel

import (
	"testing"
)

func TestGetItemName(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		config      *Config
		expected    string
		expectError bool
	}{
		{
			name:     "Valid Zeno WARC filename",
			filename: "WEB-20240109170659538-00001-endgame.local.warc.gz",
			config: &Config{
				WARCNaming: ZenoWARCNaming,
			},
			expected:    "WEB-20240109170659-endgame.local",
			expectError: false,
		},
		{
			name:     "Valid Heritrix WARC filename",
			filename: "WEB-20240109170659538-00001-12345~endgame.local~80.warc.gz",
			config: &Config{
				WARCNaming: HeritrixWARCNaming,
			},
			expected:    "WEB-20240109170659-endgame.local",
			expectError: false,
		},
		{
			name:        "Invalid filename format (missing parts)",
			filename:    "WEB-20240109170659.warc.gz",
			config:      &Config{WARCNaming: ZenoWARCNaming},
			expected:    "",
			expectError: true,
		},
		{
			name:        "Timestamp shorter than 14 digits",
			filename:    "WEB-20240109-00001-endgame.local.warc.gz",
			config:      &Config{WARCNaming: ZenoWARCNaming},
			expected:    "",
			expectError: true,
		},
		{
			name:        "Unknown WARC naming convention",
			filename:    "WEB-20240109170659538-00001-endgame.local.warc.gz",
			config:      &Config{WARCNaming: 255},
			expected:    "",
			expectError: true,
		},
		{
			name:     "Zeno WARC filename with different serial",
			filename: "API-20231231235959123-99999-service.local.warc.zst",
			config: &Config{
				WARCNaming: ZenoWARCNaming,
			},
			expected:    "API-20231231235959-service.local",
			expectError: false,
		},
		{
			name:     "Heritrix WARC filename with different PID and port",
			filename: "IMG-20230515123045000-00001-12345~imageserver.local~8080.warc.gz",
			config: &Config{
				WARCNaming: HeritrixWARCNaming,
			},
			expected:    "IMG-20230515123045-imageserver.local",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			itemName, err := getItemName(tc.filename)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none for test case: %s\nReturned: %s", tc.name, itemName)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error but got: %v for test case: %s", err, tc.name)
				}
				if itemName != tc.expected {
					t.Errorf("Expected item name '%s', but got '%s' for test case: %s", tc.expected, itemName, tc.name)
				}
			}
		})
	}
}
