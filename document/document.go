package document

import (
	"github.com/speedata/boxesandglue/backend/bag"
	"github.com/speedata/boxesandglue/backend/color"
	"github.com/speedata/boxesandglue/csshtml"
	"github.com/speedata/boxesandglue/frontend"
)

// Document is the main starting point of the PDF generation.
type Document struct {
	Title                 string
	Author                string
	Keywords              string // separated by comma
	Creator               string
	Subject               string
	defaultFontsize       bag.ScaledPoint
	currentPageDimensions PageDimensions
	doc                   *frontend.Document
	c                     *csshtml.CSS
	stylesStack           []*formattingStyles
	te                    []*frontend.Text
}

// PageDimensions contains the page size and the margins of the page.
type PageDimensions struct {
	Width        bag.ScaledPoint
	Height       bag.ScaledPoint
	MarginLeft   bag.ScaledPoint
	MarginRight  bag.ScaledPoint
	MarginTop    bag.ScaledPoint
	MarginBottom bag.ScaledPoint
}

var onecm = bag.MustSp("1cm")

func (d *Document) initPage() error {
	var err error
	if d.doc.Doc.CurrentPage == nil {
		if defaultPage, ok := d.c.Pages[""]; ok {
			wdStr, htStr := csshtml.PapersizeWdHt(defaultPage.Papersize)
			var wd, ht, mt, mb, ml, mr bag.ScaledPoint
			if wd, err = bag.Sp(wdStr); err != nil {
				return err
			}
			if ht, err = bag.Sp(htStr); err != nil {
				return err
			}

			if str := defaultPage.MarginTop; str == "" {
				mt = onecm
			} else {
				if mt, err = bag.Sp(str); err != nil {
					return err
				}
			}

			if str := defaultPage.MarginBottom; str == "" {
				mb = onecm
			} else {
				if mb, err = bag.Sp(str); err != nil {
					return err
				}
			}
			if str := defaultPage.MarginLeft; str == "" {
				ml = onecm
			} else {
				if ml, err = bag.Sp(str); err != nil {
					return err
				}
			}
			if str := defaultPage.MarginRight; str == "" {
				mr = onecm
			} else {
				if mr, err = bag.Sp(str); err != nil {
					return err
				}
			}

			// set page width / height
			d.doc.Doc.DefaultPageWidth = wd
			d.doc.Doc.DefaultPageHeight = ht

			d.currentPageDimensions = PageDimensions{
				Width:        wd,
				Height:       ht,
				MarginTop:    mt,
				MarginBottom: mb,
				MarginLeft:   ml,
				MarginRight:  mr,
			}
		} else {

			d.doc.Doc.DefaultPageWidth = bag.MustSp("210mm")
			d.doc.Doc.DefaultPageHeight = bag.MustSp("297mm")

			d.currentPageDimensions = PageDimensions{
				Width:        d.doc.Doc.DefaultPageWidth,
				Height:       d.doc.Doc.DefaultPageHeight,
				MarginTop:    onecm,
				MarginBottom: onecm,
				MarginLeft:   onecm,
				MarginRight:  onecm,
			}
		}
		d.doc.Doc.NewPage()
	}
	return err
}

// PageSize returns a struct with the dimensions of the current page.
func (d *Document) PageSize() (PageDimensions, error) {
	err := d.initPage()
	if err != nil {
		return PageDimensions{}, err
	}
	return d.currentPageDimensions, nil
}

// ParseCSSString reads CSS instructions from a string.
func (d *Document) ParseCSSString(css string) error {
	var err error
	if err = d.c.AddCSSText(css); err != nil {
		return err
	}
	return nil
}

// OutputAt writes the HTML string to the PDF.
func (d *Document) OutputAt(html string, width bag.ScaledPoint, x, y bag.ScaledPoint) error {
	if err := d.initPage(); err != nil {
		return err
	}
	if err := d.c.ReadHTMLChunk(html); err != nil {
		return err
	}
	sel, err := d.c.ApplyCSS()
	if err != nil {
		return err
	}
	if err = d.parseSelection(sel, width); err != nil {
		return err
	}
	for i, te := range d.te {
		lineWidth := width
		var opts []frontend.TypesettingOption
		if indent, ok := te.Settings[frontend.SettingIndentLeft]; ok {
			if rows, ok := te.Settings[frontend.SettingIndentLeftRows]; ok {
				opts = append(opts, frontend.IndentLeft(indent.(bag.ScaledPoint), rows.(int)))
			} else {
				opts = append(opts, frontend.IndentLeft(indent.(bag.ScaledPoint), 1))
			}
		}
		hv := frontend.HTMLValues{}
		if bw, ok := te.Settings[frontend.SettingBorderTopWidth]; ok {
			hv.BorderTopWidth = bw.(bag.ScaledPoint)
		}
		if bw, ok := te.Settings[frontend.SettingBorderBottomWidth]; ok {
			hv.BorderBottomWidth = bw.(bag.ScaledPoint)
		}
		if bw, ok := te.Settings[frontend.SettingBorderLeftWidth]; ok {
			hv.BorderLeftWidth = bw.(bag.ScaledPoint)
		}
		if bw, ok := te.Settings[frontend.SettingBorderRightWidth]; ok {
			hv.BorderRightWidth = bw.(bag.ScaledPoint)
		}
		if bw, ok := te.Settings[frontend.SettingBorderTopLeftRadius]; ok {
			hv.BorderTopLeftRadius = bw.(bag.ScaledPoint)
		}
		if wd, ok := te.Settings[frontend.SettingMarginTop]; ok {
			hv.MarginTop = wd.(bag.ScaledPoint)
		}
		if wd, ok := te.Settings[frontend.SettingMarginBottom]; ok {
			hv.MarginBottom = wd.(bag.ScaledPoint)
		}
		if wd, ok := te.Settings[frontend.SettingMarginLeft]; ok {
			hv.MarginLeft = wd.(bag.ScaledPoint)
		}
		if wd, ok := te.Settings[frontend.SettingMarginRight]; ok {
			hv.MarginRight = wd.(bag.ScaledPoint)
		}
		if wd, ok := te.Settings[frontend.SettingPaddingTop]; ok {
			hv.PaddingTop = wd.(bag.ScaledPoint)
			delete(te.Settings, frontend.SettingPaddingTop)
		}
		if wd, ok := te.Settings[frontend.SettingPaddingBottom]; ok {
			hv.PaddingBottom = wd.(bag.ScaledPoint)
			delete(te.Settings, frontend.SettingPaddingBottom)
		}
		if wd, ok := te.Settings[frontend.SettingPaddingLeft]; ok {
			hv.PaddingLeft = wd.(bag.ScaledPoint)
			delete(te.Settings, frontend.SettingPaddingLeft)
		}
		if wd, ok := te.Settings[frontend.SettingPaddingRight]; ok {
			hv.PaddingRight = wd.(bag.ScaledPoint)
			delete(te.Settings, frontend.SettingPaddingRight)
		}
		if bw, ok := te.Settings[frontend.SettingBorderTopLeftRadius]; ok {
			hv.BorderTopLeftRadius = bw.(bag.ScaledPoint)
		}
		if bw, ok := te.Settings[frontend.SettingBorderTopRightRadius]; ok {
			hv.BorderTopRightRadius = bw.(bag.ScaledPoint)
		}
		if bw, ok := te.Settings[frontend.SettingBorderBottomLeftRadius]; ok {
			hv.BorderBottomLeftRadius = bw.(bag.ScaledPoint)
		}
		if bw, ok := te.Settings[frontend.SettingBorderBottomRightRadius]; ok {
			hv.BorderBottomRightRadius = bw.(bag.ScaledPoint)
		}
		if col, ok := te.Settings[frontend.SettingBorderRightColor]; ok {
			hv.BorderRightColor = col.(*color.Color)
		}
		if col, ok := te.Settings[frontend.SettingBorderLeftColor]; ok {
			hv.BorderLeftColor = col.(*color.Color)
		}
		if col, ok := te.Settings[frontend.SettingBorderTopColor]; ok {
			hv.BorderTopColor = col.(*color.Color)
		}
		if col, ok := te.Settings[frontend.SettingBorderBottomColor]; ok {
			hv.BorderBottomColor = col.(*color.Color)
		}
		if sty, ok := te.Settings[frontend.SettingBorderRightStyle]; ok {
			hv.BorderRightStyle = sty.(frontend.BorderStyle)
		}
		if sty, ok := te.Settings[frontend.SettingBorderLeftStyle]; ok {
			hv.BorderLeftStyle = sty.(frontend.BorderStyle)
		}
		if sty, ok := te.Settings[frontend.SettingBorderTopStyle]; ok {
			hv.BorderTopStyle = sty.(frontend.BorderStyle)
		}
		if sty, ok := te.Settings[frontend.SettingBorderBottomStyle]; ok {
			hv.BorderBottomStyle = sty.(frontend.BorderStyle)
		}

		lineWidth = lineWidth - hv.MarginLeft - hv.MarginRight - hv.PaddingLeft - hv.PaddingRight - hv.BorderLeftWidth - hv.BorderRightWidth
		vl, _, err := d.doc.FormatParagraph(te, lineWidth, opts...)
		if err != nil {
			return err
		}
		vl = d.doc.HTMLBorder(vl, hv)

		d.doc.Doc.CurrentPage.OutputAt(x+hv.MarginLeft, y, vl)
		y -= vl.Height + vl.Depth
		var additionalMargin bag.ScaledPoint
		additionalMargin = hv.MarginBottom
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
	if d.doc.Doc.DefaultLanguage, err = frontend.GetLanguage("en"); err != nil {
		return nil, err
	}
	d.c = csshtml.NewCSSParser()
	d.c.Stylesheet = append(d.c.Stylesheet, csshtml.ConsumeBlock(csshtml.ParseCSSString(cssdefaults), false))

	d.c.FrontendDocument = d.doc
	return d, nil
}

// Finish writes and closes the PDF.
func (d *Document) Finish() error {
	D := d.doc.Doc
	D.Title = d.Title
	D.Author = d.Author
	D.Keywords = d.Keywords
	D.Creator = d.Creator
	D.Subject = d.Subject
	D.CurrentPage.Shipout()
	return D.Finish()
}
