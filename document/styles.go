package document

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/speedata/boxesandglue/backend/bag"
	"github.com/speedata/boxesandglue/csshtml"
	"github.com/speedata/boxesandglue/frontend"
	"golang.org/x/net/html"
)

func (d *Document) parseFontSize(fs string, dflt bag.ScaledPoint) bag.ScaledPoint {
	if strings.HasSuffix(fs, "em") {
		prefix := strings.TrimSuffix(fs, "em")
		factor, err := strconv.ParseFloat(prefix, 32)
		if err != nil {
			bag.Logger.Errorf("Cannot convert font size %s", fs)
			return bag.MustSp("10pt")
		}
		return bag.ScaledPoint(float64(dflt) * factor)
	}
	return bag.MustSp(fs)
}

func (d *Document) collectHorizontalNodes(sel *goquery.Selection) *frontend.Text {
	currentStyle := d.currentStyle()

	te := &frontend.Text{
		Settings: frontend.TypesettingSettings{
			frontend.SettingFontFamily: currentStyle.fontfamily,
			frontend.SettingLeading:    currentStyle.lineheight,
		},
	}
	n := sel.Nodes[0]
	attr, _ := csshtml.ResolveAttributes(n.Attr)
	var fontWeight string
	for k, v := range attr {
		switch k {
		case "font-style":
			switch v {
			case "italic":
				te.Settings[frontend.SettingStyle] = frontend.FontStyleItalic
			}
		case "font-weight":
			fontWeight = v
		case "font-size":
			te.Settings[frontend.SettingSize] = d.parseFontSize(v, d.currentStyle().fontsize)
		case "color":
			te.Settings[frontend.SettingColor] = v
		case "margin-top":
			te.Settings[frontend.SettingMarginTop] = d.parseFontSize(v, d.currentStyle().fontsize)
		case "margin-bottom":
			te.Settings[frontend.SettingMarginBottom] = d.parseFontSize(v, d.currentStyle().fontsize)
		case "margin-left":
			te.Settings[frontend.SettingMarginLeft] = d.parseFontSize(v, d.currentStyle().fontsize)
		case "margin-right":
			te.Settings[frontend.SettingMarginRight] = d.parseFontSize(v, d.currentStyle().fontsize)
		case "font-family":
			te.Settings[frontend.SettingFontFamily] = d.doc.FindFontFamily(v)
		case "line-height":
			te.Settings[frontend.SettingSize] = d.parseFontSize(v, d.currentStyle().lineheight)
		default:
			fmt.Println("unresolved attribute", k)
		}
	}

	if fontWeight != "" {
		te.Settings[frontend.SettingFontWeight] = frontend.ResolveFontWeight(fontWeight, currentStyle.fontweight)
	}
	sel.Contents().Each(func(i int, contents *goquery.Selection) {
		n := contents.Nodes[0]
		switch n.Type {
		case html.TextNode:
			te.Items = append(te.Items, n.Data)
		case html.ElementNode:
			itm := d.collectHorizontalNodes(contents)
			te.Items = append(te.Items, itm)
		}
	})
	return te
}

func (d *Document) processSelection(i int, sel *goquery.Selection) {
	n := sel.Nodes[0]

	// n.Type can be one of ErrorNode, TextNode, DocumentNode, ElementNode, CommentNode and DoctypeNode
	switch n.Type {
	case html.TextNode:
		// fmt.Println("textnode", n.Data)
	case html.ElementNode:
		a, _ := csshtml.ResolveAttributes(n.Attr)
		styles := d.pushStyles()
		// n.Data is the element name
		switch n.Data {
		case "h1", "p":
			te := d.collectHorizontalNodes(sel)
			d.te = append(d.te, te)
		default:
			for k, v := range a {
				switch k {
				case "display":
				case "font-size":
					styles.fontsize = d.parseFontSize(v, d.currentStyle().fontsize)
				case "font-weight":
					styles.fontweight = frontend.ResolveFontWeight(v, frontend.FontWeight400)
				case "font-family":
					styles.fontfamily = d.doc.FindFontFamily(v)
				case "margin-top", "margin-bottom", "margin-left", "margin-right":
					// ignore
				case "line-height":
					styles.lineheight = d.parseFontSize(v, d.currentStyle().lineheight)
				case "hyphens":
				case "font-style":
				default:
					fmt.Println("unhandled attribute", k)
				}
			}

		}

		sel.Contents().Each(d.processSelection)
		d.popStyles()

	default:
		fmt.Println("other node", n.Type, n.Data)
	}

}

func (d *Document) parseSelection(sel *goquery.Selection) {
	sel.Each(d.processSelection)
}

type inheritStyles struct {
	fontsize   bag.ScaledPoint
	lineheight bag.ScaledPoint
	fontweight frontend.FontWeight
	fontfamily *frontend.FontFamily
	language   string
}

func (is *inheritStyles) clone() *inheritStyles {
	newis := &inheritStyles{
		fontsize:   is.fontsize,
		lineheight: is.lineheight,
		fontfamily: is.fontfamily,
		fontweight: is.fontweight,
		language:   is.language,
	}
	return newis
}

// pushStyles creates a new style instance, pushes it onto the stack and returns
// the new style.
func (d *Document) pushStyles() *inheritStyles {
	var is *inheritStyles
	if len(d.stylesStack) == 0 {
		is = &inheritStyles{}
	} else {
		is = d.stylesStack[len(d.stylesStack)-1].clone()
	}
	d.stylesStack = append(d.stylesStack, is)
	return is
}

// popStyles removes the top style from the stack.
func (d *Document) popStyles() {
	d.stylesStack = d.stylesStack[:len(d.stylesStack)-1]
}

// currentStyle returns the current style from the stack.
func (d *Document) currentStyle() *inheritStyles {
	return d.stylesStack[len(d.stylesStack)-1]
}
