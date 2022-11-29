package document

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"

	"github.com/speedata/boxesandglue/backend/bag"
	"github.com/speedata/boxesandglue/backend/color"
	"github.com/speedata/boxesandglue/backend/document"
	"github.com/speedata/boxesandglue/backend/node"
	"github.com/speedata/boxesandglue/frontend"
)

var tenpt = bag.MustSp("10pt")
var tenptflt = bag.MustSp("10pt").ToPT()

func parseVerticalAlign(align string, styles *formattingStyles) frontend.VerticalAlignment {
	switch align {
	case "top":
		return frontend.VAlignTop
	case "middle":
		return frontend.VAlignMiddle
	case "bottom":
		return frontend.VAlignBottom
	case "inherit":
		return styles.valign
	default:
		return frontend.VAlignDefault
	}
}

func parseHorizontalAlign(align string, styles *formattingStyles) frontend.HorizontalAlignment {
	switch align {
	case "left":
		return frontend.HAlignLeft
	case "center":
		return frontend.HAlignCenter
	case "right":
		return frontend.HAlignRight
	case "inherit":
		return styles.halign
	default:
		return frontend.HAlignDefault
	}
}

func parseWidth(width string, dflt bag.ScaledPoint) (bag.ScaledPoint, error) {
	if !strings.HasSuffix(width, "%") {
		return 0, fmt.Errorf("no %% suffix")
	}
	wd, err := strconv.Atoi(strings.TrimSuffix(width, "%"))
	if err != nil {
		return 0, err
	}

	return bag.ScaledPoint(int(dflt) * wd / 100), nil
}

// parseRelativeSize converts the string fs to a scaled point. This can be an
// absolute size like 12pt but also a size like 1.2 or 2em. The provided dflt is
// the source size.
func parseRelativeSize(fs string, dflt bag.ScaledPoint) bag.ScaledPoint {
	if strings.HasSuffix(fs, "em") {
		prefix := strings.TrimSuffix(fs, "em")
		factor, err := strconv.ParseFloat(prefix, 32)
		if err != nil {
			bag.Logger.Errorf("Cannot convert relative size %s", fs)
			return bag.MustSp("10pt")
		}
		return bag.ScaledPoint(float64(dflt) * factor)
	}
	if unit, err := bag.Sp(fs); err == nil {
		return unit
	}
	if factor, err := strconv.ParseFloat(fs, 64); err == nil {
		return bag.ScaledPointFromFloat(dflt.ToPT() * factor)
	}
	switch fs {
	case "larger":
		return bag.ScaledPointFromFloat(dflt.ToPT() * 1.2)
	case "smaller":
		return bag.ScaledPointFromFloat(dflt.ToPT() / 1.2)
	case "xx-small":
		return bag.ScaledPointFromFloat(tenptflt / 1.2 / 1.2 / 1.2)
	case "x-small":
		return bag.ScaledPointFromFloat(tenptflt / 1.2 / 1.2)
	case "small":
		return bag.ScaledPointFromFloat(tenptflt / 1.2)
	case "medium":
		return tenpt
	case "large":
		return bag.ScaledPointFromFloat(tenptflt * 1.2)
	case "x-large":
		return bag.ScaledPointFromFloat(tenptflt * 1.2 * 1.2)
	case "xx-large":
		return bag.ScaledPointFromFloat(tenptflt * 1.2 * 1.2 * 1.2)
	case "xxx-large":
		return bag.ScaledPointFromFloat(tenptflt * 1.2 * 1.2 * 1.2 * 1.2)
	}
	bag.Logger.Errorf("Could not convert %s from default %s", fs, dflt)
	return dflt
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
	settings[frontend.SettingOpenTypeFeature] = ih.fontfeatures
}

func (d *Document) stylesToStyles(ih *formattingStyles, attributes map[string]string) {
	// Resolve font size first, since some of the attributes depend on the
	// current font size.
	if v, ok := attributes["font-size"]; ok {
		ih.fontsize = parseRelativeSize(v, d.currentStyle().fontsize)
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
		case "font-feature-settings":
			ih.fontfeatures = append(ih.fontfeatures, v)
		case "color":
			ih.color = d.c.FrontendDocument.GetColor(v)
		case "margin-top":
			ih.marginTop = parseRelativeSize(v, d.currentStyle().fontsize)
		case "margin-bottom":
			ih.marginBottom = parseRelativeSize(v, d.currentStyle().fontsize)
		case "font-family":
			ih.fontfamily = d.doc.FindFontFamily(v)
		case "line-height":
			ih.lineheight = parseRelativeSize(v, d.currentStyle().fontsize)
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
		childSettings := make(frontend.TypesettingSettings)
		switch item.data {
		case "a":
			var href string
			for k, v := range item.attributes {
				switch k {
				case "href":
					href = v
				}
			}
			hl := document.Hyperlink{URI: href}
			childSettings[frontend.SettingHyperlink] = hl
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

		for _, itm := range item.children {
			cld := frontend.NewText()
			sty := d.pushStyles()
			d.stylesToStyles(sty, item.styles)
			d.applySettings(cld.Settings, sty)
			for k, v := range childSettings {
				cld.Settings[k] = v
			}
			if err := d.collectHorizontalNodes(cld, itm); err != nil {
				return err
			}
			te.Items = append(te.Items, cld)
			d.popStyles()
		}
	}
	return nil
}

func (d *Document) output(item *htmlItem, currentWidth bag.ScaledPoint) error {
	// always vertical items
	styles := d.pushStyles()
	d.stylesToStyles(styles, item.styles)
	if item.data == "table" {
		d.processTable(item, currentWidth)
		d.popStyles()
		return nil
	}

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
			d.output(itm, currentWidth)
		}
	}
	if te != nil {
		d.te = append(d.te, te)
	}
	d.popStyles()
	return nil
}

func (d *Document) parseSelection(sel *goquery.Selection, wd bag.ScaledPoint) error {
	h := &htmlItem{dir: modeVertical}
	dumpElement(sel.Nodes[0], modeVertical, h)
	return d.output(h, wd)
}

type formattingStyles struct {
	fontsize     bag.ScaledPoint
	fontstyle    frontend.FontStyle
	lineheight   bag.ScaledPoint
	fontweight   frontend.FontWeight
	fontfeatures []string
	fontfamily   *frontend.FontFamily
	marginTop    bag.ScaledPoint
	marginBottom bag.ScaledPoint
	color        *color.Color
	valign       frontend.VerticalAlignment
	halign       frontend.HorizontalAlignment
	language     string
}

func (is *formattingStyles) clone() *formattingStyles {
	// inherit
	newFeatures := make([]string, len(is.fontfeatures))
	for i, f := range is.fontfeatures {
		newFeatures[i] = f
	}
	newis := &formattingStyles{
		fontsize:     is.fontsize,
		lineheight:   is.lineheight,
		fontfamily:   is.fontfamily,
		fontweight:   is.fontweight,
		fontfeatures: newFeatures,
		language:     is.language,
		fontstyle:    is.fontstyle,
		color:        is.color,
		valign:       is.valign,
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
