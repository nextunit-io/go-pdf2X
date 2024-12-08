package main

import (
	"fmt"
	"strings"

	"github.com/nextunit-io/go-pdf2X/pdf2html"
	"github.com/nextunit-io/go-pdf2X/pdf2text"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	runPdf2Text()
	runPdf2HTML()
}

func runPdf2Text() {
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

func runPdf2HTML() {
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
