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
