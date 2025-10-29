# Markdown Conversion Example

This example demonstrates how to convert Markdown documents to PDF using gopdf.

## Features

- Convert Markdown files to PDF
- Convert Markdown strings to PDF
- Custom styling support
- CommonMark and GitHub Flavored Markdown support
- Multiple page support with automatic page breaks

## Supported Markdown Elements (Phase 1)

### Currently Supported
- ✅ Headings (H1-H6)
- ✅ Paragraphs
- ✅ Automatic page breaks

### Coming Soon
- ⏳ Lists (bullet and numbered)
- ⏳ Code blocks
- ⏳ Tables
- ⏳ Images
- ⏳ Links
- ⏳ Bold and italic text
- ⏳ Blockquotes

## Usage

### Basic Conversion

```go
import "github.com/ryomak/gopdf"

// Convert a Markdown file
doc, err := gopdf.NewMarkdownDocumentFromFile("document.md", nil)
if err != nil {
    log.Fatal(err)
}

f, _ := os.Create("output.pdf")
doc.WriteTo(f)
f.Close()
```

### Convert from String

```go
markdown := `# Hello World

This is a simple Markdown document.
`

doc, err := gopdf.NewMarkdownDocument(markdown, &gopdf.MarkdownOptions{
    Mode:        gopdf.MarkdownModeDocument,
    PageSize:    gopdf.PageSizeA4,
    Orientation: gopdf.Portrait,
})
```

### Custom Styling

```go
style := gopdf.DefaultMarkdownStyle()
style.H1Size = 42
style.H2Size = 32
style.BodySize = 14
style.HeadingColor = gopdf.Color{R: 0.1, G: 0.3, B: 0.7}

doc, err := gopdf.NewMarkdownDocument(markdown, &gopdf.MarkdownOptions{
    Mode:  gopdf.MarkdownModeDocument,
    Style: style,
})
```

## Running the Example

```bash
cd examples/17_markdown_conversion
go run main.go
```

This will create three PDF files:
- `output_from_file.pdf` - Converted from sample.md
- `output_from_string.pdf` - Converted from a Markdown string
- `output_custom_style.pdf` - Converted with custom styling

## Options

### MarkdownOptions

```go
type MarkdownOptions struct {
    Mode          MarkdownMode  // "document" or "slide"
    PageSize      PageSize      // Page size (default: A4 for document, 16:9 for slide)
    Orientation   Orientation   // Portrait or Landscape
    Style         *MarkdownStyle // Custom style (optional)
    ImageBasePath string        // Base path for relative images
}
```

### MarkdownStyle

```go
type MarkdownStyle struct {
    H1Size, H2Size, H3Size, H4Size, H5Size, H6Size float64
    BodySize                                        float64
    CodeSize                                        float64
    LineSpacing                                     float64
    ParagraphSpacing                                float64
    MarginTop, MarginRight, MarginBottom, MarginLeft float64
    TextColor, HeadingColor, CodeBackground, LinkColor Color
}
```

## Modes

### Document Mode (Current)
- Multi-page document format
- Automatic page breaks
- Traditional document layout

### Slide Mode (Coming Soon)
- Presentation slides
- One slide per page
- Horizontal rules or H1 as slide delimiters
- 16:9 or 4:3 aspect ratios

## Tips

1. Use headings to structure your document
2. Keep paragraphs concise for better readability
3. Customize styles to match your branding
4. Use A4 or Letter page sizes for documents
5. Use presentation sizes (16:9, 4:3) for slides (coming soon)

## Related Examples

- See `examples/16_presentation_sizes` for presentation page sizes
- See `examples/02_hello_world` for basic text drawing
- See `examples/09_ttf_fonts` for Japanese and Unicode support (coming soon for Markdown)

## Roadmap

### Phase 1 (Current)
- ✅ Basic headings and paragraphs
- ✅ Automatic page breaks
- ✅ Custom styling

### Phase 2
- ⏳ Lists and code blocks
- ⏳ Images and links
- ⏳ Text formatting (bold, italic)

### Phase 3
- ⏳ Tables
- ⏳ Blockquotes
- ⏳ Horizontal rules

### Phase 4
- ⏳ Slide mode
- ⏳ Syntax highlighting
- ⏳ Advanced theming
