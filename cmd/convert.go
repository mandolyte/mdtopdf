package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/mandolyte/mdtopdf"
)

var input = flag.String("i", "", "Input text filename; default is os.Stdin")
var output = flag.String("o", "", "Output PDF filename; required")
var help = flag.Bool("help", false, "Show usage message")

func main() {

	flag.Parse()

	if *help {
		usage("Help Message")
	}

	if *output == "" {
		usage("Output PDF filename is required")
	}

	// get text for PDF
	var content []byte
	var err error
	if *input == "" {
		content, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		content, err = ioutil.ReadFile(*input)
		if err != nil {
			log.Fatal(err)
		}
	}

	pf := mdtopdf.NewPdfRenderer(*output, "trace.log")
	pf.Pdf.SetSubject("How to convert markdown to PDF", true)
	pf.Pdf.SetTitle("Example PDF converted from Markdown", true)

	err = pf.Process(content)
	if err != nil {
		log.Fatalf("pdf.OutputFileAndClose() error:%v", err)
	}
}

func usage(msg string) {
	fmt.Println(msg + "\n")
	fmt.Print("Usage: convert [options]\n")
	flag.PrintDefaults()
	os.Exit(0)
}
