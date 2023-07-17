package document

import (
	"github.com/speedata/boxesandglue/backend/bag"
	"github.com/speedata/boxesandglue/backend/node"
	"github.com/speedata/boxesandglue/csshtml"
	"github.com/speedata/boxesandglue/frontend"
	"github.com/speedata/boxesandglue/frontend/cssbuilder"
	"golang.org/x/net/html"
)

// Document is the main starting point of the PDF generation.
type Document struct {
	Title      string
	Author     string
	Keywords   string // separated by comma
	Creator    string
	Subject    string
	Frontend   *frontend.Document
	cssbuilder *cssbuilder.CSSBuilder
}

// PageSize returns a struct with the dimensions of the current page.
func (d *Document) PageSize() (cssbuilder.PageDimensions, error) {
	return d.cssbuilder.PageSize()
}

// ParseCSSFile parses the CSS file at the filename.
func (d *Document) ParseCSSFile(filename string) error {
	return d.cssbuilder.ParseCSSFile(filename)
}

// AddCSS permanently adds the css instructions to the current state.
func (d *Document) AddCSS(css string) {
	d.cssbuilder.AddCSS(css)
}

// ParseHTML interprets the HTML string and applies all previously read CSS data.
func (d *Document) ParseHTML(html string) (*frontend.Text, error) {
	return d.cssbuilder.ParseHTML(html)
}

// ParseHTMLFromNode interprets the HTML structure and applies all previously read CSS data.
func (d *Document) ParseHTMLFromNode(html *html.Node) (*frontend.Text, error) {
	return d.cssbuilder.ParseHTMLFromNode(html)
}

// OutputAt writes the HTML string to the PDF.
func (d *Document) OutputAt(html string, width bag.ScaledPoint, x, y bag.ScaledPoint) error {
	if err := d.cssbuilder.InitPage(); err != nil {
		return err
	}

	te, err := d.cssbuilder.ParseHTML(html)
	if err != nil {
		return err
	}

	err = d.cssbuilder.OutputAt(te, x, y, width)
	if err != nil {
		return err
	}

	return nil
}

// ShowCSS dumps the currently known CSS to a CSS like string
func (d *Document) ShowCSS() string {
	return d.cssbuilder.ShowCSS()
}

// NewWithFrontend creates a document with a boxes and glue frontend document.
func NewWithFrontend(fe *frontend.Document, cssparser *csshtml.CSS) *Document {
	d := &Document{}
	d.Frontend = fe
	d.cssbuilder = cssbuilder.New(fe, cssparser)
	return d
}

// CreateVlist returns a single vertical list ready to be placed in the PDF.
func (d *Document) CreateVlist(te *frontend.Text, wd bag.ScaledPoint) (*node.VList, error) {
	return d.cssbuilder.CreateVlist(te, wd)
}

// New writes the PDF
func New(filename string) (*Document, error) {
	var err error
	d := &Document{}
	d.Frontend, err = frontend.New(filename)
	if err != nil {
		return nil, err
	}
	d.cssbuilder = cssbuilder.New(d.Frontend, csshtml.NewCSSParser())

	if d.Frontend.Doc.DefaultLanguage, err = frontend.GetLanguage("en"); err != nil {
		return nil, err
	}
	d.Frontend.Doc.CompressLevel = 9
	return d, nil
}

// Finish writes and closes the PDF.
func (d *Document) Finish() error {
	pdfDoc := d.Frontend.Doc
	pdfDoc.Title = d.Title
	pdfDoc.Author = d.Author
	pdfDoc.Keywords = d.Keywords
	pdfDoc.Creator = d.Creator
	pdfDoc.Subject = d.Subject
	if err := d.cssbuilder.BeforeShipout(); err != nil {
		return err
	}
	pdfDoc.CurrentPage.Shipout()
	return pdfDoc.Finish()
}
