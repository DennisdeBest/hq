package cmd

import (
	"gurl/pkg/formatters"
	"gurl/pkg/styles"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag/v2"
	"gurl/pkg/extractor"
)

var (
	keepTags         bool
	jsonOutput       bool
	colorOutput      bool
	monochromeOutput bool
	style            string
	formatter        formatters.Formatter
	RootCmd          = &cobra.Command{
		Use:   "hq [options] <css-selector> [files ...]",
		Short: "A tool to extract content from HTML using CSS selectors",
		Long: `hq allows you to pipe HTML content and extract data using CSS selectors.
Example: curl -s https://example.com | hq 'div.content'`,
		Args: cobra.MinimumNArgs(1),
		RunE: runExtract,
	}
)

func init() {
	formatterEnumFlag := enumflag.New(&formatter, "", formatters.FormatterIds, enumflag.EnumCaseInsensitive)

	RootCmd.Flags().BoolVarP(&keepTags, "keep-tags", "k", false, "Keep HTML tags in the output")
	RootCmd.Flags().BoolVarP(&monochromeOutput, "monochrome-output", "M", false, "Disable colored output")
	RootCmd.Flags().BoolVarP(&colorOutput, "color-output", "C", false, "Colorize output")
	RootCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	RootCmd.Flags().StringVarP(&style, "style", "s", styles.Default, "The chroma style for the output. See available styles at https://xyproto.github.io/splash/docs/all.html")
	RootCmd.Flags().VarP(
		formatterEnumFlag,
		"formatter", "f", "The chroma formatter used, set to 'terminal16' if using 'watch -c'")

	formatterEnumFlag.RegisterCompletion(RootCmd, "formatter", formatters.Help())
}

func runExtract(cmd *cobra.Command, args []string) error {
	selector := args[0]
	files := args[1:]
	if len(files) == 0 {
		files = []string{"-"} // default: stdin
	}

	for _, path := range files {
		var r io.ReadCloser
		if path == "-" {
			r = io.NopCloser(os.Stdin) // wraps Stdin so we can Close()
		} else {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			r = f
		}

		cfg := extractor.Config{
			Selector:         selector,
			KeepTags:         keepTags,
			JSON:             jsonOutput,
			MonochromeOutput: monochromeOutput,
			ColorOutput:      colorOutput,
			Style:            style,
			Formatter:        formatter,
			Input:            r,
			Output:           os.Stdout,
		}
		if err := extractor.Extract(cfg); err != nil {
			r.Close()
			return err
		}
		r.Close()
	}
	return nil
}
