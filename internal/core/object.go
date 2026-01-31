// Package core provides low-level PDF object types and structures.
// This package implements the fundamental building blocks of PDF documents
// as defined in the PDF specification (ISO 32000-1).
package core

// Object is the interface that all PDF objects must implement.
// PDF objects can be of various types: null, boolean, numeric, string, name,
// array, dictionary, stream, or indirect reference.
type Object interface {
	// isObject is a marker method to ensure type safety.
	// Only types in this package can implement Object.
	isObject()
}

// Null represents the PDF null object.
type Null struct{}

func (Null) isObject() {}

// Boolean represents a PDF boolean value (true or false).
type Boolean bool

func (Boolean) isObject() {}

// Integer represents a PDF integer number.
type Integer int

func (Integer) isObject() {}

// Real represents a PDF real (floating-point) number.
type Real float64

func (Real) isObject() {}

// String represents a PDF string.
// In PDF, strings can be literal strings (enclosed in parentheses)
// or hexadecimal strings (enclosed in angle brackets).
type String string

func (String) isObject() {}

// Name represents a PDF name object.
// Names are used as keys in dictionaries and to identify various PDF entities.
// In PDF syntax, names are preceded by a forward slash (/).
type Name string

func (Name) isObject() {}

// Array represents a PDF array, which is an ordered collection of objects.
type Array []Object

func (Array) isObject() {}

// Dictionary represents a PDF dictionary, which is a collection of key-value pairs.
// Keys are always Name objects, and values can be any PDF object type.
type Dictionary map[Name]Object

func (Dictionary) isObject() {}

// Stream represents a PDF stream object, which consists of a dictionary
// followed by binary data. Streams are used for large amounts of data
// such as page content, images, and fonts.
type Stream struct {
	Dict Dictionary
	Data []byte
}

func (*Stream) isObject() {}

// Reference represents an indirect reference to a PDF object.
// Indirect references allow objects to be reused and create graph structures.
type Reference struct {
	ObjectNumber     int
	GenerationNumber int
}

func (*Reference) isObject() {}

// IndirectObject represents an indirect object definition in a PDF file.
// It wraps an object with its object and generation numbers.
type IndirectObject struct {
	ObjectNumber     int
	GenerationNumber int
	Object           Object
}

func (*IndirectObject) isObject() {}
