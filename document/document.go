package document

import (
	"os"
	"path/filepath"

	"github.com/speedata/boxesandglue/backend/bag"
	"github.com/speedata/boxesandglue/backend/node"
	"github.com/speedata/boxesandglue/csshtml"
	"github.com/speedata/boxesandglue/frontend"
	"github.com/speedata/boxesandglue/frontend/pdfdraw"
)

// Document is the main starting point of the PDF generation.
type Document struct {
	Title                 string
	Author                string
	Keywords              string // separated by comma
	Creator               string
	Subject               string
	defaultFontsize       bag.ScaledPoint
	defaultFontFamily     *frontend.FontFamily
	currentPageDimensions PageDimensions
	doc                   *frontend.Document
	c                     *csshtml.CSS
	stylesStack           []*formattingStyles
	te                    []*frontend.Text
}

// PageDimensions contains the page size and the margins of the page.
type PageDimensions struct {
	Width         bag.ScaledPoint
	Height        bag.ScaledPoint
	MarginLeft    bag.ScaledPoint
	MarginRight   bag.ScaledPoint
	MarginTop     bag.ScaledPoint
	MarginBottom  bag.ScaledPoint
	PageAreaLeft  bag.ScaledPoint
	PageAreaTop   bag.ScaledPoint
	ContentWidth  bag.ScaledPoint
	ContentHeight bag.ScaledPoint
	masterpage    *csshtml.Page
}

var onecm = bag.MustSp("1cm")

func (d *Document) getPageType() *csshtml.Page {
	if first, ok := d.c.Pages[":first"]; ok && len(d.doc.Doc.Pages) == 0 {
		return &first
	}
	isRight := len(d.doc.Doc.Pages)%2 == 0
	if right, ok := d.c.Pages[":right"]; ok && isRight {
		return &right
	}
	if left, ok := d.c.Pages[":left"]; ok && !isRight {
		return &left
	}
	if allPages, ok := d.c.Pages[""]; ok {
		return &allPages
	}
	return nil
}

func (d *Document) initPage() error {
	var err error
	if d.doc.Doc.CurrentPage == nil {
		if defaultPage := d.getPageType(); defaultPage != nil {
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

			res, _ := csshtml.ResolveAttributes(defaultPage.Attributes)
			styles := d.pushStyles()
			if err = d.stylesToStyles(styles, res); err != nil {
				return err
			}
			vl := node.NewVList()
			vl.Width = wd - ml - mr - styles.borderLeftWidth - styles.borderRightWidth - styles.paddingLeft - styles.paddingRight
			vl.Height = ht - mt - mb - styles.paddingTop - styles.paddingBottom - styles.borderTopWidth - styles.borderBottomWidth
			hv := frontend.HTMLValues{
				BorderLeftWidth:         styles.borderLeftWidth,
				BorderRightWidth:        styles.borderRightWidth,
				BorderTopWidth:          styles.borderTopWidth,
				BorderBottomWidth:       styles.borderBottomWidth,
				BorderTopStyle:          styles.borderTopStyle,
				BorderLeftStyle:         styles.borderLeftStyle,
				BorderRightStyle:        styles.borderRightStyle,
				BorderBottomStyle:       styles.borderBottomStyle,
				BorderTopColor:          styles.borderTopColor,
				BorderLeftColor:         styles.borderLeftColor,
				BorderRightColor:        styles.borderRightColor,
				BorderBottomColor:       styles.borderBottomColor,
				PaddingLeft:             styles.paddingLeft,
				PaddingRight:            styles.paddingRight,
				PaddingBottom:           styles.paddingBottom,
				PaddingTop:              styles.paddingTop,
				BorderTopLeftRadius:     styles.borderTopLeftRadius,
				BorderTopRightRadius:    styles.borderTopRightRadius,
				BorderBottomLeftRadius:  styles.borderBottomLeftRadius,
				BorderBottomRightRadius: styles.borderBottomRightRadius,
			}
			vl = d.doc.HTMLBorder(vl, hv)
			d.popStyles()

			// set page width / height
			d.doc.Doc.DefaultPageWidth = wd
			d.doc.Doc.DefaultPageHeight = ht
			d.currentPageDimensions = PageDimensions{
				Width:         wd,
				Height:        ht,
				PageAreaLeft:  ml + styles.borderLeftWidth + styles.paddingLeft,
				PageAreaTop:   mt - styles.borderTopWidth - styles.paddingTop,
				ContentWidth:  wd - styles.borderRightWidth - styles.paddingRight - ml - mr - styles.borderLeftWidth - styles.paddingLeft,
				ContentHeight: ht - styles.borderBottomWidth - styles.paddingBottom - mt - mb - styles.borderTopWidth - styles.paddingTop,
				MarginTop:     mt,
				MarginBottom:  mb,
				MarginLeft:    ml,
				MarginRight:   mr,
				masterpage:    defaultPage,
			}
			d.doc.Doc.NewPage()
			if styles.backgroundColor != nil {
				r := node.NewRule()
				x := pdfdraw.NewStandalone().ColorNonstroking(*styles.backgroundColor).Rect(0, 0, wd, -ht).Fill()
				r.Pre = x.String()
				rvl := node.Vpack(r)
				rvl.Attributes = node.H{"origin": "page background color"}
				d.doc.Doc.CurrentPage.OutputAt(0, ht, rvl)
			}
			d.doc.Doc.CurrentPage.OutputAt(ml, ht-mt, vl)
		} else {
			// no page master found
			d.doc.Doc.DefaultPageWidth = bag.MustSp("210mm")
			d.doc.Doc.DefaultPageHeight = bag.MustSp("297mm")

			d.currentPageDimensions = PageDimensions{
				Width:         d.doc.Doc.DefaultPageWidth,
				Height:        d.doc.Doc.DefaultPageHeight,
				ContentWidth:  d.doc.Doc.DefaultPageWidth,
				ContentHeight: d.doc.Doc.DefaultPageHeight,
				PageAreaLeft:  onecm,
				PageAreaTop:   onecm,
				MarginTop:     onecm,
				MarginBottom:  onecm,
				MarginLeft:    onecm,
				MarginRight:   onecm,
			}
			d.doc.Doc.NewPage()
		}

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

// ParseCSSFile parses the CSS file at the filename.
func (d *Document) ParseCSSFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		return err
	}
	d.c.Dirstack = []string{abs}
	return d.c.AddCSSText(string(data))
}

// ParseCSSString reads CSS instructions from a string.
func (d *Document) ParseCSSString(css string) error {
	var err error
	if err = d.c.AddCSSText(css); err != nil {
		return err
	}
	return nil
}

// NewPage puts the current page into the PDF document and starts with a new page.
func (d *Document) NewPage() error {
	if err := d.initPage(); err != nil {
		return err
	}
	if err := d.beforeShipout(); err != nil {
		return err
	}
	d.doc.Doc.CurrentPage.Shipout()
	d.doc.Doc.NewPage()
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
	var te *frontend.Text
	if te, err = d.parseSelection(sel, width); err != nil {
		return err
	}
	n, err := d.buildVlist(te, width)
	if err != nil {
		bag.Logger.Error(err)
		return err
	}
	d.doc.Doc.CurrentPage.OutputAt(x, y, n)
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
	d.c.FrontendDocument.Doc.CompressLevel = 9
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
	if err := d.beforeShipout(); err != nil {
		return err
	}
	D.CurrentPage.Shipout()
	return D.Finish()
}
