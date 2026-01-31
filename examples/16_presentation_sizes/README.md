# Presentation Sizes Example

This example demonstrates how to create presentation-style PDFs using gopdf's presentation page sizes.

## Features

- **16:9 Widescreen Format**: Modern presentation format (720 x 405 points)
- **4:3 Standard Format**: Classic presentation format (720 x 540 points)
- Multiple slides per document
- Visual aspect ratio demonstrations

## Page Sizes

### 16:9 Widescreen (PageSizePresentation16x9)
- **Dimensions**: 720 x 405 points (10" x 5.625")
- **Aspect Ratio**: 16:9 (1.777...)
- **Use Case**: Modern displays, widescreen projectors, video presentations

### 4:3 Standard (PageSizePresentation4x3)
- **Dimensions**: 720 x 540 points (10" x 7.5")
- **Aspect Ratio**: 4:3 (1.333...)
- **Use Case**: Classic projectors, traditional presentations, printed slides

## Usage

```go
import "github.com/ryomak/gopdf"

// Create a 16:9 widescreen presentation
doc := gopdf.New()
page := doc.NewPage(gopdf.PageSizePresentation16x9, gopdf.Portrait)

// Add content to the slide
page.SetFont(gopdf.FontHelvetica, 48)
page.DrawText("Slide Title", 50, 300)

// Save the PDF
f, _ := os.Create("presentation.pdf")
doc.WriteTo(f)
f.Close()
```

## Running the Example

```bash
cd examples/16_presentation_sizes
go run main.go
```

This will create two PDF files:
- `16x9_presentation.pdf` - A 3-slide widescreen presentation
- `4x3_presentation.pdf` - A 4-slide standard presentation

## Output

The example creates presentations with:
1. Title slide with presentation format information
2. Feature list slide
3. Visual aspect ratio demonstration
4. Format comparison (4:3 only)

## Comparison with Other Formats

| Format | Width | Height | Aspect Ratio | Use Case |
|--------|-------|--------|--------------|----------|
| A4 | 595 | 842 | 1:1.41 | Documents |
| Letter | 612 | 792 | 1:1.29 | Documents |
| 16:9 | 720 | 405 | 16:9 | Modern presentations |
| 4:3 | 720 | 540 | 4:3 | Classic presentations |

## Tips

- Use **16:9** for modern displays and video presentations
- Use **4:3** for compatibility with older projectors
- Consider using larger font sizes (24-48pt) for presentations
- Leave adequate margins (50-100 points) from edges
- Use high contrast colors for better visibility

## Related Examples

- See `examples/17_markdown_conversion` for creating presentations from Markdown
- See `examples/02_hello_world` for basic text drawing
- See `examples/03_graphics` for drawing shapes and graphics
