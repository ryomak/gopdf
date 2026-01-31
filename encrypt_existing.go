package gopdf

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/reader"
	"github.com/ryomak/gopdf/internal/security"
	"github.com/ryomak/gopdf/internal/writer"
)

// EncryptExistingPDF encrypts an existing PDF file without reconstructing it.
// This preserves the original fonts, layout, and all content exactly.
func EncryptExistingPDF(inputPath, outputPath string, opts EncryptionOptions) error {
	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Open output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	return EncryptExistingPDFReader(inputFile, outputFile, opts)
}

// EncryptExistingPDFReader encrypts an existing PDF from a reader to a writer.
func EncryptExistingPDFReader(input io.ReadSeeker, output io.Writer, opts EncryptionOptions) error {
	// Validate options
	if err := opts.Validate(); err != nil {
		return err
	}

	// Parse the input PDF
	r, err := reader.NewReader(input)
	if err != nil {
		return fmt.Errorf("failed to parse input PDF: %w", err)
	}

	// Check if already encrypted
	if r.IsEncrypted() {
		return fmt.Errorf("input PDF is already encrypted")
	}

	// Setup encryption
	encryptInfo, err := writer.SetupEncryption(
		opts.UserPassword,
		opts.OwnerPassword,
		opts.Permissions.toInternal(),
		opts.KeyLength,
	)
	if err != nil {
		return fmt.Errorf("failed to setup encryption: %w", err)
	}

	// Create writer
	w := writer.NewWriter(output)
	w.SetEncryption(encryptInfo)

	// Write PDF header
	if err := w.WriteHeader(); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Get all object numbers and sort them
	objNums := r.GetAllObjectNumbers()
	sort.Ints(objNums)

	// Map old object numbers to new ones (they will be the same in this case)
	// But we need to track them for the trailer
	objMapping := make(map[int]int)

	// Copy all objects with encryption
	for _, oldObjNum := range objNums {
		obj, gen, err := r.GetRawObjectWithGeneration(oldObjNum)
		if err != nil {
			// Skip objects that can't be read
			continue
		}

		// Encrypt strings in the object
		encryptedObj := encryptObject(obj, oldObjNum, gen, encryptInfo)

		// Add the object
		newObjNum, err := w.AddObject(encryptedObj)
		if err != nil {
			return fmt.Errorf("failed to add object %d: %w", oldObjNum, err)
		}

		objMapping[oldObjNum] = newObjNum
	}

	// Build trailer
	trailer := buildTrailer(r.GetTrailer(), objMapping)

	// Write trailer (encryption dict is added by writer)
	if err := w.WriteTrailer(trailer); err != nil {
		return fmt.Errorf("failed to write trailer: %w", err)
	}

	return nil
}

// encryptObject encrypts strings in an object (streams are handled by writer)
func encryptObject(obj core.Object, objNum, genNum int, encryptInfo *writer.EncryptionInfo) core.Object {
	keyLengthBytes := encryptInfo.KeyLength / 8

	switch v := obj.(type) {
	case core.String:
		// Encrypt string
		encrypted := security.EncryptString(
			string(v),
			encryptInfo.EncryptionKey,
			objNum, genNum,
			keyLengthBytes,
		)
		return core.String(encrypted)

	case core.Dictionary:
		// Recursively encrypt dictionary values
		newDict := make(core.Dictionary)
		for k, val := range v {
			newDict[k] = encryptObject(val, objNum, genNum, encryptInfo)
		}
		return newDict

	case core.Array:
		// Recursively encrypt array elements
		newArray := make(core.Array, len(v))
		for i, val := range v {
			newArray[i] = encryptObject(val, objNum, genNum, encryptInfo)
		}
		return newArray

	case *core.Stream:
		// Encrypt dictionary values (stream data is encrypted by writer.AddObject)
		newDict := make(core.Dictionary)
		for k, val := range v.Dict {
			// Don't encrypt Length, Filter, DecodeParms
			if k == core.Name("Length") || k == core.Name("Filter") || k == core.Name("DecodeParms") {
				newDict[k] = val
			} else {
				newDict[k] = encryptObject(val, objNum, genNum, encryptInfo)
			}
		}
		return &core.Stream{
			Dict: newDict,
			Data: v.Data,
		}

	default:
		return obj
	}
}

// buildTrailer builds a new trailer dictionary with updated references
func buildTrailer(oldTrailer core.Dictionary, objMapping map[int]int) core.Dictionary {
	trailer := make(core.Dictionary)

	// Copy and update references
	for k, v := range oldTrailer {
		// Skip Encrypt and ID - they will be added by writer
		if k == core.Name("Encrypt") || k == core.Name("ID") {
			continue
		}

		// Skip Prev - we're writing a new PDF without incremental updates
		if k == core.Name("Prev") {
			continue
		}

		// Update references
		trailer[k] = updateReferences(v, objMapping)
	}

	// Update Size to reflect the new object count
	maxObjNum := 0
	for _, newNum := range objMapping {
		if newNum > maxObjNum {
			maxObjNum = newNum
		}
	}
	trailer[core.Name("Size")] = core.Integer(maxObjNum + 1)

	return trailer
}

// updateReferences updates object references according to the mapping
func updateReferences(obj core.Object, objMapping map[int]int) core.Object {
	switch v := obj.(type) {
	case *core.Reference:
		if newNum, ok := objMapping[v.ObjectNumber]; ok {
			return &core.Reference{
				ObjectNumber:     newNum,
				GenerationNumber: v.GenerationNumber,
			}
		}
		return v

	case core.Dictionary:
		newDict := make(core.Dictionary)
		for k, val := range v {
			newDict[k] = updateReferences(val, objMapping)
		}
		return newDict

	case core.Array:
		newArray := make(core.Array, len(v))
		for i, val := range v {
			newArray[i] = updateReferences(val, objMapping)
		}
		return newArray

	default:
		return obj
	}
}
