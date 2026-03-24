package document

import (
	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/boxesandglue/boxesandglue/backend/document"
	"github.com/boxesandglue/boxesandglue/frontend"
	"github.com/boxesandglue/csshtml"
	"github.com/boxesandglue/htmlbag"
)

// Attachment represents a file to embed in the PDF. Re-exported from the
// backend for convenience.
type Attachment = document.Attachment

// Option configures document creation. Use the With* functions to create options.
type Option func(*config)

type config struct {
	format        document.Format
	attachments   []document.Attachment
	xmpExtensions []document.XMPExtension
}

// WithPDFUA enables PDF/UA (accessible PDF) output. All HTML elements are
// automatically tagged with their corresponding PDF structure roles (e.g.
// h1→H1, p→P, li→LI) so that screen readers and assistive technologies can
// interpret the content.
func WithPDFUA() Option {
	return func(c *config) { c.format = document.FormatPDFUA }
}

// WithPDFA3b enables PDF/A-3b output.
func WithPDFA3b() Option {
	return func(c *config) { c.format = document.FormatPDFA3b }
}

// WithPDFX3 enables PDF/X-3 output.
func WithPDFX3() Option {
	return func(c *config) { c.format = document.FormatPDFX3 }
}

// WithPDFX4 enables PDF/X-4 output.
func WithPDFX4() Option {
	return func(c *config) { c.format = document.FormatPDFX4 }
}

// WithAttachment embeds a file in the PDF document. For PDF/A-3b documents
// the file appears as an associated file. Use this together with WithPDFA3b()
// for standards like ZUGFeRD/Factur-X, or use WithZUGFeRD for convenience.
func WithAttachment(a Attachment) Option {
	return func(c *config) { c.attachments = append(c.attachments, a) }
}

// WithZUGFeRD creates a ZUGFeRD/Factur-X compliant PDF. It sets the format to
// PDF/A-3b, attaches the XML invoice data as "factur-x.xml", and adds the
// required XMP extension schema metadata.
//
// The profile parameter specifies the Factur-X conformance level:
// "MINIMUM", "BASIC WL", "BASIC", "EN 16931" (or "EN16931"), "EXTENDED", "XRECHNUNG".
func WithZUGFeRD(xmlData []byte, profile string) Option {
	// Normalize common profile shortcuts to official XMP conformance level values.
	switch profile {
	case "EN16931", "en16931", "COMFORT", "comfort":
		profile = "EN 16931"
	case "BASICWL", "BASIC WL":
		profile = "BASIC WL"
	}
	return func(c *config) {
		c.format = document.FormatPDFA3b
		c.attachments = append(c.attachments, document.Attachment{
			Name:        "factur-x.xml",
			Description: "Factur-X/ZUGFeRD invoice",
			MimeType:    "text/xml",
			Data:        xmlData,
		})
		c.xmpExtensions = append(c.xmpExtensions, document.XMPExtension{
			Schema:       "ZUGFeRD PDFA Extension Schema",
			NamespaceURI: "urn:ferd:pdfa:CrossIndustryDocument:invoice:1p0#",
			Prefix:       "zf",
			Properties: []document.XMPExtensionProperty{
				{Name: "DocumentFileName", ValueType: "Text", Category: "external", Description: "name of the embedded XML invoice file"},
				{Name: "DocumentType", ValueType: "Text", Category: "external", Description: "INVOICE"},
				{Name: "Version", ValueType: "Text", Category: "external", Description: "The actual version of the ZUGFeRD data"},
				{Name: "ConformanceLevel", ValueType: "Text", Category: "external", Description: "The conformance level of the ZUGFeRD data"},
			},
			Values: map[string]string{
				"ConformanceLevel": profile,
				"DocumentFileName": "factur-x.xml",
				"DocumentType":     "INVOICE",
				"Version":          "1.0",
			},
		})
	}
}

// Document is the main starting point of the PDF generation.
type Document struct {
	Title    string
	Author   string
	Keywords string // separated by comma
	Creator  string
	Subject  string
	Language string // BCP 47 language tag (e.g. "en", "de")
	Frontend *frontend.Document

	// PageInitCallback is called after each new page is initialized.
	// Use this for running headers/footers or page-level decorations.
	PageInitCallback func()

	// ElementCallback is called after each block element is rendered.
	// Use this for post-processing headings, paragraphs, etc.
	ElementCallback htmlbag.ElementCallbackFunc

	cssbuilder    *htmlbag.CSSBuilder
	pagesRendered bool // true after RenderPages has been called
}

// PageDimensions re-exports the htmlbag type for convenience.
type PageDimensions = htmlbag.PageDimensions

// HeadingEntry re-exports the htmlbag type for convenience.
type HeadingEntry = htmlbag.HeadingEntry

// PageSize returns the dimensions of the current page (width, height, margins).
func (d *Document) PageSize() (PageDimensions, error) {
	return d.cssbuilder.PageSize()
}

// ReadCSSFile parses the CSS file at the given path.
func (d *Document) ReadCSSFile(filename string) error {
	return d.cssbuilder.ReadCSSFile(filename)
}

// AddCSS reads CSS instructions from a string.
func (d *Document) AddCSS(css string) error {
	return d.cssbuilder.AddCSS(css)
}

// OutputAt renders an HTML fragment at an absolute position (x, y) on the
// current page with the given width. Use this for precise placement of
// content snippets (labels, letterheads, positioned boxes).
//
// For full-page rendering with automatic page breaks, use RenderPages instead.
func (d *Document) OutputAt(html string, width, x, y bag.ScaledPoint) error {
	if err := d.cssbuilder.InitPage(); err != nil {
		return err
	}
	te, err := d.cssbuilder.HTMLToText(html)
	if err != nil {
		return err
	}
	vl, err := d.cssbuilder.CreateVlist(te, width)
	if err != nil {
		return err
	}
	d.Frontend.Doc.CurrentPage.OutputAt(x, y, vl)
	return nil
}

// RenderPages renders a complete HTML document with automatic page breaks.
// Page size and margins are taken from CSS @page rules (default: A4 with 1cm
// margins). Content is distributed across pages automatically. Forced page
// breaks (page-break-before, page-break-after) are respected.
//
// After calling RenderPages, call Finish to write the PDF.
// Do not mix RenderPages and OutputAt in the same document.
func (d *Document) RenderPages(html string) error {
	d.syncCallbacks()
	if err := d.cssbuilder.InitPage(); err != nil {
		return err
	}
	te, err := d.cssbuilder.HTMLToText(html)
	if err != nil {
		return err
	}
	d.pagesRendered = true
	return d.cssbuilder.OutputPagesFromText(te)
}

// Headings returns all headings (h1–h6) found during rendering, with their
// page numbers. Only available after RenderPages has been called.
func (d *Document) Headings() []HeadingEntry {
	return d.cssbuilder.Headings
}

// NewPage starts a new page. Only needed in OutputAt mode for manual
// multi-page documents.
func (d *Document) NewPage() error {
	return d.cssbuilder.NewPage()
}

// AttachFile embeds a file in the PDF document.
func (d *Document) AttachFile(a Attachment) {
	d.Frontend.Doc.AttachFile(a)
}

// NewWithFrontend creates a document from an existing boxes and glue frontend
// document and CSS parser. The default fonts (monospace, sans, serif) are loaded.
func NewWithFrontend(fe *frontend.Document, cssparser *csshtml.CSS) (*Document, error) {
	d := &Document{}
	var err error
	d.cssbuilder, err = htmlbag.New(fe, cssparser)
	if err != nil {
		return nil, err
	}
	d.Frontend = fe
	return d, nil
}

// New creates a PDF file at the given path. It initializes a boxes and glue
// frontend document and loads a default CSS stylesheet. Options can be passed
// to select a specific PDF format (e.g. WithPDFUA(), WithPDFA3b()).
func New(filename string, opts ...Option) (*Document, error) {
	var cfg config
	for _, o := range opts {
		o(&cfg)
	}
	fe, err := frontend.New(filename)
	if err != nil {
		return nil, err
	}
	fe.Doc.Format = cfg.format
	// HTML/CSS uses RGB colors, so load sRGB profile for PDF/A-3b
	// instead of the default CMYK profile.
	if cfg.format == document.FormatPDFA3b {
		fe.Doc.LoadSRGBColorprofile()
	}
	for _, a := range cfg.attachments {
		fe.Doc.AttachFile(a)
	}
	for _, ext := range cfg.xmpExtensions {
		fe.Doc.AddXMPExtension(ext)
	}
	cs := csshtml.NewCSSParserWithDefaults()
	return NewWithFrontend(fe, cs)
}

// Finish writes and closes the PDF file.
func (d *Document) Finish() error {
	pdfDoc := d.Frontend.Doc
	pdfDoc.Title = d.Title
	pdfDoc.Author = d.Author
	pdfDoc.Keywords = d.Keywords
	pdfDoc.Creator = d.Creator
	pdfDoc.Subject = d.Subject
	pdfDoc.DefaultLanguageTag = d.Language
	if !d.pagesRendered {
		// Snippet mode: the caller placed content manually, so we need to
		// ship out the current page.
		pdfDoc.CurrentPage.Shipout()
	}
	return pdfDoc.Finish()
}

// syncCallbacks propagates Document-level callbacks to the CSSBuilder.
func (d *Document) syncCallbacks() {
	if d.PageInitCallback != nil {
		d.cssbuilder.PageInitCallback = d.PageInitCallback
	}
	if d.ElementCallback != nil {
		d.cssbuilder.ElementCallback = d.ElementCallback
	}
}
