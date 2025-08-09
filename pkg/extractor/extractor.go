package extractor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	localFormatters "gurl/pkg/formatters"
	localStyles "gurl/pkg/styles"

	"github.com/PuerkitoBio/goquery"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/grokify/html-strip-tags-go"
	"github.com/yosssi/gohtml"
)

type Config struct {
	Selector         string
	KeepTags         bool
	JSON             bool
	MonochromeOutput bool
	ColorOutput      bool
	Style            string
	Formatter        localFormatters.Formatter
	Input            io.Reader
	Output           io.Writer
}

type JSONElement struct {
	Tag        string            `json:"tag"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Text       string            `json:"text,omitempty"`
	Children   []JSONElement     `json:"children,omitempty"`
}

type JSONOutput struct {
	Selector string        `json:"selector"`
	Results  []JSONElement `json:"results"`
	Count    int           `json:"count"`
}

func Extract(config Config) error {
	doc, err := goquery.NewDocumentFromReader(config.Input)
	if err != nil {
		return fmt.Errorf("error parsing HTML: %v", err)
	}

	if config.JSON {
		return extractAsJSON(doc, config)
	}

	return extractAsText(doc, config)
}

func extractAsJSON(doc *goquery.Document, config Config) error {
	var results []JSONElement

	doc.Find(config.Selector).Each(func(i int, s *goquery.Selection) {
		element := selectionToJSONElement(s)
		results = append(results, element)
	})

	output := JSONOutput{
		Selector: config.Selector,
		Results:  results,
		Count:    len(results),
	}

	// Format JSON with proper indentation
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(output)
	if err != nil {
		return err
	}

	jsonString := buf.String()

	useColor := config.ColorOutput || (!config.MonochromeOutput && isTerminal())
	if useColor {
		if err := format(config.Output, jsonString, "json", config); err != nil {
			fmt.Fprint(config.Output, jsonString)
		}
	} else {
		fmt.Fprint(config.Output, jsonString)
	}

	return nil
}

func selectionToJSONElement(s *goquery.Selection) JSONElement {
	// Get the tag name
	tagName := goquery.NodeName(s)

	element := JSONElement{
		Tag:        tagName,
		Attributes: make(map[string]string),
	}

	// Extract attributes
	if s.Get(0) != nil {
		for _, attr := range s.Get(0).Attr {
			element.Attributes[attr.Key] = attr.Val
		}
	}

	// Remove attributes if empty
	if len(element.Attributes) == 0 {
		element.Attributes = nil
	}

	// Get child elements
	var children []JSONElement

	// Use Children() instead of Contents() to get only element nodes (not text nodes)
	s.Children().Each(func(i int, child *goquery.Selection) {
		childElement := selectionToJSONElement(child)
		children = append(children, childElement)
	})

	// Get the direct text content (not from children)
	// This gets text that is directly inside this element
	directText := s.Clone().Children().Remove().End().Text()
	directText = strings.TrimSpace(directText)

	// Clean up multiple whitespace
	if directText != "" {
		re := regexp.MustCompile(`\s+`)
		directText = re.ReplaceAllString(directText, " ")
		element.Text = directText
	}

	// Set children
	if len(children) > 0 {
		element.Children = children
	}

	return element
}

func extractAsText(doc *goquery.Document, config Config) error {
	doc.Find(config.Selector).Each(func(i int, s *goquery.Selection) {
		html, err := goquery.OuterHtml(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error getting HTML for element %d: %v\n", i, err)
			return
		}

		if config.KeepTags {
			formatted := gohtml.Format(html)

			useColor := config.ColorOutput || (!config.MonochromeOutput && isTerminal())
			if useColor {
				if err := format(config.Output, formatted, "html", config); err != nil {
					fmt.Fprintln(config.Output, formatted)
				}
			} else {
				fmt.Fprintln(config.Output, formatted)
			}
			return
		}
		cleaned := cleanTextPreserveLines(html)
		if cleaned != "" {
			fmt.Fprintln(config.Output, cleaned)
		}
	})
	return nil
}

func isTerminal() bool {
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}
	return false
}

func cleanTextPreserveLines(html string) string {
	blockElements := []string{"</div>", "</p>", "</h1>", "</h2>", "</h3>", "</h4>", "</h5>", "</h6>",
		"</li>", "</tr>", "</td>", "</th>", "</section>", "</article>", "</header>", "</footer>",
		"</nav>", "</main>", "</aside>", "<br>", "<br/>", "<br />"}

	text := html
	for _, element := range blockElements {
		text = strings.ReplaceAll(text, element, element+"\n")
	}

	text = strip.StripTags(text)

	lines := strings.Split(text, "\n")
	var cleanLines []string

	for _, line := range lines {
		re := regexp.MustCompile(`\s+`)
		line = re.ReplaceAllString(line, " ")
		line = strings.TrimSpace(line)

		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	result := strings.Join(cleanLines, "\n")

	// Add double newlines after block elements for better readability
	result = strings.ReplaceAll(result, "\n", "\n\n")

	// Remove excessive consecutive newlines (more than 2)
	re2 := regexp.MustCompile(`\n{3,}`)
	result = re2.ReplaceAllString(result, "\n\n")

	return strings.TrimSpace(result)
}

func format(w io.Writer, src, lang string, config Config) error {
	lex := lexers.Get(lang)
	if lex == nil {
		lex = lexers.Fallback
	}
	iter, _ := lex.Tokenise(nil, src)

	styles.Fallback = styles.Get(localStyles.Default)

	style := styles.Get(config.Style)

	var formatter chroma.Formatter

	if config.Formatter == localFormatters.AutoDetect {
		formatter = pickFormatter()
	} else {
		formatter = formatters.Get(config.Formatter.String())
	}

	return formatter.Format(w, style, iter)
}

func pickFormatter() chroma.Formatter {
	if tc := strings.ToLower(os.Getenv("COLORTERM")); strings.Contains(tc, "truecolor") || strings.Contains(tc, "24bit") {
		return formatters.Get("terminal16m")
	}

	if strings.Contains(os.Getenv("TERM"), "256color") {
		return formatters.Get("terminal256")
	}

	return formatters.Get("terminal8")
}
