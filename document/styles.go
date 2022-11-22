package document

import (
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/speedata/boxesandglue/backend/bag"
	"github.com/speedata/boxesandglue/backend/color"
	"github.com/speedata/boxesandglue/backend/node"
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

func (d *Document) applySettings(settings frontend.TypesettingSettings, ih *formattingStyles) {
	if ih.fontweight > 0 {
		settings[frontend.SettingFontWeight] = ih.fontweight
	}
	settings[frontend.SettingSize] = ih.fontsize
	settings[frontend.SettingFontFamily] = ih.fontfamily
	settings[frontend.SettingStyle] = ih.fontstyle
	settings[frontend.SettingLeading] = ih.lineheight
	settings[frontend.SettingMarginTop] = ih.marginTop
	settings[frontend.SettingMarginBottom] = ih.marginBottom
	settings[frontend.SettingColor] = ih.color
}

func (d *Document) stylesToStyles(ih *formattingStyles, attributes map[string]string) {
	// Resolve font size first, since some of the attributes depend on the
	// current font size.
	if v, ok := attributes["font-size"]; ok {
		ih.fontsize = d.parseFontSize(v, d.currentStyle().fontsize)
	}

	for k, v := range attributes {
		switch k {
		case "font-size":
			// already set
		case "display", "hyphens", "margin-left", "margin-right":
			// ignore for now
		case "font-style":
			switch v {
			case "italic":
				ih.fontstyle = frontend.FontStyleItalic
			case "normal":
				ih.fontstyle = frontend.FontStyleNormal
			}
		case "font-weight":
			ih.fontweight = frontend.ResolveFontWeight(v, ih.fontweight)
		case "color":
			ih.color = d.c.FrontendDocument.GetColor(v)
		case "margin-top":
			ih.marginTop = d.parseFontSize(v, d.currentStyle().fontsize)
		case "margin-bottom":
			ih.marginBottom = d.parseFontSize(v, d.currentStyle().fontsize)
		case "font-family":
			ih.fontfamily = d.doc.FindFontFamily(v)
		case "line-height":
			ih.lineheight = d.parseFontSize(v, d.currentStyle().lineheight)
		default:
			// fmt.Println("unresolved attribute", k)
		}
	}
}

func (d *Document) collectHorizontalNodes(te *frontend.Text, item *htmlItem) error {
	switch item.typ {
	case html.TextNode:
		te.Items = append(te.Items, item.data)
	case html.ElementNode:
		switch item.data {
		case "img":
			wd := bag.MustSp("3cm")
			ht := wd
			var filename string
			for k, v := range item.attributes {
				switch k {
				case "width":
					wd = bag.MustSp(v)
				case "height":
					ht = bag.MustSp(v)
				case "src":
					filename = v
				}
			}
			imgfile, err := d.doc.Doc.LoadImageFile(filename)
			if err != nil {
				panic(err)
			}

			ii := d.doc.Doc.CreateImage(imgfile, 1)
			imgNode := node.NewImage()
			imgNode.Img = ii
			imgNode.Width = wd
			imgNode.Height = ht
			hlist := node.Hpack(imgNode)

			te.Items = append(te.Items, hlist)
		}
		cld := frontend.NewText()
		sty := d.pushStyles()
		d.stylesToStyles(sty, item.styles)
		d.applySettings(cld.Settings, sty)
		for _, itm := range item.children {
			if err := d.collectHorizontalNodes(cld, itm); err != nil {
				return err
			}
			te.Items = append(te.Items, cld)
		}
		d.popStyles()
	}
	return nil
}

func (d *Document) output(item *htmlItem) error {
	// always vertical items
	styles := d.pushStyles()
	d.stylesToStyles(styles, item.styles)

	var te *frontend.Text
	cur := modeVertical
	for _, itm := range item.children {
		if itm.dir == modeHorizontal {
			if cur == modeVertical && itm.data == " " {
				// Going from vertical to horizontal. No there is only a
				// whitespace element.
				continue
			}
			// now in horizontal mode, there can be more children in horizontal
			// mode, so append all of them to a single frontend.Text element
			if te == nil {
				te = frontend.NewText()
			}
			d.applySettings(te.Settings, styles)
			if err := d.collectHorizontalNodes(te, itm); err != nil {
				return err
			}
			cur = modeHorizontal
		} else {
			if te != nil {
				d.te = append(d.te, te)
				te = nil
			}
			d.output(itm)
		}
	}
	if te != nil {
		d.te = append(d.te, te)
	}
	d.popStyles()
	return nil
}

func (d *Document) parseSelection(sel *goquery.Selection) error {
	h := &htmlItem{dir: modeVertical}
	dumpElement(sel.Nodes[0], modeVertical, h)
	return d.output(h)
}

type formattingStyles struct {
	fontsize     bag.ScaledPoint
	fontstyle    frontend.FontStyle
	lineheight   bag.ScaledPoint
	fontweight   frontend.FontWeight
	fontfamily   *frontend.FontFamily
	marginTop    bag.ScaledPoint
	marginBottom bag.ScaledPoint
	color        *color.Color
	language     string
}

func (is *formattingStyles) clone() *formattingStyles {
	// inherit
	newis := &formattingStyles{
		fontsize:   is.fontsize,
		lineheight: is.lineheight,
		fontfamily: is.fontfamily,
		fontweight: is.fontweight,
		language:   is.language,
		fontstyle:  is.fontstyle,
		color:      is.color,
	}
	return newis
}

// pushStyles creates a new style instance, pushes it onto the stack and returns
// the new style.
func (d *Document) pushStyles() *formattingStyles {
	var is *formattingStyles
	if len(d.stylesStack) == 0 {
		is = &formattingStyles{}
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
func (d *Document) currentStyle() *formattingStyles {
	return d.stylesStack[len(d.stylesStack)-1]
}
