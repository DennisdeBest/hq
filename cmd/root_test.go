package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"gurl/pkg/extractor"
)

func TestRootCommandIntegration(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		args     []string
		wantJSON bool
		selector string
	}{
		{
			name:     "Basic text extraction",
			html:     `<div class="test">Hello World</div>`,
			args:     []string{"div.test"},
			wantJSON: false,
			selector: "div.test",
		},
		{
			name:     "JSON extraction",
			html:     `<div class="test">Hello World</div>`,
			args:     []string{"-j", "div.test"},
			wantJSON: true,
			selector: "div.test",
		},
		{
			name:     "JSON with keep tags",
			html:     `<div class="test"><p>Hello World</p></div>`,
			args:     []string{"-j", "-k", "div.test"},
			wantJSON: true,
			selector: "div.test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags to default values
			keepTags = false
			jsonOutput = false
			monochromeOutput = false

			input := strings.NewReader(tt.html)
			output := &bytes.Buffer{}

			config := extractor.Config{
				Selector: tt.selector,
				Input:    input,
				Output:   output,
			}

			// Parse flags from args
			for _, arg := range tt.args {
				switch arg {
				case "-j", "--json":
					config.JSON = true
				case "-k", "--keep-tags":
					config.KeepTags = true
				case "-M", "--monochrome-output":
					config.MonochromeOutput = true
				default:
					// If it's not a flag, it's the selector
					if !strings.HasPrefix(arg, "-") {
						config.Selector = arg
					}
				}
			}

			err := extractor.Extract(config)
			if err != nil {
				t.Fatalf("Extract() error = %v", err)
			}

			result := output.String()

			if tt.wantJSON {
				// Verify it's valid JSON
				var jsonResult extractor.JSONOutput
				err = json.Unmarshal([]byte(result), &jsonResult)
				if err != nil {
					t.Errorf("Expected valid JSON, got error: %v", err)
					t.Errorf("Output: %s", result)
				}

				// Verify selector matches
				if jsonResult.Selector != tt.selector {
					t.Errorf("Expected selector %q, got %q", tt.selector, jsonResult.Selector)
				}
			} else {
				// For text output, just verify we got some content
				if strings.TrimSpace(result) == "" {
					t.Errorf("Expected text output, got empty string")
				}
			}
		})
	}
}
