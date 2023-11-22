# bagme - boxes and glue made easy

bagme is a PDF library to format HTML fragments styled with CSS using the (pure Go) “[boxes and glue](https://boxesandglue.dev)” library.
boxes and glue implements TeX's typesetting algorithms, so the idea is to get superb HTML rendering with almost no effort.

The goal is to have a PDF rendering engine for your Go software without having to do a lot of programming.


## Sample code

<img src="https://i.imgur.com/rGWsP8h.png" alt="typeset text from the frog king" width="500"/>


```go
package main

import (
	"log"

	"github.com/speedata/bagme/document"
	"github.com/speedata/boxesandglue/backend/bag"
)

var html = `<h1>The frog king</h1>

<p>In olden times when wishing still <em>helped</em> one,
   there lived a king whose daughters were all beautiful,
   but the <span class="green">youngest</span> was so beautiful that the sun itself,
   which has seen so much, was
   <span style="font-weight: bold">astonished</span> whenever it
   shone in her face.</p>

<p>Close by the king's castle lay a great dark forest,
	and under an old lime-tree in the forest was a well,
	and when the day was very warm, the king's child
	went out into the forest and sat down by the side of
	the cool <span id="important">fountain</span>, and when she was bored she took a
	golden ball, and threw it up on high and caught it,
	and this ball was her favorite plaything.
</p>`

var css = `
body {
	font-family: serif;
    font-size: 12pt;
    line-height: 14pt;
}

p {
    margin-top: 8pt;
    margin-bottom: 2pt;
}

.green {
    color: green;
}

#important {
    color: rebeccapurple;
    font-weight: bolder;
    font-style: italic;
}`

func dothings() error {
	d, err := document.New("out.pdf")
	if err != nil {
		return err
	}
	if d.AddCSS(css); err != nil {
		return err
	}
	// loads the font families serif, sans and monospace
	if err = d.Frontend.LoadIncludedFonts(); err != nil {
		return err
	}
	wd := bag.MustSp("280pt")
	colText := bag.MustSp("140pt")
	colImage := bag.MustSp("20pt")
	rowText := bag.MustSp("23cm")
	if err = d.OutputAt(html, wd, colText, rowText); err != nil {
		return err
	}

	if err = d.OutputAt(`<img src="img/frogking-a.pdf" width="4cm" height="6cm">`, wd, colImage, rowText); err != nil {
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

## Limitations

* No automatic page breaks: you have to take care of your items on the page.
* Limited HTML/CSS support: some things are implemented, most are not.

You can take a look at [the examples repository](https://github.com/speedata/bagme-examples) to see what is possible.

## Other

Contact: <gundlach@speedata.de><br>
License: New BSD License<br>
Status: Beta: You can try it, but expect API changes.<br>
Mastodon: [boxesandglue@typo.social](https://typo.social/@boxesandglue)

