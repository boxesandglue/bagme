[![Explore in Constellation](https://img.shields.io/badge/Explore%20in-Constellation-blue)](https://constellation.speedata.de/#bagme)

# bagme - boxes and glue made easy

bagme is a Go library that renders HTML/CSS to PDF using the [boxes and glue](https://boxesandglue.dev) typesetting engine.
It implements TeX's line-breaking and page-breaking algorithms, so you get high-quality typesetting with almost no effort.

<p align="center">
  <a href="https://github.com/boxesandglue/boxesandglue-examples/tree/main/bagme/basic/multipage">
    <img src="https://raw.githubusercontent.com/boxesandglue/boxesandglue-examples/main/bagme/basic/multipage/result.png" width="320" alt="Multi-page HTML rendered to PDF by bagme">
  </a>
</p>

Ready-to-run examples (HTML → PDF, accessible PDF/UA, tables, positioned snippets, multi-page documents) live in the [boxesandglue-examples repository](https://github.com/boxesandglue/boxesandglue-examples/tree/main/bagme).

## Installation

```bash
go get github.com/boxesandglue/bagme
```

The [bagme manual](https://boxesandglue.dev/bagme/) describes the API, the [htmlbag manual](https://boxesandglue.dev/htmlbag/) the supported HTML and CSS.

## Render a complete HTML document

Use `RenderPages` for automatic page breaks, `@page` margins, and multi-page output:

```go
package main

import (
	"log"

	"github.com/boxesandglue/bagme/document"
)

func main() {
	d, err := document.New("output.pdf")
	if err != nil {
		log.Fatal(err)
	}
	d.AddCSS(`
		@page { size: A4; margin: 2cm; }
		body { font-family: serif; font-size: 11pt; line-height: 1.4; }
		h1 { page-break-before: always; }
	`)
	if err := d.RenderPages(`
		<h1>Chapter 1</h1>
		<p>In olden times when wishing still helped one, there lived a king
		whose daughters were all beautiful...</p>

		<h1>Chapter 2</h1>
		<p>Close by the king's castle lay a great dark forest, and under an
		old lime-tree in the forest was a well...</p>
	`); err != nil {
		log.Fatal(err)
	}
	if err := d.Finish(); err != nil {
		log.Fatal(err)
	}
}
```

## Place HTML snippets at exact positions

Use `OutputAt` for precise placement on a page (labels, forms, letterheads):

```go
package main

import (
	"log"

	"github.com/boxesandglue/bagme/document"
	"github.com/boxesandglue/boxesandglue/backend/bag"
)

func main() {
	d, err := document.New("out.pdf")
	if err != nil {
		log.Fatal(err)
	}
	d.AddCSS(`body { font-family: serif; font-size: 12pt; }`)

	ps, _ := d.PageSize()
	w := ps.ContentWidth
	x := ps.MarginLeft
	y := ps.Height - ps.MarginTop

	d.OutputAt("<h1>Hello, World!</h1><p>Placed at exact coordinates.</p>", w, x, y)
	d.Finish()
}
```

## Features

- Automatic page breaks with CSS `@page` rules
- `page-break-before` / `page-break-after` (always, avoid)
- Page margin boxes (`@top-center`, `@bottom-right`, etc.) for headers/footers
- CSS styling: fonts, colors, margins, padding, borders (rounded), backgrounds
- Tables with colspan/rowspan, borders, and cell backgrounds
- Ordered and unordered lists
- Images (PDF, PNG) and inline SVG
- OpenType features and variable fonts via `font-feature-settings` / `font-variation-settings`
- TeX-quality line breaking (Knuth-Plass algorithm)
- Heading extraction for table of contents generation
- Pure Go — no C dependencies, no browser, single binary

## Limitations

* Limited CSS support compared to a full browser engine, see the [htmlbag limitations](https://boxesandglue.dev/htmlbag/limitations/)
* No flexbox or grid layout

## Examples

Source + rendered PDF + preview PNG for every entry in [boxesandglue-examples/bagme](https://github.com/boxesandglue/boxesandglue-examples/tree/main/bagme):

Example | What it shows
------- | -------------
[Simple](https://github.com/boxesandglue/boxesandglue-examples/tree/main/bagme/basic/simple) | Basic HTML rendering with CSS styling, lists and code blocks
[Border](https://github.com/boxesandglue/boxesandglue-examples/tree/main/bagme/basic/border) | Borders, border-radius and background colors on elements
[Accessible PDF](https://github.com/boxesandglue/boxesandglue-examples/tree/main/bagme/basic/accessible) | PDF/UA output with automatic structure tagging for screen readers
[Multipage](https://github.com/boxesandglue/boxesandglue-examples/tree/main/bagme/basic/multipage) | Automatic page breaks, forced breaks with `page-break-before`
[Positioned](https://github.com/boxesandglue/boxesandglue-examples/tree/main/bagme/basic/position) | Manual placement with `OutputAt` using absolute coordinates
[Table](https://github.com/boxesandglue/boxesandglue-examples/tree/main/bagme/basic/table) | Tables with repeating headers across page breaks
[Math](https://github.com/boxesandglue/boxesandglue-examples/tree/main/bagme/basic/math) | Inline and display MathML, tagged for PDF/UA-2
[Barcode](https://github.com/boxesandglue/boxesandglue-examples/tree/main/bagme/basic/barcode) | QR code, EAN-13 and Code 128 with the `<barcode>` element
[ZUGFeRD](https://github.com/boxesandglue/boxesandglue-examples/tree/main/bagme/formats/zugferd) | ZUGFeRD / Factur-X invoice as PDF/A-3b with the invoice XML attached
[Accessible ZUGFeRD](https://github.com/boxesandglue/boxesandglue-examples/tree/main/bagme/formats/zugferd-accessible) | The same invoice, also tagged as PDF/UA-1

## Ecosystem

bagme is part of a broader ecosystem of PDF, typesetting and publishing technologies.

**[Explore the constellation →](https://constellation.speedata.de/#bagme)**

## Contact

Contact: <gundlach@speedata.de><br>
License: New BSD License<br>
Mastodon: [boxesandglue@typo.social](https://typo.social/@boxesandglue)
