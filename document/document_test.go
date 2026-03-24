package document

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/boxesandglue/boxesandglue/backend/document"
)

func tempPDF(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "test.pdf")
}

func TestNew(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename)
	if err != nil {
		t.Fatal(err)
	}
	if d.Frontend == nil {
		t.Fatal("Frontend should not be nil")
	}
	if err := d.RenderPages("<p>test</p>"); err != nil {
		t.Fatal(err)
	}
	if err := d.Finish(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatal("PDF file not created:", err)
	}
	if info.Size() == 0 {
		t.Fatal("PDF file is empty")
	}
}

func TestNewInvalidPath(t *testing.T) {
	_, err := New("/nonexistent/dir/test.pdf")
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

func TestWithPDFUA(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename, WithPDFUA())
	if err != nil {
		t.Fatal(err)
	}
	if err := d.RenderPages("<p>Accessible PDF</p>"); err != nil {
		t.Fatal(err)
	}
	if err := d.Finish(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("PDF file is empty")
	}
}

func TestWithPDFA3b(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename, WithPDFA3b())
	if err != nil {
		t.Fatal(err)
	}
	if err := d.RenderPages("<p>Archive PDF</p>"); err != nil {
		t.Fatal(err)
	}
	if err := d.Finish(); err != nil {
		t.Fatal(err)
	}
}

func TestWithPDFX3(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename, WithPDFX3())
	if err != nil {
		t.Fatal(err)
	}
	if err := d.RenderPages("<p>Print PDF</p>"); err != nil {
		t.Fatal(err)
	}
	if err := d.Finish(); err != nil {
		t.Fatal(err)
	}
}

func TestWithPDFX4(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename, WithPDFX4())
	if err != nil {
		t.Fatal(err)
	}
	if err := d.RenderPages("<p>Print PDF X4</p>"); err != nil {
		t.Fatal(err)
	}
	if err := d.Finish(); err != nil {
		t.Fatal(err)
	}
}

func TestAddCSS(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.AddCSS(`body { font-size: 14pt; color: red; }`); err != nil {
		t.Fatal(err)
	}
	if err := d.RenderPages("<p>Styled text</p>"); err != nil {
		t.Fatal(err)
	}
	if err := d.Finish(); err != nil {
		t.Fatal(err)
	}
}

func TestReadCSSFileNotFound(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename)
	if err != nil {
		t.Fatal(err)
	}
	err = d.ReadCSSFile("/nonexistent/style.css")
	if err == nil {
		t.Fatal("expected error for nonexistent CSS file")
	}
}

func TestPageSizeDefault(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.cssbuilder.InitPage(); err != nil {
		t.Fatal(err)
	}
	ps, err := d.PageSize()
	if err != nil {
		t.Fatal(err)
	}
	// Default is A4: 210mm x 297mm
	a4Width := bag.MustSP("210mm")
	a4Height := bag.MustSP("297mm")
	if ps.Width != a4Width {
		t.Errorf("expected width %d (210mm), got %d", a4Width, ps.Width)
	}
	if ps.Height != a4Height {
		t.Errorf("expected height %d (297mm), got %d", a4Height, ps.Height)
	}
}

func TestPageSizeCustom(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.AddCSS(`@page { size: 100mm 200mm; margin: 10mm; }`); err != nil {
		t.Fatal(err)
	}
	if err := d.cssbuilder.InitPage(); err != nil {
		t.Fatal(err)
	}
	ps, err := d.PageSize()
	if err != nil {
		t.Fatal(err)
	}
	expectedWidth := bag.MustSP("100mm")
	expectedHeight := bag.MustSP("200mm")
	if ps.Width != expectedWidth {
		t.Errorf("expected width %d (100mm), got %d", expectedWidth, ps.Width)
	}
	if ps.Height != expectedHeight {
		t.Errorf("expected height %d (200mm), got %d", expectedHeight, ps.Height)
	}
}

func TestRenderPages(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename)
	if err != nil {
		t.Fatal(err)
	}
	html := `<h1>Title</h1><p>Hello, world!</p><p>Second paragraph.</p>`
	if err := d.RenderPages(html); err != nil {
		t.Fatal(err)
	}
	if err := d.Finish(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("PDF file is empty")
	}
}

func TestOutputAt(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename)
	if err != nil {
		t.Fatal(err)
	}
	width := bag.MustSP("80mm")
	x := bag.MustSP("20mm")
	y := bag.MustSP("250mm")
	if err := d.OutputAt("<p>Positioned text</p>", width, x, y); err != nil {
		t.Fatal(err)
	}
	if err := d.Finish(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("PDF file is empty")
	}
}

func TestHeadings(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename)
	if err != nil {
		t.Fatal(err)
	}
	html := `<h1>First</h1><p>Text</p><h2>Second</h2><p>More text</p><h1>Third</h1>`
	if err := d.RenderPages(html); err != nil {
		t.Fatal(err)
	}
	headings := d.Headings()
	if len(headings) != 3 {
		t.Fatalf("expected 3 headings, got %d", len(headings))
	}
	if headings[0].Text != "First" {
		t.Errorf("expected heading text 'First', got %q", headings[0].Text)
	}
	if headings[1].Text != "Second" {
		t.Errorf("expected heading text 'Second', got %q", headings[1].Text)
	}
	if headings[2].Text != "Third" {
		t.Errorf("expected heading text 'Third', got %q", headings[2].Text)
	}
	if err := d.Finish(); err != nil {
		t.Fatal(err)
	}
}

func TestHeadingsEmpty(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.RenderPages("<p>No headings here</p>"); err != nil {
		t.Fatal(err)
	}
	headings := d.Headings()
	if len(headings) != 0 {
		t.Errorf("expected 0 headings, got %d", len(headings))
	}
	if err := d.Finish(); err != nil {
		t.Fatal(err)
	}
}

func TestMetadata(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename)
	if err != nil {
		t.Fatal(err)
	}
	d.Title = "Test Title"
	d.Author = "Test Author"
	d.Keywords = "test,pdf,go"
	d.Creator = "bagme-test"
	d.Subject = "Testing"
	d.Language = "de"

	if err := d.RenderPages("<p>Metadata test</p>"); err != nil {
		t.Fatal(err)
	}
	if err := d.Finish(); err != nil {
		t.Fatal(err)
	}

	pdfDoc := d.Frontend.Doc
	if pdfDoc.Title != "Test Title" {
		t.Errorf("expected title 'Test Title', got %q", pdfDoc.Title)
	}
	if pdfDoc.Author != "Test Author" {
		t.Errorf("expected author 'Test Author', got %q", pdfDoc.Author)
	}
	if pdfDoc.Keywords != "test,pdf,go" {
		t.Errorf("expected keywords 'test,pdf,go', got %q", pdfDoc.Keywords)
	}
	if pdfDoc.Creator != "bagme-test" {
		t.Errorf("expected creator 'bagme-test', got %q", pdfDoc.Creator)
	}
	if pdfDoc.Subject != "Testing" {
		t.Errorf("expected subject 'Testing', got %q", pdfDoc.Subject)
	}
	if pdfDoc.DefaultLanguageTag != "de" {
		t.Errorf("expected language 'de', got %q", pdfDoc.DefaultLanguageTag)
	}
}

func TestNewPage(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename)
	if err != nil {
		t.Fatal(err)
	}
	width := bag.MustSP("80mm")
	x := bag.MustSP("20mm")
	y := bag.MustSP("250mm")
	if err := d.OutputAt("<p>Page 1</p>", width, x, y); err != nil {
		t.Fatal(err)
	}
	if err := d.NewPage(); err != nil {
		t.Fatal(err)
	}
	if err := d.OutputAt("<p>Page 2</p>", width, x, y); err != nil {
		t.Fatal(err)
	}
	if err := d.Finish(); err != nil {
		t.Fatal(err)
	}
}

func TestAttachFile(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename)
	if err != nil {
		t.Fatal(err)
	}
	d.AttachFile(Attachment{
		Name:        "data.txt",
		Description: "Test attachment",
		MimeType:    "text/plain",
		Data:        []byte("Hello, attachment!"),
	})
	if err := d.RenderPages("<p>With attachment</p>"); err != nil {
		t.Fatal(err)
	}
	if err := d.Finish(); err != nil {
		t.Fatal(err)
	}
}

func TestWithAttachmentOption(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename, WithAttachment(Attachment{
		Name:        "invoice.xml",
		Description: "Invoice data",
		MimeType:    "text/xml",
		Data:        []byte("<invoice/>"),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if err := d.RenderPages("<p>With attachment option</p>"); err != nil {
		t.Fatal(err)
	}
	if err := d.Finish(); err != nil {
		t.Fatal(err)
	}
}

func TestWithZUGFeRD(t *testing.T) {
	filename := tempPDF(t)
	xmlData := []byte(`<?xml version="1.0"?><invoice/>`)
	d, err := New(filename, WithZUGFeRD(xmlData, "EN 16931"))
	if err != nil {
		t.Fatal(err)
	}
	if err := d.RenderPages("<p>ZUGFeRD invoice</p>"); err != nil {
		t.Fatal(err)
	}
	if err := d.Finish(); err != nil {
		t.Fatal(err)
	}
}

func TestWithZUGFeRDProfileNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"EN16931", "EN 16931"},
		{"en16931", "EN 16931"},
		{"COMFORT", "EN 16931"},
		{"comfort", "EN 16931"},
		{"BASICWL", "BASIC WL"},
		{"BASIC WL", "BASIC WL"},
		{"MINIMUM", "MINIMUM"},
		{"BASIC", "BASIC"},
		{"EXTENDED", "EXTENDED"},
		{"XRECHNUNG", "XRECHNUNG"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var cfg config
			opt := WithZUGFeRD([]byte("<xml/>"), tt.input)
			opt(&cfg)
			if len(cfg.xmpExtensions) == 0 {
				t.Fatal("expected XMP extension")
			}
			level := cfg.xmpExtensions[0].Values["ConformanceLevel"]
			if level != tt.expected {
				t.Errorf("profile %q: expected ConformanceLevel %q, got %q", tt.input, tt.expected, level)
			}
		})
	}
}

func TestWithZUGFeRDSetsFormat(t *testing.T) {
	var cfg config
	opt := WithZUGFeRD([]byte("<xml/>"), "BASIC")
	opt(&cfg)
	if cfg.format != document.FormatPDFA3b {
		t.Errorf("WithZUGFeRD should set format to PDF/A-3b, got %d", cfg.format)
	}
	if len(cfg.attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(cfg.attachments))
	}
	if cfg.attachments[0].Name != "factur-x.xml" {
		t.Errorf("expected attachment name 'factur-x.xml', got %q", cfg.attachments[0].Name)
	}
}

func TestPageInitCallback(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename)
	if err != nil {
		t.Fatal(err)
	}
	callCount := 0
	d.PageInitCallback = func() {
		callCount++
	}
	if err := d.RenderPages("<p>Callback test</p>"); err != nil {
		t.Fatal(err)
	}
	if err := d.Finish(); err != nil {
		t.Fatal(err)
	}
	if callCount == 0 {
		t.Error("PageInitCallback was never called")
	}
}

func TestReadCSSFile(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename)
	if err != nil {
		t.Fatal(err)
	}
	cssFile := filepath.Join(t.TempDir(), "test.css")
	if err := os.WriteFile(cssFile, []byte(`body { font-size: 16pt; }`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := d.ReadCSSFile(cssFile); err != nil {
		t.Fatal(err)
	}
	if err := d.RenderPages("<p>CSS from file</p>"); err != nil {
		t.Fatal(err)
	}
	if err := d.Finish(); err != nil {
		t.Fatal(err)
	}
}

func TestRenderPagesWithCSS(t *testing.T) {
	filename := tempPDF(t)
	d, err := New(filename)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.AddCSS(`@page { size: A5; margin: 15mm; }`); err != nil {
		t.Fatal(err)
	}
	html := `<h1>A5 Document</h1><p>This is rendered on A5 paper with 15mm margins.</p>`
	if err := d.RenderPages(html); err != nil {
		t.Fatal(err)
	}
	if err := d.Finish(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("PDF file is empty")
	}
}

func TestMultipleOptionsChaining(t *testing.T) {
	var cfg config
	opts := []Option{
		WithPDFA3b(),
		WithAttachment(Attachment{
			Name:     "a.txt",
			MimeType: "text/plain",
			Data:     []byte("A"),
		}),
		WithAttachment(Attachment{
			Name:     "b.txt",
			MimeType: "text/plain",
			Data:     []byte("B"),
		}),
	}
	for _, o := range opts {
		o(&cfg)
	}
	if len(cfg.attachments) != 2 {
		t.Errorf("expected 2 attachments, got %d", len(cfg.attachments))
	}
}
