package document

import (
	"github.com/speedata/boxesandglue/backend/bag"
	"github.com/speedata/boxesandglue/backend/color"
	"github.com/speedata/boxesandglue/backend/node"
	"github.com/speedata/boxesandglue/frontend"
)

func (d *Document) outputAtVertical(te *frontend.Text, width bag.ScaledPoint) (ret node.Node, err error) {
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
	if c, ok := te.Settings[frontend.SettingBackgroundColor]; ok {
		hv.BackgroundColor = c.(*color.Color)
	}
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
	if lw, ok := te.Settings[frontend.SettingWidth]; ok {
		if lws, ok := lw.(string); ok {
			lineWidth = parseRelativeSize(lws, lineWidth, d.defaultFontsize)
		}
	}
	lineWidth = lineWidth - hv.MarginLeft - hv.MarginRight - hv.PaddingLeft - hv.PaddingRight - hv.BorderLeftWidth - hv.BorderRightWidth
	if bx, ok := te.Settings[frontend.SettingBox]; ok && bx.(bool) {
		var newvl node.Node
		var n node.Node
		var prevMarginBottom bag.ScaledPoint
		for _, itm := range te.Items {
			if txt, ok := itm.(*frontend.Text); ok {
				if mt, ok := txt.Settings[frontend.SettingMarginTop]; ok {
					if marginTop, ok := mt.(bag.ScaledPoint); ok {
						g := node.NewGlue()
						if marginTop > prevMarginBottom {
							g.Width = marginTop
						} else {
							g.Width = prevMarginBottom
						}
						if g.Width != 0 {
							g.Attributes = node.H{"origin": "margin bottom + margin top (collapse)"}
							newvl = node.InsertAfter(newvl, node.Tail(newvl), g)
						}
					}
				}
				n, err = d.outputAtVertical(txt, lineWidth)
				if err != nil {
					return
				}
				if hv.MarginLeft > 0 {
					g := node.NewGlue()
					g.Width = hv.MarginLeft
					g.Attributes = node.H{"origin": "margin left"}
					node.InsertAfter(g, g, n)
					n = node.Hpack(g)
				}
				if prepend, ok := txt.Settings[frontend.SettingPrepend]; ok {
					if p, ok := prepend.(node.Node); ok {
						g := node.NewGlue()
						g.Stretch = bag.Factor
						g.Shrink = bag.Factor
						g.StretchOrder = node.StretchFil
						g.ShrinkOrder = node.StretchFil
						p = node.InsertBefore(p, p, g)
						p = node.HpackTo(p, 0)
						p.(*node.HList).Depth = 0
						n = node.InsertAfter(p, node.Tail(p), n)
						hl := node.Hpack(n)
						hl.VAlign = node.VAlignTop
						n = hl

					}
				}
				newvl = node.InsertAfter(newvl, node.Tail(newvl), n)

				if mb, ok := txt.Settings[frontend.SettingMarginBottom]; ok {
					if marginBottom, ok := mb.(bag.ScaledPoint); ok {
						prevMarginBottom = marginBottom
					}
				}
			}
		}
		if prevMarginBottom > 0 {
			g := node.NewGlue()
			g.Width = prevMarginBottom
			g.Attributes = node.H{"origin": "margin bottom"}
			newvl = node.InsertAfter(newvl, node.Tail(newvl), g)
		}
		vl := node.Vpack(newvl)
		vl = d.doc.HTMLBorder(vl, hv)
		ret = vl
		return
	}

	var vl *node.VList
	vl, _, err = d.doc.FormatParagraph(te, lineWidth, opts...)
	if err != nil {
		return
	}
	vl = d.doc.HTMLBorder(vl, hv)
	ml := node.NewGlue()
	mr := node.NewGlue()
	ml.Width = hv.MarginLeft
	mr.Width = hv.MarginRight
	ml.Attributes = node.H{"origin": "margin left"}
	mr.Attributes = node.H{"origin": "margin right"}
	var n node.Node
	n = node.InsertBefore(vl, vl, ml)
	node.InsertAfter(n, vl, mr)
	n = node.Hpack(n)
	vl = node.Vpack(n)
	ret = vl
	return
}

func (d *Document) outputToVList(text *frontend.Text, width bag.ScaledPoint) (node.Node, error) {
	return d.outputAtVertical(text, width)
}

func hasContents(areaAttributes map[string]string) bool {
	return areaAttributes["content"] != "none" && areaAttributes["content"] != "normal"
}

type pageMarginBox struct {
	minWidth    bag.ScaledPoint
	maxWidth    bag.ScaledPoint
	areaWidth   bag.ScaledPoint
	hasContents bool
	widthAuto   bool
	halign      frontend.HorizontalAlignment
	x           bag.ScaledPoint
	y           bag.ScaledPoint
	wd          bag.ScaledPoint
	ht          bag.ScaledPoint
}

// turn content: `"page " counter(page) " of " counter(pages)` into a meaningful
// string.
func (d *Document) parseContent(in string) string {
	var result []rune
	inString := false
	for _, r := range in {
		switch r {
		case '"':
			inString = !inString
		default:
			if inString {
				result = append(result, r)
			}
		}
	}
	return string(result)
}

func (d *Document) beforeShipout() error {
	dimensions := d.currentPageDimensions
	mp := dimensions.masterpage
	if mp != nil {
		pageMarginBoxes := make(map[string]*pageMarginBox)
		for areaName, attr := range mp.PageArea {
			pmb := &pageMarginBox{
				widthAuto: true,
			}
			pmb.hasContents = hasContents(attr)
			if wd, ok := attr["width"]; ok {
				if wd != "auto" {
					pmb.areaWidth = parseRelativeSize(wd, dimensions.Width, dimensions.Width)
				}
			}

			pageMarginBoxes[areaName] = pmb
		}
		for areaName := range mp.PageArea {
			pmb := pageMarginBoxes[areaName]
			switch areaName {
			case "top-left-corner":
				pmb.x = 0
				pmb.y = d.doc.Doc.DefaultPageHeight
				pmb.wd = dimensions.MarginLeft
				pmb.ht = dimensions.MarginTop
			case "top-right-corner":
				pmb.x = dimensions.Width - dimensions.MarginRight
				pmb.y = d.doc.Doc.DefaultPageHeight
				pmb.wd = dimensions.MarginRight
				pmb.ht = dimensions.MarginTop
			case "bottom-left-corner":
				pmb.x = 0
				pmb.y = dimensions.MarginBottom
				pmb.wd = dimensions.MarginLeft
				pmb.ht = dimensions.MarginBottom
			case "bottom-right-corner":
				pmb.x = dimensions.Width - dimensions.MarginRight
				pmb.y = dimensions.MarginBottom
				pmb.wd = dimensions.MarginRight
				pmb.ht = dimensions.MarginBottom
			case "top-left", "top-center", "top-right":
				pmb.x = dimensions.MarginLeft
				pmb.y = d.doc.Doc.DefaultPageHeight
				pmb.wd = dimensions.Width - dimensions.MarginLeft - dimensions.MarginRight
				pmb.ht = dimensions.MarginTop
				switch areaName {
				case "top-left":
					pmb.halign = frontend.HAlignLeft
				case "top-center":
					pmb.halign = frontend.HAlignCenter
				case "top-right":
					pmb.halign = frontend.HAlignRight
				}
			case "bottom-left", "bottom-center", "bottom-right":
				pmb.x = dimensions.MarginLeft
				pmb.y = dimensions.MarginTop
				pmb.wd = dimensions.Width - dimensions.MarginLeft - dimensions.MarginRight
				pmb.ht = dimensions.MarginTop
				switch areaName {
				case "bottom-left":
					pmb.halign = frontend.HAlignLeft
				case "bottom-center":
					pmb.halign = frontend.HAlignCenter
				case "bottom-right":
					pmb.halign = frontend.HAlignRight
				}
			}
		}
		// todo: calculate the area size
		for _, areaName := range []string{"top-left-corner", "top-left", "top-center", "top-right", "top-right-corner", "right-top", "right-middle", "right-bottom", "bottom-right-corner", "bottom-right", "bottom-center", "bottom-left", "bottom-left-corner", "left-bottom", "left-middle", "left-top"} {
			if area, ok := mp.PageArea[areaName]; ok {
				if !hasContents(area) {
					continue
				}
				styles := d.pushStyles()
				d.stylesToStyles(styles, area)
				pmb := pageMarginBoxes[areaName]

				vl := node.NewVList()
				var err error
				if c, ok := area["content"]; ok {
					c = d.parseContent(c)
					if c != "" {
						txt := frontend.NewText()
						d.applySettings(txt.Settings, styles)
						txt.Settings[frontend.SettingSize] = d.defaultFontsize
						txt.Settings[frontend.SettingHeight] = pmb.ht - styles.borderTopWidth - styles.borderBottomWidth
						txt.Settings[frontend.SettingVAlign] = styles.valign

						txt.Items = append(txt.Items, c)
						vl, _, err = d.doc.FormatParagraph(txt, pmb.wd-styles.borderLeftWidth-styles.borderRightWidth, frontend.Family(d.defaultFontFamily), frontend.HorizontalAlign(pmb.halign))
						if err != nil {
							return err
						}
					} else {
						vl = node.NewVList()
						vl.Width = pmb.wd - styles.borderLeftWidth - styles.borderRightWidth
						vl.Height = pmb.ht - styles.borderTopWidth - styles.borderBottomWidth
					}
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
						BackgroundColor:         styles.backgroundColor,
					}
					vl = d.doc.HTMLBorder(vl, hv)
					d.doc.Doc.CurrentPage.OutputAt(pmb.x, pmb.y, vl)
					d.popStyles()

				}
			}
		}
	}
	return nil
}
