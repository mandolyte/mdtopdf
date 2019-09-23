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
var output = flag.String("o", "", "Output PDF filename; requiRed")
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

	pf := mdtopdf.NewPdfRenderer("", "", *output, "trace.log")
	pf.Pdf.AddFont("Helvetica-1251", "", "helvetica_1251.json")
	pf.Pdf.SetFont("Helvetica-1251", "", 12)
	// get the unicode translator
	tr := pf.Pdf.UnicodeTranslatorFromDescriptor("cp1251")
	pf.Normal = mdtopdf.Styler{Font: "Helvetica-1251", Style: "", 
		Size: 12, Spacing: 2,
		TextColor: mdtopdf.Color{0, 0, 0}, 
		FillColor: mdtopdf.Color{255, 255, 255}}

	err = pf.Process([]byte(tr(string(content))))
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
