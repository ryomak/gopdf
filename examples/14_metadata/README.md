# PDF Metadata Example

This example demonstrates how to add metadata to a PDF document using the gopdf library.

## Overview

PDF metadata (Document Information Dictionary) allows you to embed information about the document such as:
- Title
- Author
- Subject
- Keywords
- Creator
- Producer
- Creation Date
- Modification Date
- Custom fields

## Running the Example

```bash
go run main.go
```

This will create `metadata_example.pdf` with embedded metadata.

## Features Demonstrated

### Writing Metadata

```go
metadata := gopdf.Metadata{
    Title:        "PDF Metadata Example",
    Author:       "gopdf Library",
    Subject:      "Demonstration of PDF Metadata",
    Keywords:     "PDF, metadata, gopdf, example",
    Creator:      "gopdf Example Program",
    Producer:     "gopdf v1.0",
    CreationDate: time.Date(2025, 1, 29, 12, 0, 0, 0, time.UTC),
    ModDate:      time.Now(),
}
doc.SetMetadata(metadata)
```

### Reading Metadata

```go
reader, err := gopdf.Open("metadata_example.pdf")
if err != nil {
    panic(err)
}
defer reader.Close()

metadata := reader.Info()

fmt.Printf("Title: %s\n", metadata.Title)
fmt.Printf("Author: %s\n", metadata.Author)
fmt.Printf("Subject: %s\n", metadata.Subject)
fmt.Printf("Keywords: %s\n", metadata.Keywords)
fmt.Printf("Creator: %s\n", metadata.Creator)
fmt.Printf("Producer: %s\n", metadata.Producer)
fmt.Printf("CreationDate: %s\n", metadata.CreationDate.Format(time.RFC3339))
fmt.Printf("ModDate: %s\n", metadata.ModDate.Format(time.RFC3339))
```

### Custom Metadata Fields

You can also add custom metadata fields:

```go
metadata := gopdf.Metadata{
    Title:  "My Document",
    Author: "John Doe",
    Custom: map[string]string{
        "Department": "Engineering",
        "Project":    "gopdf",
        "Version":    "1.0.0",
    },
}
doc.SetMetadata(metadata)
```

### Default Values

If you don't set certain fields, gopdf will use default values:
- `Producer`: "gopdf" (if not specified)
- `CreationDate`: current time (if not specified)

### Special Characters

The library automatically handles special characters and non-ASCII text (Japanese, Chinese, etc.):

```go
metadata := gopdf.Metadata{
    Title:   "日本語タイトル",
    Author:  "田中太郎",
    Subject: "テスト文書",
}
doc.SetMetadata(metadata)
```

## Viewing Metadata

You can view the metadata in various PDF viewers:

### macOS Preview
1. Open the PDF in Preview
2. Go to Tools > Show Inspector (Cmd+I)
3. Click the "Info" tab

### Adobe Acrobat Reader
1. Open the PDF in Acrobat Reader
2. Go to File > Properties (Ctrl/Cmd+D)
3. View the "Description" tab

### Command Line (exiftool)
```bash
exiftool metadata_example.pdf
```

### Command Line (pdfinfo)
```bash
pdfinfo metadata_example.pdf
```

## Reference

See the [metadata specification documentation](../../docs/metadata-specification.md) for detailed information about PDF metadata structure and encoding.
