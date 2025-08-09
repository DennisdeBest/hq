# hq

**hq** is a small command‑line tool that lets you pipe (or pass) HTML and extract elements with a familiar **CSS selector** syntax. Think of it as *jq for HTML*‑–with colorised output, JSON export, and first‑class support for shell completion.

> `curl -s https://example.com | hq 'div.article h2.title' -C`

<!-- TOC -->
* [hq](#hq)
  * [Features](#features)
  * [Installation](#installation)
    * [Quick Install (Linux and macOS)](#quick-install-linux-and-macos)
    * [Manual Installation](#manual-installation)
  * [Usage](#usage)
    * [Output formats](#output-formats)
      * [Default (plain text)](#default-plain-text)
      * [Keep tags (`-k` or `--keep-tags`)](#keep-tags--k-or---keep-tags)
      * [JSON output (`-j` or `--json`)](#json-output--j-or---json)
    * [Watching live output](#watching-live-output)
    * [Piping to jq](#piping-to-jq)
    * [More examples](#more-examples)
  * [Shell completion](#shell-completion)
  * [Color styles and formatters](#color-styles-and-formatters)
    * [Export formatters](#export-formatters)
<!-- TOC -->

---

## Features

* **Selector first** – the first positional argument is the CSS selector you want to extract.
* **HTML from anywhere** – read from *stdin* by default or supply one or more filenames.
* **Pretty output**
    * keep original tags (`-k`) or strip to plain text
    * JSON dump of the matched nodes (`-j`)
    * full ANSI color with automatic formatter detection, or force a specific formatter with `-f`.
* **Shell completion** out‑of‑the‑box (bash, zsh, fish, PowerShell).
* Single static binary, no dependencies.

---

![CLI demo](docs/demo.gif)

### Why another HTML-selector tool?

I mainly built **hq** to dive deeper into how parsers, selectors and syntax-highlighting work in Go, and that learning exercise turned into a tool with a few features I couldn’t find elsewhere.
Key practical differences compared with `pup`, `htmlq`, `cheerio-cli`, … :
- **First-class color output** – any element you extract can be syntax-highlighted via Chroma (true-color, 256-color, 16-color, HTML, SVG).
- **Selector-first CLI** – no sub-commands; the very first argument is always the CSS selector, which feels closer to how `jq` works.
- **Modern CSS support** – ships with `:has()`, `:nth-child()`, attribute selectors, etc., courtesy of the Go `cascadia` matcher.
- **Shell completion out of the box** – generates bash / zsh / fish / PowerShell completion.
- **Single static binary** – cross-compiles to a ~8 MB executable with zero runtime dependencies.
- **JSON export designed for piping** – count, text, raw HTML and attributes are all present, so you rarely need a second pass over the HTML.

## Installation

### Quick Install (Linux and macOS)

```bash
# Install the latest version with one command
curl -sSL https://raw.githubusercontent.com/dennisdebest/hq/master/install.sh | bash
```

This script will:
- Detect your OS and architecture
- Download the appropriate binary from the latest GitHub release
- Install it to `/usr/local/bin` (or `~/.local/bin` if you don't have write permissions)
- Add the installation directory to your PATH if needed

### Manual Installation

```bash
# Go ≥1.22
go install github.com/dennisdebest/hq@latest

# or build manually (stripped static binary ~5 MB)
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o hq .
```

---

## Usage

```text
hq [options] <css-selector> [files ...]

Options
  -k, --keep-tags            keep HTML tags in the output
  -j, --json                 output as JSON
  -C, --color-output         force color even when not on a TTY
  -M, --monochrome-output    disable color entirely
  -s, --style <name>         Chroma color style (default: monokai)
  -f, --formatter <name>     Chroma formatter (auto|terminal|terminal16|…)
```

If no *files* are given, **hq** reads from *stdin* – perfect for piping:

```bash
curl -s https://www.timeanddate.com/worldclock/ \
  | hq 'section.fixed td:has(a[href="/worldclock/france/paris"]) + td' -j -C
```

Read multiple files in one go (each printed in order):

```bash
hq '.post h2' index.html archive.html
```

Force stdin explicitly with a single dash:

```bash
cat page.html | hq '.sidebar' -C -           # "-" is treated as stdin
```

### Output formats

**hq** supports three different output formats depending on the flags you use:

#### Default (plain text)

Extracts just the text content from matched elements:

```shell
echo '<div><h1>Hello World</h1><p>This is a paragraph.</p></div>' | hq 'h1, p'
```

```text

Hello World
This is a paragraph.
```
#### Keep tags (`-k` or `--keep-tags`)

Preserves the original HTML tags:

```shell
echo '<div><h1 class="title">Hello World</h1><p>This is a paragraph.</p></div>' | hq 'h1, p' -k
```

```html
<h1 class="title">Hello World</h1>
<p>This is a paragraph.</p>
```

#### JSON output (`-j` or `--json`)

Returns structured JSON data with detailed information about each matched element:

```shell
echo '<div><h1 class="title">Hello World</h1><p id="intro">This is a paragraph.</p></div>' | hq 'h1, p' -j
```

```json

{
  "selector": "h1, p",
  "results": [
    {
      "tag": "h1",
      "text": "Hello World",
      "html": "<h1 class=\"title\">Hello World</h1>",
      "attributes": {
        "class": "title"
      }
    },
    {
      "tag": "p",
      "text": "This is a paragraph.",
      "html": "<p id=\"intro\">This is a paragraph.</p>",
      "attributes": {
        "id": "intro"
      }
    }
  ],
  "count": 2
}
```

##### JSON cheat-sheet

Top-level keys
- **selector :string** – the CSS selector you passed on the command line
- **results :array\<Node\>** – matched elements in document order
- **count :number** – total number of matches

`Node` object
- **tag :string** – HTML tag name (lower-case)
- **text :string?** – plain-text content (absent if the element has no text)
- **html :string?** – full outer HTML (present when you add `-k/--keep-tags`)
- **attributes :object<string,string>?** – attributes map; omitted when empty
- **children :array\<Node\>?** – nested elements; every child is itself a `Node` with the same structure (supports arbitrary depth)


### Watching live output

Use the 8/16‑color formatter to stay compatible with *util‑linux* `watch`:

```bash
watch -c "curl -s https://example.com | hq '.price' -C -f terminal16"
```

### Piping to jq

Since **hq** can output JSON with the `-j` flag, you can pipe the results to `jq` for further processing:

```bash
# Extract all links and get their href attributes
curl -s https://example.com | hq 'a' -j | jq '.results[].attributes.href'

# Count the number of paragraphs in a page
curl -s https://example.com | hq 'p' -j | jq '.count'

# Extract all image sources
curl -s https://example.com | hq 'img' -j | jq -r '.results[].attributes.src'

# Find all headings and their text content
curl -s https://example.com | hq 'h1, h2, h3' -j | jq '.results[] | {tag: .tag, text: .text}'
```

### More examples

Extract and format a table:

```bash
# Extract a table and keep the HTML tags for proper formatting
curl -s https://example.com | hq 'table.data' -k

# Extract just the text from table cells
curl -s https://example.com | hq 'table.data td'
```

Extract specific elements with complex selectors:

```bash
# Find all paragraphs inside articles with a specific class
hq 'article.blog p' blog.html

# Extract elements with specific attributes
curl -s https://example.com | hq 'div[data-type="product"]'

# Find elements using pseudo-selectors
curl -s https://example.com | hq 'ul.menu li:first-child'
```

---

## Shell completion

Generate a completion script and source or install it:

```bash
# bash
hq completion bash > /etc/bash_completion.d/hq

# zsh
hq completion zsh > ${ZDOTDIR:-~}/.zfunc/_hq

# oh-my-zsh
hq completion zsh > ~/.oh-my-zsh/completions/_hq
```

Then reload your shell. Tab‑completion now suggests flags *and* enum values, e.g.:

```bash
hq -f <TAB><TAB>
terminal   terminal16   terminal256   terminal16m   html   svg
```

---

## Color styles and formatters

*Styles* are Chroma themes – see the full gallery at [https://xyproto.github.io/splash/docs/all.html](https://xyproto.github.io/splash/docs/all.html).

*Formatters* decide **how many colors** will be emitted:

| Name          | Colours                                  | When to use                                       |
|---------------|------------------------------------------|---------------------------------------------------|
| `autodetect`  | picks best of the below based on `$TERM` |                                                   |
| `terminal16m` | 24‑bit                                   | modern terminals (Kitty, iTerm, Windows Terminal) |
| `terminal256` | 256                                      | tmux, older ncurses apps                          |
| `terminal16`  | 8/16                                     | inside `watch -c`, TTYs without 256‑color support |

### Export formatters

For saving or embedding syntax‑highlighted output in other documents:

| Name   | Output format | Use case                                         |
|--------|---------------|--------------------------------------------------|
| `html` | HTML          | Embed colourised code in web pages or documents  |
| `svg`  | SVG           | Vector graphics for presentations or print media |

Example usage:

```shell
# Extract part of the HTML into a new HTML file
curl -s https://example.com | hq 'pre code' -k -f html > highlighted.html

# Extract part of the HTML into a new HTML file with the styles needed for the colorized output
curl -s https://example.com | hq 'pre code' -k -f html -C > highlighted.html

# Create an SVG with colourised output for presentations
echo '<div class="highlight"><code>console.log("Hello");</code></div>' \
  | hq 'code' -k -f svg -C > code-snippet.svg
```

## License
Released under the MIT License – see [LICENSE](./LICENSE) for full text.
