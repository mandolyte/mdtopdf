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

// Styler is the struct to capture the styling features for text
type Styler struct {
	Font    string
	Style   string
	Size    float64
	Spacing float64
}

// PdfRenderer is the struct to manage conversion of a markdown object
// to PDF format.
type PdfRenderer struct {
	Pdf                *gofpdf.Fpdf
	Orientation, units string
	Papersize, fontdir string

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

	// headings
	H1 Styler
	H2 Styler
	H3 Styler
	H4 Styler
	H5 Styler
	H6 Styler

	cs states
}

// NewPdfRenderer creates and configures an PdfRenderer object,
// which satisfies the Renderer interface.
func NewPdfRenderer(pdfFile, tracerFile string) *PdfRenderer {

	r := new(PdfRenderer)

	// set filenames
	r.pdfFile = pdfFile
	r.tracerFile = tracerFile

	// Global things
	r.Orientation = "portrait"
	r.units = "pt"
	r.Papersize = "A4"
	r.fontdir = "."

	// Normal Text
	r.Normal = Styler{Font: "Arial", Style: "", Size: 12, Spacing: 2}

	// Link text
	r.Link = Styler{Font: "Arial", Style: "iu", Size: 12, Spacing: 2}

	// Backticked text
	r.Backtick = Styler{Font: "Courier", Style: "", Size: 12, Spacing: 2}

	// Headings
	r.H1 = Styler{Font: "Arial", Style: "b", Size: 24, Spacing: 5}
	r.H2 = Styler{Font: "Arial", Style: "b", Size: 22, Spacing: 5}
	r.H3 = Styler{Font: "Arial", Style: "b", Size: 20, Spacing: 5}
	r.H4 = Styler{Font: "Arial", Style: "b", Size: 18, Spacing: 5}
	r.H5 = Styler{Font: "Arial", Style: "b", Size: 16, Spacing: 5}
	r.H6 = Styler{Font: "Arial", Style: "b", Size: 14, Spacing: 5}

	//r.inBlockquote = false
	//r.inHeading = false
	r.Blockquote = Styler{Font: "Arial", Style: "i", Size: 12, Spacing: 2}

	r.Pdf = gofpdf.New(r.Orientation, r.units, r.Papersize, r.fontdir)
	r.Pdf.AddPage()
	// set default font
	r.setFont(r.Normal)
	r.mleft, r.mtop, r.mright, r.mbottom = r.Pdf.GetMargins()
	r.em = r.Pdf.GetStringWidth("m")
	r.IndentValue = 3 * r.em

	//r.current = r.Normal // set default
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

func (r *PdfRenderer) setFont(s Styler) {
	r.Pdf.SetFont(s.Font, s.Style, s.Size)
}

func (r *PdfRenderer) write(s Styler, t string) {
	r.Pdf.Write(s.Size+s.Spacing, t)
}

func (r *PdfRenderer) multiCell(s Styler, t string) {
	r.Pdf.MultiCell(0, s.Size+s.Spacing, t, "", "", false)
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
		//r.setFont(r.current)
		currentStyle := r.cs.peek().textStyle
		r.setFont(currentStyle)
		s := string(node.Literal)
		s = strings.Replace(s, "\n", " ", -1)
		r.Tracer("Text", s)

		if r.cs.peek().containerType == bf.Link {
			//r.writeLink(currentStyle, string(node.Literal), r.cs.peek().destination)
			r.writeLink(currentStyle, s, r.cs.peek().destination)
		} else {
			r.write(currentStyle, s)
		}

	case bf.Softbreak:
		r.Tracer("Softbreak", "Output newline")
		r.cr()
	case bf.Hardbreak:
		r.Tracer("Hardbreak", "Output newline")
		r.cr()
	case bf.Emph:
		r.processEmph(node, entering)
	case bf.Strong:
		r.processStrong(node, entering)
	case bf.Del:
		if entering {
			r.Tracer("DEL (entering)", "Not handled")
		} else {
			r.Tracer("DEL (leaving)", "Not handled")
		}
	case bf.HTMLSpan:
		r.Tracer("HTMLSpan", "Not handled")
	case bf.Link:
		r.processLink(node, entering)
	case bf.Image:
		r.processImage(node, entering)
	case bf.Code:
		r.processCode(node)
	case bf.Document:
		r.Tracer("Document", "Not Handled")
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
			r.Tracer("Table (entering)", "Not handled")
		} else {
			r.Tracer("Table (leaving)", "Not handled")
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
			r.Tracer("TableCell (entering)", "Not handled")
		} else {
			r.Tracer("TableCell (leaving)", "Not handled")
		}
	case bf.TableHead:
		if entering {
			r.Tracer("TableHead (entering)", "Not handled")
		} else {
			r.Tracer("TableHead (leaving)", "Not handled")
		}
	case bf.TableBody:
		if entering {
			r.Tracer("TableBody (entering)", "Not handled")
		} else {
			r.Tracer("TableBody (leaving)", "Not handled")
		}
	case bf.TableRow:
		if entering {
			r.Tracer("TableRow (entering)", "Not handled")
		} else {
			r.Tracer("TableRow (leaving)", "Not handled")
		}
	default:
		panic("Unknown node type " + node.Type.String())
	}
	return bf.GoToNext
}

// RenderHeader writes HTML document preamble and TOC if requested.
func (r *PdfRenderer) RenderHeader(w io.Writer, ast *bf.Node) {
	r.Tracer("RenderHeader", "Not handled")
}

// RenderFooter writes HTML document footer.
func (r *PdfRenderer) RenderFooter(w io.Writer, ast *bf.Node) {
	r.Tracer("RenderFooter", "Not handled")
}

func (r *PdfRenderer) cr() {
	//r.Tracer("fpdf.Ln()", fmt.Sprintf("LH=%v", r.current.Size+r.current.Spacing))
	//r.Pdf.Ln(r.current.Size + r.current.Spacing)
	LH := r.cs.peek().textStyle.Size + r.cs.peek().textStyle.Spacing
	r.Tracer("cr()", fmt.Sprintf("LH=%v", LH))
	r.write(r.cs.peek().textStyle, "\n")
}

// Tracer traces parse and pdf generation activity.
// Output goes to Stdout when DebugMode value is set to true
func (r *PdfRenderer) Tracer(source, msg string) {
	if r.tracerFile != "" {
		indent := strings.Repeat("-", len(r.cs.stack)-1)
		r.w.WriteString(fmt.Sprintf("%v[%v] %v\n", indent, source, msg))
	}
}
