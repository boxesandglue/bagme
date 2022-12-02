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
// the source size. The root is the document's default value.
func parseRelativeSize(fs string, dflt bag.ScaledPoint, root bag.ScaledPoint) bag.ScaledPoint {
	// TODO: apply percentage such as 110%
	if strings.HasSuffix(fs, "rem") {
		prefix := strings.TrimSuffix(fs, "rem")
		factor, err := strconv.ParseFloat(prefix, 32)
		if err != nil {
			bag.Logger.Errorf("Cannot convert relative size %s", fs)
			return bag.MustSp("10pt")
		}
		return bag.ScaledPoint(float64(root) * factor)
	}
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
	settings[frontend.SettingIndentLeft] = ih.indent
	settings[frontend.SettingIndentLeftRows] = ih.indentRows
	settings[frontend.SettingLeading] = ih.lineheight
	settings[frontend.SettingMarginTop] = ih.marginTop
	settings[frontend.SettingMarginBottom] = ih.marginBottom
	settings[frontend.SettingColor] = ih.color
	settings[frontend.SettingOpenTypeFeature] = ih.fontfeatures
	settings[frontend.SettingPreserveWhitespace] = ih.preserveWhitespace
	settings[frontend.SettingYOffset] = ih.yoffset
	settings[frontend.SettingHAlign] = ih.halign

}

func (d *Document) stylesToStyles(ih *formattingStyles, attributes map[string]string) {
	// Resolve font size first, since some of the attributes depend on the
	// current font size.
	if v, ok := attributes["font-size"]; ok {
		ih.fontsize = parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
	}
	for k, v := range attributes {
		switch k {
		case "font-size":
			// already set
		case "display", "hyphens", "margin-left", "margin-right":
			// ignore for now
		case "color":
			ih.color = d.c.FrontendDocument.GetColor(v)
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
		case "list-style-type":
			ih.listStyleType = v
		case "font-family":
			ih.fontfamily = d.doc.FindFontFamily(v)
		case "line-height":
			ih.lineheight = parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
		case "margin-bottom":
			ih.marginBottom = parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
		case "margin-top":
			ih.marginTop = parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
		case "padding-inline-start":
			ih.paddingInlineStart = parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
		case "text-align":
			ih.halign = parseHorizontalAlign(v, ih)
		case "text-indent":
			ih.indent = parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
			ih.indentRows = 1
		case "vertical-align":
			if v == "sub" {
				ih.yoffset = -1 * ih.fontsize * 1000 / 5000
			} else if v == "super" {
				ih.yoffset = ih.fontsize * 1000 / 5000
			}
		case "white-space":
			ih.preserveWhitespace = (v == "pre")
		default:
			fmt.Println("unresolved attribute", k, v)
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
	// item is guaranteed to be in vertical direction
	styles := d.pushStyles()
	d.stylesToStyles(styles, item.styles)
	var prepend []any
	switch item.data {
	case "html":
		if fs, ok := item.styles["font-size"]; ok {
			rfs := parseRelativeSize(fs, 0, 0)
			d.defaultFontsize = rfs
		}
	case "table":
		d.processTable(item, currentWidth)
		d.popStyles()
		return nil
	case "ol", "ul":
		styles.olCounter = 0
		vspace := frontend.NewText()
		vspace.Settings[frontend.SettingMarginTop] = styles.marginTop
		d.te = append(d.te, vspace)
	case "li":
		settings := make(frontend.TypesettingSettings)
		d.applySettings(settings, styles)
		item := styles.listStyleType
		switch styles.listStyleType {
		case "disc":
			item = "•"
		case "circle":
			item = "◦"
		case "square":
			item = "□"
		case "decimal":
			item = fmt.Sprintf("%d.", styles.olCounter)
		}
		item += " "
		n, err := d.doc.BuildNodelistFromString(settings, item)
		if err != nil {
			return err
		}

		g := node.NewGlue()
		g.Stretch = 1 * bag.Factor
		g.StretchOrder = node.StretchFil
		n = node.InsertBefore(n, n, g)
		n = node.HpackTo(n, styles.paddingInlineStart)
		prepend = append(prepend, n)
		styles.indent = styles.paddingInlineStart
		styles.indentRows = -1
	}

	var te *frontend.Text
	cur := modeVertical
	for _, itm := range item.children {
		if itm.dir == modeHorizontal {
			// Going from vertical to horizontal.
			if cur == modeVertical && itm.data == " " {
				// there is only a whitespace element.
				continue
			}
			// now in horizontal mode, there can be more children in horizontal
			// mode, so append all of them to a single frontend.Text element
			if itm.typ == html.TextNode && cur == modeVertical {
				itm.data = strings.TrimLeft(itm.data, " ")
			}
			if te == nil {
				te = frontend.NewText()
				for _, itm := range prepend {
					te.Items = append(te.Items, itm)
				}
			}
			d.applySettings(te.Settings, styles)
			if err := d.collectHorizontalNodes(te, itm); err != nil {
				return err
			}
			cur = modeHorizontal
		} else {
			// still vertical
			if itm.data == "li" {
				styles.olCounter++
			}
			if te != nil {
				d.te = append(d.te, te)
				te = nil
			}
			d.output(itm, currentWidth)
		}
	}
	switch item.data {
	case "ul", "ol":
		vspace := frontend.NewText()
		vspace.Settings[frontend.SettingMarginBottom] = styles.marginBottom
		d.te = append(d.te, vspace)
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
	color              *color.Color
	fontfamily         *frontend.FontFamily
	fontfeatures       []string
	fontsize           bag.ScaledPoint
	fontstyle          frontend.FontStyle
	fontweight         frontend.FontWeight
	halign             frontend.HorizontalAlignment
	indent             bag.ScaledPoint
	indentRows         int
	language           string
	lineheight         bag.ScaledPoint
	listStyleType      string
	marginBottom       bag.ScaledPoint
	marginTop          bag.ScaledPoint
	paddingInlineStart bag.ScaledPoint
	olCounter          int
	preserveWhitespace bool
	valign             frontend.VerticalAlignment
	yoffset            bag.ScaledPoint
}

func (is *formattingStyles) clone() *formattingStyles {
	// inherit
	newFeatures := make([]string, len(is.fontfeatures))
	for i, f := range is.fontfeatures {
		newFeatures[i] = f
	}
	newis := &formattingStyles{
		color:              is.color,
		fontfamily:         is.fontfamily,
		fontfeatures:       newFeatures,
		fontsize:           is.fontsize,
		fontstyle:          is.fontstyle,
		fontweight:         is.fontweight,
		language:           is.language,
		lineheight:         is.lineheight,
		listStyleType:      is.listStyleType,
		olCounter:          is.olCounter,
		preserveWhitespace: is.preserveWhitespace,
		valign:             is.valign,
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
