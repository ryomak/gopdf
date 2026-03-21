package gopdf

import (
	"bytes"
	"testing"
)

func TestTextIterator_EmptyElements(t *testing.T) {
	it := &TextIterator{
		elements: nil,
		index:    0,
	}

	if it.HasNext() {
		t.Error("HasNext() should be false for empty iterator")
	}
	if it.Next() {
		t.Error("Next() should return false for empty iterator")
	}
	if it.Element() != nil {
		t.Error("Element() should be nil when no Next() called")
	}
	if it.Count() != 0 {
		t.Errorf("Count() = %d, want 0", it.Count())
	}
}

func TestTextIterator_BasicIteration(t *testing.T) {
	elements := []TextElement{
		{Text: "Hello", X: 10, Y: 100, Size: 12},
		{Text: "World", X: 50, Y: 100, Size: 12},
		{Text: "Test", X: 10, Y: 80, Size: 14},
	}

	it := &TextIterator{
		elements: elements,
		index:    0,
	}

	if it.Count() != 3 {
		t.Errorf("Count() = %d, want 3", it.Count())
	}

	// Iterate through all elements
	var collected []string
	for it.Next() {
		elem := it.Element()
		if elem == nil {
			t.Fatal("Element() should not be nil after Next() returns true")
		}
		collected = append(collected, elem.Text)
	}

	if len(collected) != 3 {
		t.Fatalf("collected %d elements, want 3", len(collected))
	}

	expected := []string{"Hello", "World", "Test"}
	for i, want := range expected {
		if collected[i] != want {
			t.Errorf("element[%d] = %q, want %q", i, collected[i], want)
		}
	}

	// After iteration, Next should return false
	if it.Next() {
		t.Error("Next() should return false after all elements consumed")
	}
	if it.Element() != nil {
		t.Error("Element() should be nil after iteration complete")
	}
}

func TestTextIterator_Reset(t *testing.T) {
	elements := []TextElement{
		{Text: "A", X: 10, Y: 100, Size: 12},
		{Text: "B", X: 50, Y: 100, Size: 12},
	}

	it := &TextIterator{
		elements: elements,
		index:    0,
	}

	// Consume all
	for it.Next() {
	}

	if it.HasNext() {
		t.Error("HasNext() should be false after consuming all")
	}

	// Reset
	it.Reset()

	if !it.HasNext() {
		t.Error("HasNext() should be true after reset")
	}
	if it.Element() != nil {
		t.Error("Element() should be nil after reset")
	}

	// Should be able to iterate again
	if !it.Next() {
		t.Error("Next() should return true after reset")
	}
	if it.Element().Text != "A" {
		t.Errorf("after reset, first element = %q, want %q", it.Element().Text, "A")
	}
}

func TestTextIterator_HasNext(t *testing.T) {
	elements := []TextElement{
		{Text: "Only", X: 10, Y: 100, Size: 12},
	}

	it := &TextIterator{
		elements: elements,
		index:    0,
	}

	if !it.HasNext() {
		t.Error("HasNext() should be true before consuming")
	}

	it.Next()

	if it.HasNext() {
		t.Error("HasNext() should be false after consuming the only element")
	}
}

func TestAllPagesTextIterator_Basic(t *testing.T) {
	// Create a PDF with text on multiple pages
	doc := New()

	page1 := doc.AddPage(PageSizeA4, Portrait)
	if err := page1.SetFont(FontHelvetica, 12); err != nil {
		t.Fatal(err)
	}
	_ = page1.DrawText("Page1Text", 100, 700)

	page2 := doc.AddPage(PageSizeA4, Portrait)
	if err := page2.SetFont(FontHelvetica, 12); err != nil {
		t.Fatal(err)
	}
	_ = page2.DrawText("Page2Text", 100, 700)

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("OpenReader failed: %v", err)
	}
	defer reader.Close()

	// Test AllPagesTextIterator
	allIter := reader.AllPagesTextIterator()

	if allIter.CurrentPage() != -1 {
		t.Errorf("initial CurrentPage() = %d, want -1", allIter.CurrentPage())
	}

	// Should be able to iterate
	foundTexts := make(map[int][]string)
	for allIter.Next() {
		elem := allIter.Element()
		if elem != nil && elem.Text != "" {
			foundTexts[allIter.CurrentPage()] = append(foundTexts[allIter.CurrentPage()], elem.Text)
		}
	}

	// We should have found text on at least one page
	if len(foundTexts) == 0 {
		t.Error("expected to find text on at least one page")
	}
}

func TestAllPagesTextIterator_Reset(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)
	if err := page.SetFont(FontHelvetica, 12); err != nil {
		t.Fatal(err)
	}
	_ = page.DrawText("Test", 100, 700)

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	allIter := reader.AllPagesTextIterator()

	// Consume all
	count1 := 0
	for allIter.Next() {
		count1++
	}

	// Reset and iterate again
	allIter.Reset()
	count2 := 0
	for allIter.Next() {
		count2++
	}

	if count1 != count2 {
		t.Errorf("after reset, got %d elements, want %d", count2, count1)
	}
}

func TestAllPagesTextIterator_ElementWithoutNext(t *testing.T) {
	doc := New()
	doc.AddPage(PageSizeA4, Portrait)

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	allIter := reader.AllPagesTextIterator()

	// Element without calling Next should return nil
	if allIter.Element() != nil {
		t.Error("Element() should be nil before calling Next()")
	}
}

func TestPDFReader_TextIterator(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)
	if err := page.SetFont(FontHelvetica, 12); err != nil {
		t.Fatal(err)
	}
	_ = page.DrawText("IteratorTest", 100, 700)

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	// Valid page
	iter, err := reader.TextIterator(0)
	if err != nil {
		t.Fatalf("TextIterator(0) error: %v", err)
	}
	if iter == nil {
		t.Fatal("TextIterator should not be nil")
	}

	// Invalid page
	_, err = reader.TextIterator(99)
	if err == nil {
		t.Error("TextIterator(99) should return error for out-of-range page")
	}
}
