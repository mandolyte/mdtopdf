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

// Package mdtopdf converts markdown to PDF.
package mdtopdf

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jung-kurt/gofpdf"
	bf "github.com/russross/blackfriday/v2"
)

// Color is a RGB set of ints; for a nice picker
// see https://www.w3schools.com/colors/colors_picker.asp
type Color struct {
	Red, Green, Blue int
}

// Styler is the struct to capture the styling features for text
// Size and Spacing are specified in points.
// The sum of Size and Spacing is used as line height value
// in the gofpdf API
type Styler struct {
	Font      string
	Style     string
	Size      float64
	Spacing   float64
	TextColor Color
	FillColor Color
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
	Normal Styler
	em     float64

	// link text
	Link Styler

	// backticked text
	Backtick Styler

	// blockquote text
	Blockquote  Styler
	IndentValue float64

	// Headings
	H1 Styler
	H2 Styler
	H3 Styler
	H4 Styler
	H5 Styler
	H6 Styler

	// Table styling
	THeader Styler
	TBody   Styler

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
	r.Normal = Styler{Font: "Arial", Style: "", Size: 12, Spacing: 2,
		TextColor: Color{0, 0, 0}, FillColor: Color{255, 255, 255}}

	// Link text
	r.Link = Styler{Font: "Arial", Style: "iu", Size: 12, Spacing: 2,
		TextColor: Color{0, 0, 0}, FillColor: Color{255, 255, 255}}

	// Backticked text
	r.Backtick = Styler{Font: "Courier", Style: "", Size: 12, Spacing: 2,
		TextColor: Color{37, 27, 14}, FillColor: Color{200, 200, 200}}

	// Headings
	r.H1 = Styler{Font: "Arial", Style: "b", Size: 24, Spacing: 5,
		TextColor: Color{0, 0, 0}, FillColor: Color{255, 255, 255}}
	r.H2 = Styler{Font: "Arial", Style: "b", Size: 22, Spacing: 5,
		TextColor: Color{0, 0, 0}, FillColor: Color{255, 255, 255}}
	r.H3 = Styler{Font: "Arial", Style: "b", Size: 20, Spacing: 5,
		TextColor: Color{0, 0, 0}, FillColor: Color{255, 255, 255}}
	r.H4 = Styler{Font: "Arial", Style: "b", Size: 18, Spacing: 5,
		TextColor: Color{0, 0, 0}, FillColor: Color{255, 255, 255}}
	r.H5 = Styler{Font: "Arial", Style: "b", Size: 16, Spacing: 5,
		TextColor: Color{0, 0, 0}, FillColor: Color{255, 255, 255}}
	r.H6 = Styler{Font: "Arial", Style: "b", Size: 14, Spacing: 5,
		TextColor: Color{0, 0, 0}, FillColor: Color{255, 255, 255}}

	//r.inBlockquote = false
	//r.inHeading = false
	r.Blockquote = Styler{Font: "Arial", Style: "i", Size: 12, Spacing: 2,
		TextColor: Color{0, 0, 0}, FillColor: Color{255, 255, 255}}

	// Table Header Text
	r.THeader = Styler{Font: "Arial", Style: "B", Size: 12, Spacing: 2,
		TextColor: Color{0, 0, 0}, FillColor: Color{180, 180, 180}}

	// Table Body Text
	r.TBody = Styler{Font: "Arial", Style: "", Size: 12, Spacing: 2,
		TextColor: Color{0, 0, 0}, FillColor: Color{240, 240, 240}}

	r.Pdf = gofpdf.New(r.orientation, r.units, r.papersize, r.fontdir)
	r.Pdf.AddPage()
	// set default font
	r.setStyler(r.Normal)
	r.mleft, r.mtop, r.mright, r.mbottom = r.Pdf.GetMargins()
	r.em = r.Pdf.GetStringWidth("m")
	r.IndentValue = 3 * r.em

	//r.current = r.normal // set default
	r.cs = states{stack: make([]*containerState, 0)}
	initcurrent := &containerState{containerType: bf.Paragraph,
		listkind:  notlist,
		textStyle: r.Normal, leftMargin: r.mleft}
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

func (r *PdfRenderer) setStyler(s Styler) {
	r.Pdf.SetFont(s.Font, s.Style, s.Size)
	r.Pdf.SetTextColor(s.TextColor.Red, s.TextColor.Green, s.TextColor.Blue)
	r.Pdf.SetFillColor(s.FillColor.Red, s.FillColor.Green, s.FillColor.Blue)
}

func (r *PdfRenderer) write(s Styler, t string) {
	r.Pdf.Write(s.Size+s.Spacing, t)
}

func (r *PdfRenderer) multiCell(s Styler, t string) {
	r.Pdf.MultiCell(0, s.Size+s.Spacing, t, "", "", true)
}

func (r *PdfRenderer) writeLink(s Styler, display, url string) {
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
		r.processText(node)
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
		r.processTable(node, entering)
	case bf.TableHead:
		r.processTableHead(node, entering)
	case bf.TableBody:
		r.processTableBody(node, entering)
	case bf.TableRow:
		r.processTableRow(node, entering)
	case bf.TableCell:
		r.processTableCell(node, entering)
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
	LH := r.cs.peek().textStyle.Size + r.cs.peek().textStyle.Spacing
	r.tracer("cr()", fmt.Sprintf("LH=%v", LH))
	r.write(r.cs.peek().textStyle, "\n")
	//r.Pdf.Ln(-1)
}

// Tracer traces parse and pdf generation activity.
func (r *PdfRenderer) tracer(source, msg string) {
	if r.tracerFile != "" {
		indent := strings.Repeat("-", len(r.cs.stack)-1)
		r.w.WriteString(fmt.Sprintf("%v[%v] %v\n", indent, source, msg))
	}
}
