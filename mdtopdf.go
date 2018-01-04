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

	// current settings
	current Styler

	// normal text
	Normal Styler

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

	// state booleans
	inBlockquote      bool
	inHeading         bool
	inUnorderedItem   bool
	inOrderedItem     bool
	inDefinitionItem  bool
	currentItemNumber int
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
	r.current = r.Normal // set default

	// Backticked text
	r.Backtick = Styler{Font: "Courier", Style: "", Size: 12, Spacing: 2}

	// Headings
	r.H1 = Styler{Font: "Arial", Style: "b", Size: 24, Spacing: 5}
	r.H2 = Styler{Font: "Arial", Style: "b", Size: 22, Spacing: 5}
	r.H3 = Styler{Font: "Arial", Style: "b", Size: 20, Spacing: 5}
	r.H4 = Styler{Font: "Arial", Style: "b", Size: 18, Spacing: 5}
	r.H5 = Styler{Font: "Arial", Style: "b", Size: 16, Spacing: 5}
	r.H6 = Styler{Font: "Arial", Style: "b", Size: 14, Spacing: 5}

	r.inBlockquote = false
	r.inHeading = false
	r.IndentValue = 36
	r.Blockquote = Styler{Font: "Arial", Style: "i", Size: 12, Spacing: 2}

	r.Pdf = gofpdf.New(r.Orientation, r.units,
		r.Papersize, r.fontdir)
	r.Pdf.AddPage()
	// set default font
	r.setFont(r.Normal)
	r.mleft, r.mtop, r.mright, r.mbottom = r.Pdf.GetMargins()
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

	// just in case, change all CRLF to LF
	content = []byte(strings.Replace(string(content), "\r\n", "\n", -1))
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
		r.setFont(r.current)
		s := string(node.Literal)
		s = strings.Replace(s, "\n", " ", -1)
		r.Tracer("Text", s)
		if r.inUnorderedItem {
			r.write(r.current, "- ")
		} else if r.inOrderedItem {
			r.write(r.current, fmt.Sprintf("%v. ", r.currentItemNumber))
		}
		r.write(r.current, s)
		/*
			if r.inHeading || r.inUnorderedItem || r.inOrderedItem || r.inDefinitionItem {
				r.write(r.current, "\n") // output a newline with heading LH size
			}
		*/
	case bf.Softbreak:
		r.Tracer("Softbreak", "Not handled")
	case bf.Hardbreak:
		r.Tracer("Hardbreak", "Not handled")
	case bf.Emph:
		if entering {
			r.Tracer("Emph (entering)", "")
			r.current.Style += "i"
		} else {
			r.Tracer("Emph (leaving)", "")
			r.current.Style = strings.Replace(r.current.Style, "i", "", -1)
		}
	case bf.Strong:
		if entering {
			r.Tracer("Strong (entering)", "")
			r.current.Style += "b"
		} else {
			r.Tracer("Strong (leaving)", "")
			r.current.Style = strings.Replace(r.current.Style, "b", "", -1)
		}
	case bf.Del:
		if entering {
			r.Tracer("DEL (entering)", "Not handled")
		} else {
			r.Tracer("DEL (leaving)", "Not handled")
		}
	case bf.HTMLSpan:
		r.Tracer("HTMLSpan", "Not handled")
	case bf.Link:
		// mark it but don't link it if it is not a safe link: no smartypants
		//dest := node.LinkData.Destination
		if entering {
			r.Tracer("Link (entering)", "Not handled")
		} else {
			r.Tracer("Link (leaving)", "Not handled")
		}
	case bf.Image:
		if entering {
			r.Tracer("Image (entering)", "Not handled")
		} else {
			r.Tracer("Image (leaving)", "Not handled")
		}
	case bf.Code:
		r.Tracer("Code", "")
		r.setFont(r.Backtick)
		r.write(r.Backtick, string(node.Literal))
	case bf.Document:
		r.Tracer("Document", "Not Handled")
		//break
	case bf.Paragraph:
		if entering {
			r.Tracer("Paragraph (entering)", "")
			if r.inUnorderedItem || r.inOrderedItem || r.inDefinitionItem {
				r.Tracer("Para within a list", "breaking")
				break
			}
			if r.inBlockquote {
				// no change to styler
			} else {
				r.current = r.Normal
			}
			//r.cr()
		} else {
			r.Tracer("Paragraph (leaving)", "")
			if r.inUnorderedItem || r.inOrderedItem || r.inDefinitionItem {
				r.Tracer("Para within a list", "breaking")
				break
			}
			r.cr()
			r.cr()
		}
	case bf.BlockQuote:
		if entering {
			r.Tracer("BlockQuote (entering)", "")
			r.inBlockquote = true
			curleftmargin, _, _, _ := r.Pdf.GetMargins()
			r.Pdf.SetLeftMargin(curleftmargin + r.IndentValue)
			r.current = r.Blockquote
		} else {
			r.Tracer("BlockQuote (leaving)", "")
			r.inBlockquote = false
			curleftmargin, _, _, _ := r.Pdf.GetMargins()
			r.Pdf.SetLeftMargin(curleftmargin - r.IndentValue)
			r.current = r.Normal
			r.cr()
		}
	case bf.HTMLBlock:
		r.Tracer("HTMLBlock", "Not handled")
	case bf.Heading:
		if entering {
			r.cr()
			r.inHeading = true
			switch node.HeadingData.Level {
			case 1:
				r.Tracer("Heading (1, entering)", fmt.Sprintf("%v", node.HeadingData))
				r.current = r.H1
			case 2:
				r.Tracer("Heading (2, entering)", fmt.Sprintf("%v", node.HeadingData))
				r.current = r.H2
			case 3:
				r.Tracer("Heading (3, entering)", fmt.Sprintf("%v", node.HeadingData))
				r.current = r.H3
			case 4:
				r.Tracer("Heading (4, entering)", fmt.Sprintf("%v", node.HeadingData))
				r.current = r.H4
			case 5:
				r.Tracer("Heading (5, entering)", fmt.Sprintf("%v", node.HeadingData))
				r.current = r.H5
			case 6:
				r.Tracer("Heading (6, entering)", fmt.Sprintf("%v", node.HeadingData))
				r.current = r.H6
			}
		} else {
			r.Tracer("Heading (leaving)", "")
			r.current = r.Normal
			r.inHeading = false
		}
	case bf.HorizontalRule:
		r.Tracer("HorizontalRule", "Not handled")
	case bf.List:
		listKind := "Unordered"
		if node.ListFlags&bf.ListTypeOrdered != 0 {
			listKind = "Ordered"
		}
		if node.ListFlags&bf.ListTypeDefinition != 0 {
			listKind = "Definition"
		}
		if entering {
			switch listKind {
			case "Unordered":
			case "Ordered":
				r.currentItemNumber = 0
			case "Definition":
			}
			curleftmargin, _, _, _ := r.Pdf.GetMargins()
			r.Pdf.SetLeftMargin(curleftmargin + r.IndentValue)
			r.Tracer(fmt.Sprintf("%v List (entering)", listKind),
				fmt.Sprintf("%v", node.ListData))
		} else {
			curleftmargin, _, _, _ := r.Pdf.GetMargins()
			r.Pdf.SetLeftMargin(curleftmargin - r.IndentValue)
			r.Tracer(fmt.Sprintf("%v List (leaving)", listKind),
				fmt.Sprintf("%v", node.ListData))
			r.cr()
		}
	case bf.Item:
		listKind := "Unordered"
		if node.ListFlags&bf.ListTypeOrdered != 0 {
			listKind = "Ordered"
		}
		if node.ListFlags&bf.ListTypeDefinition != 0 {
			listKind = "Definition"
		}
		if entering {
			switch listKind {
			case "Unordered":
				r.inUnorderedItem = true
			case "Ordered":
				r.inOrderedItem = true
				r.currentItemNumber++
			case "Definition":
				r.inDefinitionItem = true
			}
			r.Tracer(fmt.Sprintf("%v Item (entering)", listKind),
				fmt.Sprintf("%v", node.ListData))
		} else {
			switch listKind {
			case "Unordered":
				r.inUnorderedItem = false
			case "Ordered":
				r.inOrderedItem = false
			case "Definition":
				r.inDefinitionItem = false
			}
			r.Tracer(fmt.Sprintf("%v Item (leaving)", listKind),
				fmt.Sprintf("%v", node.ListData))
			r.cr()
		}
	case bf.CodeBlock:
		r.Tracer("Codeblock", fmt.Sprintf("%v", node.CodeBlockData))
		r.Pdf.SetFillColor(200, 220, 255)
		r.setFont(r.Backtick)
		lines := strings.Split(strings.TrimSpace(string(node.Literal)), "\n")
		for n := range lines {
			r.Pdf.CellFormat(0, r.Backtick.Size,
				lines[n], "", 1, "LT", true, 0, "")
		}
		r.cr()

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
	r.Tracer("cr()", fmt.Sprintf("LH=%v", r.current.Size+r.current.Spacing))
	r.write(r.current, "\n")
}

// Tracer traces parse and pdf generation activity.
// Output goes to Stdout when DebugMode value is set to true
func (r *PdfRenderer) Tracer(source, msg string) {
	if r.tracerFile != "" {
		r.w.WriteString(fmt.Sprintf("[%v] %v\n", source, msg))
	}
}
