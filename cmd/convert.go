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
var pathToSyntaxFiles = flag.String("s", "", "Path to github.com/jessp01/gohighlight/syntax_files")
var title = flag.String("title", "", "Presentation title")
var author = flag.String("author", "", "Author; used if -footer is passed")
var unicodeSupport = flag.String("unicode-encoding", "", "e.g 'cp1251'")
var fontFile = flag.String("font-file", "", "path to font file to use")
var fontName = flag.String("font-name", "", "Font name ID; e.g 'Helvetica-1251'")
var themeArg = flag.String("theme", "light", "[light|dark]")
var hrAsNewPage = flag.Bool("new-page-on-hr", false, "Interpret HR as a new page; useful for presentations")
var printFooter = flag.Bool("with-footer", false, "Print doc footer (author  title  page number)")
var pageSize = flag.String("page-size", "A4", "[A3 | A4 | A5]")
var help = flag.Bool("help", false, "Show usage message")

var opts []mdtopdf.RenderOption

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	if *help {
		usage("Help Message")
	}

	if *output == "" {
		usage("Output PDF filename is required")
	}

	if *hrAsNewPage == true {
		opts = append(opts, mdtopdf.IsHorizontalRuleNewPage(true))
	}

	if *unicodeSupport != "" {
		opts = append(opts, mdtopdf.WithUnicodeTranslator(*unicodeSupport))
	}

	if *pathToSyntaxFiles != "" {
		opts = append(opts, mdtopdf.SetSyntaxHighlightBaseDir(*pathToSyntaxFiles))
	} else {
		opts = append(opts, mdtopdf.SetSyntaxHighlightBaseDir("../highlight/syntax_files"))
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

	theme := mdtopdf.LIGHT
	textColor := mdtopdf.Colorlookup("black")
	fillColor := mdtopdf.Colorlookup("white")
	backgroundColor := "white"
	if *themeArg == "dark" {
	    theme = mdtopdf.DARK
	    backgroundColor = "eerieblack"
	    textColor = mdtopdf.Colorlookup("darkgray")
	    fillColor = mdtopdf.Colorlookup("black")
	}

	pf := mdtopdf.NewPdfRenderer("", *pageSize, *output, "trace.log", opts, theme)
	pf.Pdf.SetSubject(*title, true)
	pf.Pdf.SetTitle(*title, true)
	pf.BackgroundColor = mdtopdf.Colorlookup(backgroundColor)

	if *fontFile != "" && *fontName != "" {
		pf.Pdf.AddFont(*fontName, "", *fontFile)
		pf.Pdf.SetFont(*fontName, "", 12)
		pf.Normal = mdtopdf.Styler{Font: *fontName, Style: "",
			Size: 12, Spacing: 2,
			FillColor: fillColor,
			TextColor: textColor}
	}

	if *printFooter {
		pf.Pdf.SetFooterFunc(func() {
			color := mdtopdf.Colorlookup(backgroundColor)
			pf.Pdf.SetFillColor(color.Red, color.Green, color.Blue)
			// Position at 1.5 cm from bottom
			pf.Pdf.SetY(-15)
			// Arial italic 8
			pf.Pdf.SetFont("Arial", "I", 8)
			// Text color in gray
			pf.Pdf.SetTextColor(128, 128, 128)
			w, _, _ := pf.Pdf.PageSize(pf.Pdf.PageNo())
			//fmt.Printf("Width: %d, height: %d, unit: %s\n", w, h, u)
			pf.Pdf.SetX(4)
			pf.Pdf.CellFormat(0, 10, fmt.Sprintf("%s", *author), "", 0, "", true, 0, "")
			pf.Pdf.SetX(w/2 - float64(len(*title)))
			pf.Pdf.CellFormat(0, 10, fmt.Sprintf("%s", *title), "", 0, "", true, 0, "")
			pf.Pdf.SetX(-40)
			pf.Pdf.CellFormat(0, 10, fmt.Sprintf("Page %d", pf.Pdf.PageNo()), "", 0, "", true, 0, "")
		})
	}

	err = pf.Process(content)
	if err != nil {
		log.Fatalf("error:%v", err)
	}
}

func usage(msg string) {
	fmt.Println(msg + "\n")
	fmt.Print("Usage: convert [options]\n")
	flag.PrintDefaults()
	os.Exit(0)
}
