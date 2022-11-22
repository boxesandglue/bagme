package document

import (
	"github.com/speedata/boxesandglue/backend/bag"
	"github.com/speedata/boxesandglue/csshtml"
	"github.com/speedata/boxesandglue/frontend"
)

// Document is the main starting point of the PDF generation.
type Document struct {
	doc         *frontend.Document
	c           *csshtml.CSS
	stylesStack []*formattingStyles
	te          []*frontend.Text
}

// ParseCSSString reads CSS instructions from a string.
func (d *Document) ParseCSSString(css string) error {
	var err error
	d.c.AddCSSText(css)
	if defaultPage, ok := d.c.Pages[""]; ok {
		wdStr, htStr := csshtml.PapersizeWdHt(defaultPage.Papersize)
		var wd, ht bag.ScaledPoint
		wd, err = bag.Sp(wdStr)
		if err != nil {
			return err
		}
		ht, err = bag.Sp(htStr)
		if err != nil {
			return err
		}
		// set page width / height
		d.doc.Doc.DefaultPageWidth = wd
		d.doc.Doc.DefaultPageHeight = ht
	}
	return nil
}

// OutputAt writes the HTML string to the PDF.
func (d *Document) OutputAt(html string, width bag.ScaledPoint, x, y bag.ScaledPoint) error {
	err := d.c.ReadHTMLChunk(html)
	if err != nil {
		return err
	}
	sel, err := d.c.ApplyCSS()
	if err != nil {
		return err
	}
	if err = d.parseSelection(sel); err != nil {
		return err
	}
	for i, te := range d.te {
		vl, _, err := d.doc.FormatParagraph(te, width)
		if err != nil {
			return err
		}
		d.doc.Doc.CurrentPage.OutputAt(x, y, vl)
		y -= vl.Height + vl.Depth
		var additionalMargin bag.ScaledPoint
		if mb, ok := te.Settings[frontend.SettingMarginBottom]; ok {
			additionalMargin = mb.(bag.ScaledPoint)
		}
		if i+1 < len(d.te) {
			if mt, ok := d.te[i+1].Settings[frontend.SettingMarginTop]; ok {
				marginTop := mt.(bag.ScaledPoint)
				if marginTop > additionalMargin {
					additionalMargin = marginTop
				}

			}
		}
		y -= additionalMargin
	}

	d.te = nil
	return nil
}

// New writes the PDF
func New(filename string) (*Document, error) {
	var err error
	d := &Document{}
	d.doc, err = frontend.New(filename)
	if err != nil {
		return nil, err
	}
	d.doc.Doc.CompressLevel = 0
	if d.doc.Doc.DefaultLanguage, err = frontend.GetLanguage("en"); err != nil {
		return nil, err
	}
	d.c = csshtml.NewCSSParserWithDefaults()
	d.c.FrontendDocument = d.doc
	d.doc.Doc.NewPage()
	return d, nil
}

// Finish writes and closes the PDF.
func (d *Document) Finish() error {
	d.doc.Doc.CurrentPage.Shipout()
	return d.doc.Doc.Finish()
}
