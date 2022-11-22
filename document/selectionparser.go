package document

import (
	"fmt"
	"regexp"

	"github.com/speedata/boxesandglue/csshtml"
	"golang.org/x/net/html"
)

var (
	isSpace          = regexp.MustCompile(`^\s*$`)
	reLeadcloseWhtsp = regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
	reInsideWS       = regexp.MustCompile(`\n|[\s\p{Zs}]{2,}`) //to match 2 or more whitespace symbols inside a string or NL
)

type mode int

func (m mode) String() string {
	if m == modeHorizontal {
		return "→"
	}
	return "↓"
}

const (
	modeHorizontal mode = iota
	modeVertical
)

var preserveWhitespace = []bool{false}

type htmlItem struct {
	typ        html.NodeType
	data       string
	dir        mode
	attributes map[string]string
	styles     map[string]string
	children   []*htmlItem
}

func (itm *htmlItem) String() string {
	switch itm.typ {
	case html.TextNode:
		return fmt.Sprintf("%q", itm.data)
	case html.ElementNode:
		return fmt.Sprintf("<%s>", itm.data)
	default:
		return fmt.Sprintf("%s", itm.data)
	}
}

func dumpElement(thisNode *html.Node, direction mode, firstItem *htmlItem) {
	newDir := direction
	for {
		if thisNode == nil {
			break
		}

		switch thisNode.Type {
		case html.CommentNode:
			// ignore
		case html.TextNode:
			itm := &htmlItem{}
			ws := preserveWhitespace[len(preserveWhitespace)-1]
			txt := thisNode.Data
			if !ws {
				if isSpace.MatchString(txt) {
					txt = " "
				}
			}
			if !isSpace.MatchString(txt) {
				if direction == modeVertical {
					newDir = modeHorizontal
				}
			}
			if txt != "" {
				if !ws {
					txt = reLeadcloseWhtsp.ReplaceAllString(txt, " ")
					txt = reInsideWS.ReplaceAllString(txt, " ")
				}
			}
			itm.data = txt
			itm.typ = html.TextNode
			firstItem.children = append(firstItem.children, itm)
		case html.ElementNode:
			ws := preserveWhitespace[len(preserveWhitespace)-1]
			eltname := thisNode.Data

			if eltname == "body" || eltname == "address" || eltname == "article" || eltname == "aside" || eltname == "blockquote" || eltname == "br" || eltname == "canvas" || eltname == "dd" || eltname == "div" || eltname == "dl" || eltname == "dt" || eltname == "fieldset" || eltname == "figcaption" || eltname == "figure" || eltname == "footer" || eltname == "form" || eltname == "h1" || eltname == "h2" || eltname == "h3" || eltname == "h4" || eltname == "h5" || eltname == "h6" || eltname == "header" || eltname == "hr" || eltname == "li" || eltname == "main" || eltname == "nav" || eltname == "noscript" || eltname == "ol" || eltname == "p" || eltname == "pre" || eltname == "section" || eltname == "table" || eltname == "tfoot" || eltname == "thead" || eltname == "tbody" || eltname == "tr" || eltname == "td" || eltname == "th" || eltname == "ul" || eltname == "video" {
				newDir = modeVertical
			} else if eltname == "b" || eltname == "big" || eltname == "i" || eltname == "small" || eltname == "tt" || eltname == "abbr" || eltname == "acronym" || eltname == "cite" || eltname == "code" || eltname == "dfn" || eltname == "em" || eltname == "kbd" || eltname == "strong" || eltname == "samp" || eltname == "var" || eltname == "a" || eltname == "bdo" || eltname == "img" || eltname == "map" || eltname == "object" || eltname == "q" || eltname == "script" || eltname == "span" || eltname == "sub" || eltname == "sup" || eltname == "button" || eltname == "input" || eltname == "label" || eltname == "select" || eltname == "textarea" {
				newDir = modeHorizontal
			} else {
				// keep dir
			}

			itm := &htmlItem{
				typ:  html.ElementNode,
				data: thisNode.Data,
				dir:  newDir,
			}
			firstItem.children = append(firstItem.children, itm)
			attributes := thisNode.Attr
			if len(attributes) > 0 {
				itm.styles, itm.attributes = csshtml.ResolveAttributes(attributes)
				for key, value := range itm.styles {
					if key == "white-space" {
						if value == "pre" {
							ws = true
						} else {
							ws = false
						}
					}
				}
			}
			if thisNode.FirstChild != nil {
				preserveWhitespace = append(preserveWhitespace, ws)
				dumpElement(thisNode.FirstChild, newDir, itm)
				preserveWhitespace = preserveWhitespace[:len(preserveWhitespace)-1]
			}
		default:
			fmt.Println(thisNode.Type)
		}
		thisNode = thisNode.NextSibling
	}
}
