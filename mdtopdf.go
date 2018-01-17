/*
 * Markdown to PDF Converter
 * Available at http://github.com/mandolyte/mdtopdf
 *
 * Copyright © 2018 Cecil New <cecil.new@gmail.com>.
 * Distributed under the MIT License.
 * See README.md for details.
 *
 * Dependencies
 * This package depends on two other packages:
 *
 * Blackfriday Markdown Processor
 *   Available at http://github.com/russross/blackfriday
 *
 * gofpdf - a PDF document generator with high level support for
 *   text, drawing and images.
 *   Available at https://github.com/jung-kurt/gofpdf
 */
/*
Package mdtopdf implements a PDF document generator for markdown documents.

Introduction

This package depends on two other packages:
* The BlackFriday v2 parser to read the markdown source
* The `gofpdf` packace to generate the PDF

The tests included here are from the BlackFriday package.
See the "testdata" folder.
The tests create PDF files and thus while the tests may complete
without errors, visual inspection of the created PDF is the
only way to determine if the tests *really* pass!

The tests create log files that trace the BlackFriday parser
callbacks. This is a valuable debug tool showing each callback
and data provided in each while the AST is presented.

Installation

To install the package, run the usual `go get`:

	go get github.com/mandolyte/mdtopdf


Quick start

In the `cmd` folder is an example using the package. It demonstrates
a number of features. The test PDF was created with this command:

	go run convert.go -i test.md -o test.pdf

*/

package mdtopdf

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jung-kurt/gofpdf"
	bf "gopkg.in/russross/blackfriday.v2"
)

// styler is the struct to capture the styling features for text
// Size and Spacing are specified in points.
// The sum of Size and Spacing is used as line height value
// in the gofpdf API
type styler struct {
	Font    string
	Style   string
	Size    float64
	Spacing float64
}

// PdfRenderer is the struct to manage conversion of a markdown object
// to PDF format.
type PdfRenderer struct {
	// Pdf can be used to access the underlying created gofpdf object
	// prior to processing the markdown source
	Pdf                *gofpdf.Fpdf
	orientation, units string
	papersize, fontdir string

	// trace/log file if present
	pdfFile, tracerFile string
	w                   *bufio.Writer

	// default margins for safe keeping
	mleft, mtop, mright, mbottom float64

	// normal text
	normal styler
	em     float64

	// link text
	link styler

	// backticked text
	backtick styler

	// blockquote text
	blockquote  styler
	indentValue float64

	// headings
	h1 styler
	h2 styler
	h3 styler
	h4 styler
	h5 styler
	h6 styler

	cs states
}

// NewPdfRenderer creates and configures an PdfRenderer object,
// which satisfies the Renderer interface.
func NewPdfRenderer(orient, papersz, pdfFile, tracerFile string) *PdfRenderer {

	r := new(PdfRenderer)

	// set filenames
	r.pdfFile = pdfFile
	r.tracerFile = tracerFile

	// Global things
	r.orientation = "portrait"
	if orient != "" {
		r.orientation = orient
	}
	r.units = "pt"
	r.papersize = "Letter"
	if papersz != "" {
		r.papersize = papersz
	}

	r.fontdir = "."

	// Normal Text
	r.normal = styler{Font: "Arial", Style: "", Size: 12, Spacing: 2}

	// Link text
	r.link = styler{Font: "Arial", Style: "iu", Size: 12, Spacing: 2}

	// Backticked text
	r.backtick = styler{Font: "Courier", Style: "", Size: 12, Spacing: 2}

	// Headings
	r.h1 = styler{Font: "Arial", Style: "b", Size: 24, Spacing: 5}
	r.h2 = styler{Font: "Arial", Style: "b", Size: 22, Spacing: 5}
	r.h3 = styler{Font: "Arial", Style: "b", Size: 20, Spacing: 5}
	r.h4 = styler{Font: "Arial", Style: "b", Size: 18, Spacing: 5}
	r.h5 = styler{Font: "Arial", Style: "b", Size: 16, Spacing: 5}
	r.h6 = styler{Font: "Arial", Style: "b", Size: 14, Spacing: 5}

	//r.inBlockquote = false
	//r.inHeading = false
	r.blockquote = styler{Font: "Arial", Style: "i", Size: 12, Spacing: 2}

	r.Pdf = gofpdf.New(r.orientation, r.units, r.papersize, r.fontdir)
	r.Pdf.AddPage()
	// set default font
	r.setFont(r.normal)
	r.mleft, r.mtop, r.mright, r.mbottom = r.Pdf.GetMargins()
	r.em = r.Pdf.GetStringWidth("m")
	r.indentValue = 3 * r.em

	//r.current = r.normal // set default
	r.cs = states{stack: make([]*containerState, 0)}
	initcurrent := &containerState{containerType: bf.Paragraph,
		listkind:  notlist,
		textStyle: r.normal, leftMargin: r.mleft}
	r.cs.push(initcurrent)
	return r
}

// Process takes the markdown content, parses it to generate the PDF
func (r *PdfRenderer) Process(content []byte) error {

	// try to open tracer
	var f *os.File
	var err error
	if r.tracerFile != "" {
		f, err = os.Create(r.tracerFile)
		if err != nil {
			return fmt.Errorf("os.Create() on tracefile error:%v", err)
		}
		defer f.Close()
		r.w = bufio.NewWriter(f)
		defer r.w.Flush()
	}

	// Preprocess content by changing all CRLF to LF
	s := string(content)
	s = strings.Replace(s, "\r\n", "\n", -1)

	content = []byte(s)
	_ = bf.Run(content, bf.WithRenderer(r))

	err = r.Pdf.OutputFileAndClose(r.pdfFile)
	if err != nil {
		return fmt.Errorf("Pdf.OutputFileAndClose() error on %v:%v", r.pdfFile, err)
	}
	return nil
}

func (r *PdfRenderer) setFont(s styler) {
	r.Pdf.SetFont(s.Font, s.Style, s.Size)
}

func (r *PdfRenderer) write(s styler, t string) {
	r.Pdf.Write(s.Size+s.Spacing, t)
}

func (r *PdfRenderer) multiCell(s styler, t string) {
	r.Pdf.MultiCell(0, s.Size+s.Spacing, t, "", "", false)
}

func (r *PdfRenderer) writeLink(s styler, display, url string) {
	r.Pdf.WriteLinkString(s.Size+s.Spacing, display, url)
}

// RenderNode is a default renderer of a single node of a syntax tree. For
// block nodes it will be called twice: first time with entering=true, second
// time with entering=false, so that it could know when it's working on an open
// tag and when on close. It writes the result to w.
//
// The return value is a way to tell the calling walker to adjust its walk
// pattern: e.g. it can terminate the traversal by returning Terminate. Or it
// can ask the walker to skip a subtree of this node by returning SkipChildren.
// The typical behavior is to return GoToNext, which asks for the usual
// traversal to the next node.
// (above taken verbatim from the blackfriday v2 package)
func (r *PdfRenderer) RenderNode(w io.Writer, node *bf.Node, entering bool) bf.WalkStatus {
	switch node.Type {
	case bf.Text:
		//r.setFont(r.current)
		currentStyle := r.cs.peek().textStyle
		r.setFont(currentStyle)
		s := string(node.Literal)
		s = strings.Replace(s, "\n", " ", -1)
		r.tracer("Text", s)

		if r.cs.peek().containerType == bf.Link {
			//r.writeLink(currentStyle, string(node.Literal), r.cs.peek().destination)
			r.writeLink(currentStyle, s, r.cs.peek().destination)
		} else {
			r.write(currentStyle, s)
		}

	case bf.Softbreak:
		r.tracer("Softbreak", "Output newline")
		r.cr()
	case bf.Hardbreak:
		r.tracer("Hardbreak", "Output newline")
		r.cr()
	case bf.Emph:
		r.processEmph(node, entering)
	case bf.Strong:
		r.processStrong(node, entering)
	case bf.Del:
		if entering {
			r.tracer("DEL (entering)", "Not handled")
		} else {
			r.tracer("DEL (leaving)", "Not handled")
		}
	case bf.HTMLSpan:
		r.tracer("HTMLSpan", "Not handled")
	case bf.Link:
		r.processLink(node, entering)
	case bf.Image:
		r.processImage(node, entering)
	case bf.Code:
		r.processCode(node)
	case bf.Document:
		r.tracer("Document", "Not Handled")
	case bf.Paragraph:
		r.processParagraph(node, entering)
	case bf.BlockQuote:
		r.processBlockQuote(node, entering)
	case bf.HTMLBlock:
		r.processHTMLBlock(node)
	case bf.Heading:
		r.processHeading(node, entering)
	case bf.HorizontalRule:
		r.processHorizontalRule(node)
	case bf.List:
		r.processList(node, entering)
	case bf.Item:
		r.processItem(node, entering)
	case bf.CodeBlock:
		r.processCodeblock(node)
	case bf.Table:
		if entering {
			r.tracer("Table (entering)", "Not handled")
		} else {
			r.tracer("Table (leaving)", "Not handled")
		}
	case bf.TableCell:
		/*
			openTag := tdTag
			closeTag := tdCloseTag
			if node.IsHeader {
				openTag = thTag
				closeTag = thCloseTag
			}
		*/
		if entering {
			r.tracer("TableCell (entering)", "Not handled")
		} else {
			r.tracer("TableCell (leaving)", "Not handled")
		}
	case bf.TableHead:
		if entering {
			r.tracer("TableHead (entering)", "Not handled")
		} else {
			r.tracer("TableHead (leaving)", "Not handled")
		}
	case bf.TableBody:
		if entering {
			r.tracer("TableBody (entering)", "Not handled")
		} else {
			r.tracer("TableBody (leaving)", "Not handled")
		}
	case bf.TableRow:
		if entering {
			r.tracer("TableRow (entering)", "Not handled")
		} else {
			r.tracer("TableRow (leaving)", "Not handled")
		}
	default:
		panic("Unknown node type " + node.Type.String())
	}
	return bf.GoToNext
}

// RenderHeader is not supported.
func (r *PdfRenderer) RenderHeader(w io.Writer, ast *bf.Node) {
	r.tracer("RenderHeader", "Not handled")
}

// RenderFooter is not supported.
func (r *PdfRenderer) RenderFooter(w io.Writer, ast *bf.Node) {
	r.tracer("RenderFooter", "Not handled")
}

func (r *PdfRenderer) cr() {
	//r.tracer("fpdf.Ln()", fmt.Sprintf("LH=%v", r.current.Size+r.current.Spacing))
	//r.Pdf.Ln(r.current.Size + r.current.Spacing)
	LH := r.cs.peek().textStyle.Size + r.cs.peek().textStyle.Spacing
	r.tracer("cr()", fmt.Sprintf("LH=%v", LH))
	r.write(r.cs.peek().textStyle, "\n")
}

// Tracer traces parse and pdf generation activity.
// Output goes to Stdout when DebugMode value is set to true
func (r *PdfRenderer) tracer(source, msg string) {
	if r.tracerFile != "" {
		indent := strings.Repeat("-", len(r.cs.stack)-1)
		r.w.WriteString(fmt.Sprintf("%v[%v] %v\n", indent, source, msg))
	}
}
