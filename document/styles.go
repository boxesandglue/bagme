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
	if strings.HasSuffix(fs, "%") {
		p := strings.TrimSuffix(fs, "%")
		f, err := strconv.ParseFloat(p, 64)
		if err != nil {
			panic(err)
		}
		ret := bag.MultiplyFloat(dflt, f/100)
		return ret
	}
	if strings.HasSuffix(fs, "rem") {
		if root == 0 {
			bag.Logger.Warn("Calculating an rem size without a body font size results in a size of 0.")
			return 0
		}

		prefix := strings.TrimSuffix(fs, "rem")
		factor, err := strconv.ParseFloat(prefix, 32)
		if err != nil {
			bag.Logger.Errorf("Cannot convert relative size %s", fs)
			return bag.MustSp("10pt")
		}
		return bag.ScaledPoint(float64(root) * factor)
	}
	if strings.HasSuffix(fs, "em") {
		if dflt == 0 {
			bag.Logger.Warn("Calculating an em size without a body font size results in a size of 0.")
			return 0
		}
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
	settings[frontend.SettingBackgroundColor] = ih.backgroundColor
	settings[frontend.SettingBorderTopWidth] = ih.borderTopWidth
	settings[frontend.SettingBorderLeftWidth] = ih.borderLeftWidth
	settings[frontend.SettingBorderRightWidth] = ih.borderRightWidth
	settings[frontend.SettingBorderBottomWidth] = ih.borderBottomWidth
	settings[frontend.SettingBorderTopColor] = ih.borderTopColor
	settings[frontend.SettingBorderLeftColor] = ih.borderLeftColor
	settings[frontend.SettingBorderRightColor] = ih.borderRightColor
	settings[frontend.SettingBorderBottomColor] = ih.borderBottomColor
	settings[frontend.SettingBorderTopStyle] = ih.borderTopStyle
	settings[frontend.SettingBorderLeftStyle] = ih.borderLeftStyle
	settings[frontend.SettingBorderRightStyle] = ih.borderRightStyle
	settings[frontend.SettingBorderBottomStyle] = ih.borderBottomStyle
	settings[frontend.SettingBorderTopLeftRadius] = ih.borderTopLeftRadius
	settings[frontend.SettingBorderTopRightRadius] = ih.borderTopRightRadius
	settings[frontend.SettingBorderBottomLeftRadius] = ih.borderBottomLeftRadius
	settings[frontend.SettingBorderBottomRightRadius] = ih.borderBottomRightRadius
	settings[frontend.SettingColor] = ih.color
	if ih.fontexpansion != nil {
		settings[frontend.SettingFontExpansion] = *ih.fontexpansion
	} else {
		settings[frontend.SettingFontExpansion] = 0.05
	}
	settings[frontend.SettingFontFamily] = ih.fontfamily
	settings[frontend.SettingHAlign] = ih.halign
	settings[frontend.SettingHangingPunctuation] = ih.hangingPunctuation
	settings[frontend.SettingIndentLeft] = ih.indent
	settings[frontend.SettingIndentLeftRows] = ih.indentRows
	settings[frontend.SettingLeading] = ih.lineheight
	settings[frontend.SettingMarginBottom] = ih.marginBottom
	settings[frontend.SettingMarginRight] = ih.marginRight
	settings[frontend.SettingMarginLeft] = ih.marginLeft
	settings[frontend.SettingMarginTop] = ih.marginTop
	settings[frontend.SettingOpenTypeFeature] = ih.fontfeatures
	settings[frontend.SettingPaddingRight] = ih.paddingRight
	settings[frontend.SettingPaddingLeft] = ih.paddingLeft
	settings[frontend.SettingPaddingTop] = ih.paddingTop
	settings[frontend.SettingPaddingBottom] = ih.paddingBottom
	settings[frontend.SettingPreserveWhitespace] = ih.preserveWhitespace
	settings[frontend.SettingSize] = ih.fontsize
	settings[frontend.SettingStyle] = ih.fontstyle
	settings[frontend.SettingYOffset] = ih.yoffset
	settings[frontend.SettingTabSize] = ih.tabsize
	settings[frontend.SettingTabSizeSpaces] = ih.tabsizeSpaces

	if ih.width != "" {
		settings[frontend.SettingWidth] = ih.width
	}

}

func (d *Document) stylesToStyles(ih *formattingStyles, attributes map[string]string) error {
	// Resolve font size first, since some of the attributes depend on the
	// current font size.
	if v, ok := attributes["font-size"]; ok {
		ih.fontsize = parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
	}
	for k, v := range attributes {
		switch k {
		case "font-size":
			// already set
		case "hyphens":
			// ignore for now
		case "display":
			ih.hide = (v == "none")
		case "background-color":
			ih.backgroundColor = d.doc.GetColor(v)
		case "border-right-width", "border-left-width", "border-top-width", "border-bottom-width":
			size := parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
			switch k {
			case "border-right-width":
				ih.borderRightWidth = size
			case "border-left-width":
				ih.borderLeftWidth = size
			case "border-top-width":
				ih.borderTopWidth = size
			case "border-bottom-width":
				ih.borderBottomWidth = size
			}
		case "border-top-right-radius", "border-top-left-radius", "border-bottom-right-radius", "border-bottom-left-radius":
			size := parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
			switch k {
			case "border-top-right-radius":
				ih.borderTopRightRadius = size
			case "border-top-left-radius":
				ih.borderTopLeftRadius = size
			case "border-bottom-left-radius":
				ih.borderBottomLeftRadius = size
			case "border-bottom-right-radius":
				ih.borderBottomRightRadius = size
			}
		case "border-right-style", "border-left-style", "border-top-style", "border-bottom-style":
			var sty frontend.BorderStyle
			switch v {
			case "none":
				// default
			case "solid":
				sty = frontend.BorderStyleSolid
			default:
				bag.Logger.DPanicf("not implemented: border style %q", v)
			}
			switch k {
			case "border-right-style":
				ih.borderRightStyle = sty
			case "border-left-style":
				ih.borderLeftStyle = sty
			case "border-top-style":
				ih.borderTopStyle = sty
			case "border-bottom-style":
				ih.borderBottomStyle = sty
			}

		case "border-right-color":
			ih.borderRightColor = d.c.FrontendDocument.GetColor(v)
		case "border-left-color":
			ih.borderLeftColor = d.c.FrontendDocument.GetColor(v)
		case "border-top-color":
			ih.borderTopColor = d.c.FrontendDocument.GetColor(v)
		case "border-bottom-color":
			ih.borderBottomColor = d.c.FrontendDocument.GetColor(v)
		case "border-spacing":
			// ignore
		case "color":
			ih.color = d.c.FrontendDocument.GetColor(v)
		case "content":
			// ignore
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
			if ih.fontfamily == nil {
				bag.Logger.Errorf("font family %q not found", v)
				return fmt.Errorf("font family %q not found", v)
			}
		case "hanging-punctuation":
			switch v {
			case "allow-end":
				ih.hangingPunctuation = frontend.HangingPunctuationAllowEnd
			}
		case "line-height":
			ih.lineheight = parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
		case "margin-bottom":
			ih.marginBottom = parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
		case "margin-left":
			ih.marginLeft = parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
		case "margin-right":
			ih.marginRight = parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
		case "margin-top":
			ih.marginTop = parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
		case "padding-inline-start":
			ih.paddingInlineStart = parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
		case "padding-bottom":
			ih.paddingBottom = parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
		case "padding-left":
			ih.paddingLeft = parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
		case "padding-right":
			ih.paddingRight = parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
		case "padding-top":
			ih.paddingTop = parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
		case "tab-size":
			if ts, err := strconv.Atoi(v); err == nil {
				ih.tabsizeSpaces = ts
			} else {
				ih.tabsize = parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
			}
		case "text-align":
			ih.halign = parseHorizontalAlign(v, ih)
		case "text-indent":
			ih.indent = parseRelativeSize(v, d.currentStyle().fontsize, d.defaultFontsize)
			ih.indentRows = 1
		case "user-select":
			// ignore
		case "vertical-align":
			if v == "sub" {
				ih.yoffset = -1 * ih.fontsize * 1000 / 5000
			} else if v == "super" {
				ih.yoffset = ih.fontsize * 1000 / 5000
			}
		case "width":
			ih.width = v
		case "white-space":
			ih.preserveWhitespace = (v == "pre")
		case "-bag-font-expansion":
			if strings.HasSuffix(v, "%") {
				p := strings.TrimSuffix(v, "%")
				f, err := strconv.ParseFloat(p, 64)
				if err != nil {
					return err
				}
				fe := f / 100
				ih.fontexpansion = &fe
			}
		default:
			fmt.Println("unresolved attribute", k, v)
		}
	}
	return nil
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
			if err := d.stylesToStyles(sty, item.styles); err != nil {
				return err
			}
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

func (d *Document) output(item *htmlItem, currentWidth bag.ScaledPoint) (*frontend.Text, error) {
	// item is guaranteed to be in vertical direction
	newte := frontend.NewText()
	styles := d.pushStyles()
	if err := d.stylesToStyles(styles, item.styles); err != nil {
		return nil, err
	}
	d.applySettings(newte.Settings, styles)
	newte.Settings[frontend.SettingDebug] = item.data
	switch item.data {
	case "html":
		if fs, ok := item.styles["font-size"]; ok {
			rfs := parseRelativeSize(fs, 0, 0)
			d.defaultFontsize = rfs
		}
		if ffs, ok := item.styles["font-family"]; ok {
			ff := d.doc.FindFontFamily(ffs)
			d.defaultFontFamily = ff
		}
	case "body":
		if ffs, ok := item.styles["font-family"]; ok {
			ff := d.doc.FindFontFamily(ffs)
			d.defaultFontFamily = ff
		}
	case "table":
		txt, err := d.processTable(item, currentWidth)
		d.popStyles()
		if err != nil {
			return nil, err
		}
		return txt, nil
	case "ol", "ul":
		styles.olCounter = 0
	case "li":
		var item string
		if strings.HasPrefix(styles.listStyleType, `"`) && strings.HasSuffix(styles.listStyleType, `"`) {
			item = strings.TrimPrefix(styles.listStyleType, `"`)
			item = strings.TrimSuffix(item, `"`)
		} else {
			switch styles.listStyleType {
			case "disc":
				item = "•"
			case "circle":
				item = "◦"
			case "none":
				item = ""
			case "square":
				item = "□"
			case "decimal":
				item = fmt.Sprintf("%d.", styles.olCounter)
			default:
				bag.Logger.Errorf("unhandled list-style-type: %q", styles.listStyleType)
				item = "•"
			}
			item += " "
		}
		n, err := d.doc.BuildNodelistFromString(newte.Settings, item)
		if err != nil {
			return nil, err
		}
		newte.Settings[frontend.SettingPrepend] = n
	}

	var te *frontend.Text
	cur := modeVertical

	// display = "none"
	if styles.hide {
		d.popStyles()
		return newte, nil
	}

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
				styles = d.pushStyles()
			}
			d.applySettings(te.Settings, styles)
			if err := d.collectHorizontalNodes(te, itm); err != nil {
				return nil, err
			}
			cur = modeHorizontal
		} else {
			// still vertical
			if itm.data == "li" {
				styles.olCounter++
			}
			if te != nil {
				newte.Items = append(newte.Items, te)
				te = nil
			}
			te, err := d.output(itm, currentWidth)
			if err != nil {
				return nil, err
			}
			if len(te.Items) > 0 {
				newte.Items = append(newte.Items, te)
			}
		}
	}
	if item.dir == modeVertical && cur == modeVertical {
		newte.Settings[frontend.SettingBox] = true
	}
	switch item.data {
	case "ul", "ol":
		ulte := frontend.NewText()
		d.applySettings(ulte.Settings, styles)
		ulte.Settings[frontend.SettingDebug] = item.data
		ulte.Settings[frontend.SettingBox] = true
	}
	if te != nil {
		newte.Items = append(newte.Items, te)
		d.popStyles()
		te = nil
	}
	d.popStyles()
	return newte, nil
}

func (d *Document) parseSelection(sel *goquery.Selection, wd bag.ScaledPoint) (*frontend.Text, error) {
	h := &htmlItem{dir: modeVertical}
	dumpElement(sel.Nodes[0], modeVertical, h)
	return d.output(h, wd)
}

type formattingStyles struct {
	backgroundColor         *color.Color
	borderLeftWidth         bag.ScaledPoint
	borderRightWidth        bag.ScaledPoint
	borderBottomWidth       bag.ScaledPoint
	borderTopWidth          bag.ScaledPoint
	borderTopLeftRadius     bag.ScaledPoint
	borderTopRightRadius    bag.ScaledPoint
	borderBottomLeftRadius  bag.ScaledPoint
	borderBottomRightRadius bag.ScaledPoint
	borderLeftColor         *color.Color
	borderRightColor        *color.Color
	borderBottomColor       *color.Color
	borderTopColor          *color.Color
	borderLeftStyle         frontend.BorderStyle
	borderRightStyle        frontend.BorderStyle
	borderBottomStyle       frontend.BorderStyle
	borderTopStyle          frontend.BorderStyle
	color                   *color.Color
	hide                    bool
	fontfamily              *frontend.FontFamily
	fontfeatures            []string
	fontsize                bag.ScaledPoint
	fontstyle               frontend.FontStyle
	fontweight              frontend.FontWeight
	fontexpansion           *float64
	halign                  frontend.HorizontalAlignment
	hangingPunctuation      frontend.HangingPunctuation
	indent                  bag.ScaledPoint
	indentRows              int
	language                string
	lineheight              bag.ScaledPoint
	listStyleType           string
	marginBottom            bag.ScaledPoint
	marginLeft              bag.ScaledPoint
	marginRight             bag.ScaledPoint
	marginTop               bag.ScaledPoint
	paddingInlineStart      bag.ScaledPoint
	paddingBottom           bag.ScaledPoint
	paddingLeft             bag.ScaledPoint
	paddingRight            bag.ScaledPoint
	paddingTop              bag.ScaledPoint
	olCounter               int
	preserveWhitespace      bool
	tabsize                 bag.ScaledPoint
	tabsizeSpaces           int
	valign                  frontend.VerticalAlignment
	width                   string
	yoffset                 bag.ScaledPoint
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
		hangingPunctuation: is.hangingPunctuation,
		fontexpansion:      is.fontexpansion,
		language:           is.language,
		lineheight:         is.lineheight,
		listStyleType:      is.listStyleType,
		olCounter:          is.olCounter,
		preserveWhitespace: is.preserveWhitespace,
		tabsize:            is.tabsize,
		tabsizeSpaces:      is.tabsizeSpaces,
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
