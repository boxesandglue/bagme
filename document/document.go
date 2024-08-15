package document

import (
	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/boxesandglue/boxesandglue/frontend"
	"github.com/boxesandglue/csshtml"
	"github.com/boxesandglue/htmlbag"
)

// Document is the main starting point of the PDF generation.
type Document struct {
	Title      string
	Author     string
	Keywords   string // separated by comma
	Creator    string
	Subject    string
	Frontend   *frontend.Document
	cssbuilder *htmlbag.CSSBuilder
}

// PageSize returns a struct with the dimensions of the current page.
func (d *Document) PageSize() (htmlbag.PageDimensions, error) {
	return d.cssbuilder.PageSize()
}

// ReadCSSFile parses the CSS file at the filename.
func (d *Document) ReadCSSFile(filename string) error {
	return d.cssbuilder.ReadCSSFile(filename)
}

// OutputAt writes the HTML string to the PDF.
func (d *Document) OutputAt(html string, width bag.ScaledPoint, x, y bag.ScaledPoint) error {
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
	p := d.Frontend.Doc.CurrentPage
	p.OutputAt(x, y, vl)
	return nil
}

// NewWithFrontend creates a document from a boxes and glue frontend document
// and a csshtml CSS parser. The default fonts for the families monospace, sans
// and serif are loaded.
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

// New creates a PDF file with the provided file name, initializes a new boxes
// and glue frontend document and loads a default CSS stylesheet.
func New(filename string) (*Document, error) {
	var err error
	d := &Document{}
	d.Frontend, err = frontend.New(filename)
	if err != nil {
		return nil, err
	}
	cs := csshtml.NewCSSParserWithDefaults()
	return NewWithFrontend(d.Frontend, cs)
}

// Finish writes and closes the PDF.
func (d *Document) Finish() error {
	pdfDoc := d.Frontend.Doc
	pdfDoc.Title = d.Title
	pdfDoc.Author = d.Author
	pdfDoc.Keywords = d.Keywords
	pdfDoc.Creator = d.Creator
	pdfDoc.Subject = d.Subject
	// if err := d.cssbuilder.BeforeShipout(); err != nil {
	// 	return err
	// }
	pdfDoc.CurrentPage.Shipout()
	return pdfDoc.Finish()
}
