# bagme - boxes and glue made easy

bagme is a proof of concept using the “boxes and glue” PDF library to format HTML fragments styled with CSS.
boxes and glue implements TeX's typesetting algorithms, so the idea is to get superb HTML rendering with almost no effort.

The goal is to have a PDF rendering engine for your Go software without having to do a lot of programming.


## Sample code

```go
package main

import (
	"log"

	"github.com/speedata/bagme/document"

	"github.com/speedata/boxesandglue/backend/bag"
)

var css = `
@font-face {
    font-family: CrimsonPro;
    src: url("fonts/crimsonpro/CrimsonPro-Regular.ttf");
}

@font-face {
    font-family: CrimsonPro;
    src: url("fonts/crimsonpro/CrimsonPro-Bold.ttf");
    font-weight: bold;
}

@font-face {
    font-family: CrimsonPro;
    src: url("fonts/crimsonpro/CrimsonPro-Italic.ttf");
    font-style: italic;
}

@font-face {
    font-family: CrimsonPro;
    src: url("fonts/crimsonpro/CrimsonPro-BoldItalic.ttf");
    font-style: italic;
    font-weight: bold;
}

body {
	font-family: CrimsonPro;
    font-size: 12pt;
    line-height: 13pt;
}

p {
    margin-top: 6pt;
    margin-bottom: 6pt;
}`

var html = `
<h1>The frog king</h1>

<p>In olden times when wishing still <em>helped</em> one,
   there lived a king whose daughters were all beautiful,
   but the youngest was so beautiful that the sun itself,
   which has seen so much, was
   <span style="font-weight: bold">astonished</span> whenever it
   shone in her face.</p>

<p>Close by the king's castle lay a great dark forest,
	and under an old lime-tree in the forest was a well,
	and when the day was very warm, the king's child
	went out into the forest and sat down by the side of
	the cool fountain, and when she was bored she took a
	golden ball, and threw it up on high and caught it,
	and this ball was her favorite plaything.
</p>`

func dothings() error {
	d, err := document.New("out.pdf")
	if err != nil {
		return err
	}
	if err = d.ParseCSSString(css); err != nil {
		return err
	}
	wd := bag.MustSp("280pt")
	col := bag.MustSp("1cm")
	row := bag.MustSp("23cm")
	if err = d.OutputAt(html, wd, col, row); err != nil {
		return err
	}
	return d.Finish()
}

func main() {
	if err := dothings(); err != nil {
		log.Fatal(err)
	}
}
```

<img src="https://i.imgur.com/xa9t10p.png" alt="typeset text from the frog king" width="500"/>


Contact: <gundlach@speedata.de><br>
Mastodon: [boxesandglue@typo.social](https://typo.social/@boxesandglue)

