package document

import (
	"strconv"
	"strings"

	"github.com/speedata/boxesandglue/backend/bag"
	"github.com/speedata/boxesandglue/backend/node"
	"github.com/speedata/boxesandglue/frontend"
)

func (d *Document) processTr(item *htmlItem) (*frontend.TableRow, error) {
	tr := &frontend.TableRow{}
	for _, itm := range item.children {
		if itm.data == "td" {
			styles := d.pushStyles()
			tc := &frontend.TableCell{}
			for k, v := range itm.styles {
				switch k {
				case "padding-top":
					tc.PaddingTop = bag.MustSp(v)
				case "padding-bottom":
					tc.PaddingBottom = bag.MustSp(v)
				case "padding-left":
					tc.PaddingLeft = bag.MustSp(v)
				case "padding-right":
					tc.PaddingRight = bag.MustSp(v)
				case "border-top-width":
					tc.BorderTopWidth = bag.MustSp(v)
				case "border-bottom-width":
					tc.BorderBottomWidth = bag.MustSp(v)
				case "border-left-width":
					tc.BorderLeftWidth = bag.MustSp(v)
				case "border-right-width":
					tc.BorderRightWidth = bag.MustSp(v)
				case "border-top-color":
					tc.BorderTopColor = d.doc.GetColor(v)
				case "border-bottom-color":
					tc.BorderBottomColor = d.doc.GetColor(v)
				case "border-left-color":
					tc.BorderLeftColor = d.doc.GetColor(v)
				case "border-right-color":
					tc.BorderRightColor = d.doc.GetColor(v)
				case "vertical-align":
					styles.valign = parseVerticalAlign(v, styles)
				}
			}
			for k, v := range itm.attributes {
				switch k {
				case "rowspan":
					rs, err := strconv.Atoi(v)
					if err != nil {
						return nil, err
					}
					tc.ExtraRowspan = rs - 1
				case "colspan":
					cs, err := strconv.Atoi(v)
					if err != nil {
						return nil, err
					}
					tc.ExtraColspan = cs - 1
				}
			}
			tc.VAlign = styles.valign
			if err := d.output(itm, 0); err != nil {
				return nil, err
			}
			for _, te := range d.te {
				tc.Contents = append(tc.Contents, te)
			}
			d.te = d.te[:0]
			tr.Cells = append(tr.Cells, tc)
			d.popStyles()
		}
	}
	return tr, nil
}

func (d *Document) processTbody(item *htmlItem) (frontend.TableRows, error) {
	var trows frontend.TableRows
	for _, itm := range item.children {
		if itm.data == "tr" {
			styles := d.pushStyles()
			for k, v := range itm.styles {
				switch k {
				case "vertical-align":
					styles.valign = parseVerticalAlign(v, styles)
				}
			}

			tr, err := d.processTr(itm)
			if err != nil {
				return nil, err
			}
			d.popStyles()
			trows = append(trows, tr)
		}
	}
	return trows, nil
}

func (d *Document) processTable(item *htmlItem, maxwd bag.ScaledPoint) error {
	saveText := d.te
	d.te = []*frontend.Text{}
	tbl := &frontend.Table{}
	tbl.Stretch = false
	tbl.MaxWidth = maxwd
	tableText := frontend.NewText()

	for k, v := range item.styles {
		switch k {
		case "margin-top":
			m := parseRelativeSize(v, d.currentStyle().fontsize)
			tableText.Settings[frontend.SettingMarginTop] = m
		case "margin-bottom":
			m := parseRelativeSize(v, d.currentStyle().fontsize)
			tableText.Settings[frontend.SettingMarginBottom] = m
		case "margin-left":
			m := parseRelativeSize(v, d.currentStyle().fontsize)
			tableText.Settings[frontend.SettingMarginLeft] = m
		case "margin-right":
			m := parseRelativeSize(v, d.currentStyle().fontsize)
			tableText.Settings[frontend.SettingMarginRight] = m
		case "width":
			if k == "auto" {
				// ignore, settings ok
			} else if strings.HasSuffix(v, "%") {
				wd, err := parseWidth(v, maxwd)
				if err != nil {
					return err
				}
				tbl.MaxWidth = wd
				tbl.Stretch = true
			}
		}
	}

	var rows frontend.TableRows
	var err error
	for _, itm := range item.children {
		if itm.data == "tbody" {
			styles := d.pushStyles()
			for k, v := range itm.styles {
				switch k {
				case "vertical-align":
					styles.valign = parseVerticalAlign(v, styles)
				}
			}
			if rows, err = d.processTbody(itm); err != nil {
				return err
			}
			d.popStyles()
		}
	}

	tbl.Rows = rows
	vl, err := d.doc.BuildTable(tbl)
	if err != nil {
		return err
	}
	hl := node.Hpack(vl[0])
	tableText.Items = append(tableText.Items, hl)
	d.te = saveText
	d.te = append(d.te, tableText)
	return nil
}
