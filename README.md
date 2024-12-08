# go-pdf2X
[![PDF2Text](https://github.com/nextunit-io/go-pdf2X/actions/workflows/pdf2text.yml/badge.svg?branch=main)](https://github.com/nextunit-io/go-pdf2X/actions/workflows/pdf2text.yml)
[![PDF2Html](https://github.com/nextunit-io/go-pdf2X/actions/workflows/pdf2html.yml/badge.svg?branch=main)](https://github.com/nextunit-io/go-pdf2X/actions/workflows/pdf2html.yml)

## pdf2text

Lib to abstract the pdftotext cli library

### Preconditions

For this library it is necessary that `pdftotext` is installed in the version above `24.11.x - 25.x.x`. 
With homebrew it is possible to install it via `brew install poppler` [Homebrew Poppler](https://formulae.brew.sh/formula/poppler).

### Usage

This is a module for an easy usage in GO. It abstracts the CLI commands in go methods that can be used for further development.

```go
package main

import (
	"fmt"
	"strings"

	"github.com/nextunit-io/go-pdf2X/pdf2text"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	client, err := pdf2text.NewClient()
	checkErr(err)

	v, err := client.GetVersion()
	checkErr(err)
	fmt.Printf("Client is using version %s\n\n", *v)

	encs, err := client.GetEncodings()
	checkErr(err)
	fmt.Printf("The encodings are available: %s\n\n", strings.Join(encs, ", "))

	pdf, err := client.Get("test/Test_PDF.pdf", pdf2text.Options{Layout: true})
	checkErr(err)
	fmt.Println("Output of the PDF:")
	fmt.Println(*pdf)
}

```

This example will print upon the command `go run main.go` for example:
```text
Client is using version 24.11.0

The encodings are available: ASCII7, Big5, Big5ascii, EUC-CN, EUC-JP, GBK, ISO-2022-CN, ISO-2022-JP, ISO-2022-KR, ISO-8859-6, ISO-8859-7, ISO-8859-8, ISO-8859-9, KOI8-R, Latin1, Latin2, Shift-JIS, Symbol, TIS-620, UTF-16, UTF-8, Windows-1255, ZapfDingbats

Output of the PDF:
Test PDF
```

It might look different because of different installations on every computer.

### Get options

When it comes to the use of the get function, it is required to provide the file and options for processing the file. Therefor here is a short overview what kind of options are available:

```go
type Options struct {
	FirstPage     *int     // first page to convert
	LastPage      *int     // last page to convert
	Resolution    *int     // resolution, in DPI (default is 72)
	X             *int     // x-coordinate of the crop area top left corner
	Y             *int     // y-coordinate of the crop area top left corner
	Width         *int     // width of crop area in pixels (default is 0)
	Height        *int     // height of crop area in pixels (default is 0)
	Layout        bool     // maintain original physical layout
	Fixed         *string  // assume fixed-pitch (or tabular) text
	Raw           bool     // keep strings in content stream order
	NoDiag        bool     // discard diagonal text
	HtmlMeta      bool     // generate a simple HTML file, including the meta information
	Tsv           bool     // generate a simple TSV file, including the meta information for bounding boxes
	Enc           *string  // output text encoding name
	Eol           *string  // output end-of-line convention (unix, dos, or mac)
	Nopgbrk       bool     // don't insert page breaks between pages
	Bbox          bool     // output bounding box for each word and page size to html. Sets -htmlmeta
	BboxLayout    bool     // like -bbox but with extra layout bounding box data.  Sets -htmlmeta
	CropBox       bool     // use the crop box rather than media box
	ColSpacing    *float32 // how much spacing we allow after a word before considering adjacent text to be a new column, as a fraction of the font size (default is 0.7, old releases had a 0.3 default)
	OwnerPassword *string  // owner password (for encrypted files)
	UserPassword  *string  // user password (for encrypted files)
}
```
## pdf2html

Lib to abstract the pdftohtml cli library

### Preconditions

For this library it is necessary that `pdftohtml` is installed in the version above `24.11.x - 25.x.x`. 
With homebrew it is possible to install it via `brew install poppler` [Homebrew Poppler](https://formulae.brew.sh/formula/poppler).

### Usage

This is a module for an easy usage in GO. It abstracts the CLI commands in go methods that can be used for further development.

```go
package main

import (
	"fmt"
	"strings"

	"github.com/nextunit-io/go-pdf2X/pdf2html"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	client, err := pdf2html.NewClient()
	checkErr(err)

	v, err := client.GetVersion()
	checkErr(err)
	fmt.Printf("Client is using version %s\n\n", *v)

	pdf, err := client.GetXML("test/Test_PDF.pdf", pdf2html.Options{
		Xml: true,
	})
	checkErr(err)
	fmt.Println("Output of the PDF as XML:")
	fmt.Println(*pdf)

	pdf, err = client.GetHTML("test/Test_PDF.pdf", pdf2html.Options{})
	checkErr(err)
	fmt.Println("Output of the PDF as HTML:")
	fmt.Println(*pdf)
}

```

This example will print upon the command `go run main.go` for example:
```text
Client is using version 24.11.0

Output of the PDF as XML:
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE pdf2xml SYSTEM "pdf2xml.dtd">

<pdf2xml producer="poppler" version="24.11.0">
<page number="1" position="absolute" top="0" left="0" height="1263" width="892">
        <fontspec id="0" size="21" family="IMPLXZ+Calibri" color="#000000"/>
<text top="106" left="106" width="79" height="25" font="0">Test PDF </text>
</page>
</pdf2xml>

Output of the PDF as HTML:
<!DOCTYPE html>
<html>
<head>
<title>Microsoft Word - Dokument1</title>
<meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>
<meta name="generator" content="pdftohtml 0.36"/>
</head>
<frameset cols="100,*">
<frame name="links" src="test_Test_PDF.pdf-2084797302_ind.html"/>
<frame name="contents" src="test_Test_PDF.pdf-2084797302s.html"/>
</frameset>
</html>
```

It might look different because of different installations on every computer.

### Get options

When it comes to the use of the get function, it is required to provide the file and options for processing the file. Therefor here is a short overview what kind of options are available:

```go
type Options struct {
	FirstPage     *int     // first page to convert
	LastPage      *int     // last page to convert
	Quite         bool     // don't print any messages or errors
	ExchangeLinks bool     // exchange .pdf links by .html
	ComplexDoc    bool     // generate complex document
	SingleDoc     bool     // generate single document that includes all pages
	IgnoreImages  bool     // ignore images
	NoFrames      bool     // generate no frames
	Stdout        bool     // use standard output
	Zoom          *float32 // zoom the pdf document (default 1.5)
	Xml           bool     // output for XML post-processing
	NoRoundCoord  bool     // do not round coordinates (with XML output only)
	Hidden        bool     // output hidden text
	NoMerge       bool     // do not merge paragraphs
	Enc           *string  // output text encoding name
	Fmt           *string  // image file format for Splash output (png or jpg)
	OwnerPassword *string  // owner password (for encrypted files)
	UserPassword  *string  // user password (for encrypted files)
	NoDrm         bool     // override document DRM settings
	Wbt           *int     // word break threshold (default 10 percent)
	FontFullName  bool     // outputs font full name
}
```