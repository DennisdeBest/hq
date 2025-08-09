package extractor

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

const testHTML = `<div class="container" id="main">
  <h1>Title</h1>
  <p>Some text with <a href="https://example.com">a link</a> inside.</p>
  <ul>
    <li>Item 1</li>
    <li>Item 2</li>
  </ul>
</div>`

func TestExtractAsJSON(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		selector string
		expected JSONOutput
	}{
		{
			name:     "Extract div with nested structure",
			html:     testHTML,
			selector: "div.container",
			expected: JSONOutput{
				Selector: "div.container",
				Count:    1,
				Results: []JSONElement{
					{
						Tag: "div",
						Attributes: map[string]string{
							"class": "container",
							"id":    "main",
						},
						Children: []JSONElement{
							{
								Tag:  "h1",
								Text: "Title",
							},
							{
								Tag:  "p",
								Text: "Some text with inside.",
								Children: []JSONElement{
									{
										Tag: "a",
										Attributes: map[string]string{
											"href": "https://example.com",
										},
										Text: "a link",
									},
								},
							},
							{
								Tag: "ul",
								Children: []JSONElement{
									{
										Tag:  "li",
										Text: "Item 1",
									},
									{
										Tag:  "li",
										Text: "Item 2",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:     "Extract multiple elements",
			html:     `<div><p>First</p></div><div><p>Second</p></div>`,
			selector: "div",
			expected: JSONOutput{
				Selector: "div",
				Count:    2,
				Results: []JSONElement{
					{
						Tag: "div",
						Children: []JSONElement{
							{
								Tag:  "p",
								Text: "First",
							},
						},
					},
					{
						Tag: "div",
						Children: []JSONElement{
							{
								Tag:  "p",
								Text: "Second",
							},
						},
					},
				},
			},
		},
		{
			name:     "Extract element with only text",
			html:     `<h1>Simple heading</h1>`,
			selector: "h1",
			expected: JSONOutput{
				Selector: "h1",
				Count:    1,
				Results: []JSONElement{
					{
						Tag:  "h1",
						Text: "Simple heading",
					},
				},
			},
		},
		{
			name:     "Extract element with attributes but no text",
			html:     `<img src="test.jpg" alt="Test image">`,
			selector: "img",
			expected: JSONOutput{
				Selector: "img",
				Count:    1,
				Results: []JSONElement{
					{
						Tag: "img",
						Attributes: map[string]string{
							"src": "test.jpg",
							"alt": "Test image",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.html)
			output := &bytes.Buffer{}

			config := Config{
				Selector: tt.selector,
				JSON:     true,
				Input:    input,
				Output:   output,
			}

			err := Extract(config)
			if err != nil {
				t.Fatalf("Extract() error = %v", err)
			}

			var result JSONOutput
			err = json.Unmarshal(output.Bytes(), &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v\nOutput: %s", err, output.String())
			}

			// Compare the results
			if !compareJSONOutput(result, tt.expected) {
				t.Errorf("Extract() result mismatch")
				expectedJSON, _ := json.MarshalIndent(tt.expected, "", "  ")
				actualJSON, _ := json.MarshalIndent(result, "", "  ")
				t.Errorf("Expected:\n%s", expectedJSON)
				t.Errorf("Got:\n%s", actualJSON)
			}
		})
	}
}

func TestExtractAsText(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		selector string
		keepTags bool
		expected []string // Use slice of expected strings for flexible matching
	}{
		{
			name:     "Extract text without tags",
			html:     `<div><p>Hello <strong>world</strong>!</p></div>`,
			selector: "div",
			keepTags: false,
			expected: []string{"Hello world!"},
		},
		{
			name:     "Extract with tags preserved",
			html:     `<div><p>Hello</p></div>`,
			selector: "div",
			keepTags: true,
			expected: []string{"<div>", "<p>", "Hello", "</p>", "</div>"},
		},
		{
			name:     "Extract multiple elements as text",
			html:     `<p>First paragraph</p><p>Second paragraph</p>`,
			selector: "p",
			keepTags: false,
			expected: []string{"First paragraph", "Second paragraph"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.html)
			output := &bytes.Buffer{}

			config := Config{
				Selector:         tt.selector,
				KeepTags:         tt.keepTags,
				JSON:             false,
				MonochromeOutput: true, // Disable color for consistent testing
				Input:            input,
				Output:           output,
			}

			err := Extract(config)
			if err != nil {
				t.Fatalf("Extract() error = %v", err)
			}

			result := output.String()

			// Check that all expected strings are present
			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Extract() text output missing expected content")
					t.Errorf("Expected to contain: %q", expected)
					t.Errorf("Got: %q", result)
					break
				}
			}
		})
	}
}

func TestCleanTextPreserveLines(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "Simple HTML with block elements",
			html:     `<div><p>First paragraph</p><p>Second paragraph</p></div>`,
			expected: "First paragraph\n\nSecond paragraph",
		},
		{
			name:     "HTML with inline elements",
			html:     `<p>Text with <strong>bold</strong> and <em>italic</em></p>`,
			expected: "Text with bold and italic",
		},
		{
			name:     "HTML with line breaks",
			html:     `<p>Line 1<br>Line 2<br/>Line 3</p>`,
			expected: "Line 1\n\nLine 2\n\nLine 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanTextPreserveLines(tt.html)
			if result != tt.expected {
				t.Errorf("cleanTextPreserveLines() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// Helper function to compare JSONOutput structs
func compareJSONOutput(a, b JSONOutput) bool {
	if a.Selector != b.Selector || a.Count != b.Count {
		return false
	}
	if len(a.Results) != len(b.Results) {
		return false
	}
	for i, result := range a.Results {
		if !compareJSONElement(result, b.Results[i]) {
			return false
		}
	}
	return true
}

// Helper function to compare JSONElement structs recursively
func compareJSONElement(a, b JSONElement) bool {
	if a.Tag != b.Tag || a.Text != b.Text {
		return false
	}

	// Compare attributes (handle nil maps)
	aAttrs := a.Attributes
	bAttrs := b.Attributes
	if aAttrs == nil {
		aAttrs = make(map[string]string)
	}
	if bAttrs == nil {
		bAttrs = make(map[string]string)
	}

	if len(aAttrs) != len(bAttrs) {
		return false
	}
	for key, value := range aAttrs {
		if bAttrs[key] != value {
			return false
		}
	}

	// Compare children
	if len(a.Children) != len(b.Children) {
		return false
	}
	for i, child := range a.Children {
		if !compareJSONElement(child, b.Children[i]) {
			return false
		}
	}

	return true
}
