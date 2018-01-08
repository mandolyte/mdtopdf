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
		if r.cs.peek().listkind == unordered {
			r.Tracer("Text in list",
				fmt.Sprintf("Type is %v", r.cs.peek().listkind))
			if r.cs.peek().firstParagraph {
				r.Tracer("... is first paragraph", "Add bullet")
				//r.write(currentStyle, "- ")
				r.Pdf.CellFormat(3*r.em, r.Normal.Size+r.Normal.Spacing,
					"- ",
					"", 0, "RB", false, 0, "")
			}
		} else if r.cs.peek().listkind == ordered {
			n := r.cs.peek().itemNumber
			r.Tracer("Text in list",
				fmt.Sprintf("Type is %v, item #%v",
					r.cs.peek().listkind, n))
			//r.write(currentStyle, fmt.Sprintf("%v. ", n))
			// right justify the item number in a cell which
			// is the width of 3*em (normal font)
			if r.cs.peek().firstParagraph {
				r.Tracer("... is first paragraph", "Add number")
				r.Pdf.CellFormat(3*r.em, r.Normal.Size+r.Normal.Spacing,
					fmt.Sprintf("%v. ", n),
					"", 0, "RB", false, 0, "")
			}
		}
		r.write(currentStyle, s)

	case bf.Softbreak:
		r.Tracer("Softbreak", "Not handled")
	case bf.Hardbreak:
		r.Tracer("Hardbreak", "Not handled")
	case bf.Emph:
		if entering {
			r.Tracer("Emph (entering)", "")
			r.cs.peek().textStyle.Style += "i"
		} else {
			r.Tracer("Emph (leaving)", "")
			r.cs.peek().textStyle.Style = strings.Replace(
				r.cs.peek().textStyle.Style, "i", "", -1)
		}
	case bf.Strong:
		if entering {
			r.Tracer("Strong (entering)", "")
			r.cs.peek().textStyle.Style += "b"
		} else {
			r.Tracer("Strong (leaving)", "")
			r.cs.peek().textStyle.Style = strings.Replace(
				r.cs.peek().textStyle.Style, "b", "", -1)
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
			lm, tm, rm, bm := r.Pdf.GetMargins()
			r.Tracer("... Margins (left, top, right, bottom:",
				fmt.Sprintf("%v %v %v %v", lm, tm, rm, bm))
			if r.cs.peek().containerType == bf.Item {
				t := r.cs.peek().listkind
				if t == unordered || t == ordered || t == definition {
					if r.cs.peek().firstParagraph {
						r.Tracer("First Para within a list", "breaking")
					} else {
						r.Tracer("Not First Para within a list", "indent etc.")
						r.cr()
						r.cr()
						//curleftmargin, _, _, _ := r.Pdf.GetMargins()
						r.cs.peek().leftMargin += r.IndentValue
						r.Pdf.SetLeftMargin(r.cs.peek().leftMargin)
						r.Tracer("... Left Margin Set:",
							fmt.Sprintf("%v", r.cs.peek().leftMargin))
					}
					r.Tracer("Para within a list", "breaking")
					break
				}
			}
			r.cr()
		} else {
			r.Tracer("Paragraph (leaving)", "")
			lm, tm, rm, bm := r.Pdf.GetMargins()
			r.Tracer("... Margins (left, top, right, bottom:",
				fmt.Sprintf("%v %v %v %v", lm, tm, rm, bm))
			if r.cs.peek().containerType == bf.Item {
				t := r.cs.peek().listkind
				if t == unordered || t == ordered || t == definition {
					if r.cs.peek().firstParagraph {
						r.cs.peek().firstParagraph = false
					} else {
						r.Tracer("Not First Para within a list", "undent etc.")
						r.cr()
						//curleftmargin, _, _, _ := r.Pdf.GetMargins()
						r.cs.peek().leftMargin -= r.IndentValue
						r.Pdf.SetLeftMargin(r.cs.peek().leftMargin)
						r.Tracer("... Left Margin Set:",
							fmt.Sprintf("%v", r.cs.peek().leftMargin))
					}
					r.Tracer("Para within a list", "breaking")
					break
				}
			}
			r.cr()
			r.cr()
		}
	case bf.BlockQuote:
		if entering {
			r.Tracer("BlockQuote (entering)", "")
			curleftmargin, _, _, _ := r.Pdf.GetMargins()
			x := &containerState{containerType: bf.BlockQuote,
				textStyle: r.Blockquote, listkind: notlist,
				leftMargin: curleftmargin + r.IndentValue}
			r.cs.push(x)
			r.Pdf.SetLeftMargin(curleftmargin + r.IndentValue)
		} else {
			r.Tracer("BlockQuote (leaving)", "")
			curleftmargin, _, _, _ := r.Pdf.GetMargins()
			r.Pdf.SetLeftMargin(curleftmargin - r.IndentValue)
			r.cs.pop()
			r.cr()
		}
	case bf.HTMLBlock:
		r.Tracer("HTMLBlock", "Not handled")
	case bf.Heading:
		if entering {
			r.cr()
			//r.inHeading = true
			switch node.HeadingData.Level {
			case 1:
				r.Tracer("Heading (1, entering)", fmt.Sprintf("%v", node.HeadingData))
				x := &containerState{containerType: bf.Heading,
					textStyle: r.H1, listkind: notlist,
					leftMargin: r.cs.peek().leftMargin}
				r.cs.push(x)
			case 2:
				r.Tracer("Heading (2, entering)", fmt.Sprintf("%v", node.HeadingData))
				x := &containerState{containerType: bf.Heading,
					textStyle: r.H2, listkind: notlist,
					leftMargin: r.cs.peek().leftMargin}
				r.cs.push(x)
			case 3:
				r.Tracer("Heading (3, entering)", fmt.Sprintf("%v", node.HeadingData))
				x := &containerState{containerType: bf.Heading,
					textStyle: r.H3, listkind: notlist,
					leftMargin: r.cs.peek().leftMargin}
				r.cs.push(x)
			case 4:
				r.Tracer("Heading (4, entering)", fmt.Sprintf("%v", node.HeadingData))
				x := &containerState{containerType: bf.Heading,
					textStyle: r.H4, listkind: notlist,
					leftMargin: r.cs.peek().leftMargin}
				r.cs.push(x)
			case 5:
				r.Tracer("Heading (5, entering)", fmt.Sprintf("%v", node.HeadingData))
				x := &containerState{containerType: bf.Heading,
					textStyle: r.H5, listkind: notlist,
					leftMargin: r.cs.peek().leftMargin}
				r.cs.push(x)
			case 6:
				r.Tracer("Heading (6, entering)", fmt.Sprintf("%v", node.HeadingData))
				x := &containerState{containerType: bf.Heading,
					textStyle: r.H6, listkind: notlist,
					leftMargin: r.cs.peek().leftMargin}
				r.cs.push(x)
			}
		} else {
			r.Tracer("Heading (leaving)", "")
			r.cr()
			r.cs.pop()
		}
	case bf.HorizontalRule:
		r.Tracer("HorizontalRule", "Not handled")
	case bf.List:
		kind := unordered
		if node.ListFlags&bf.ListTypeOrdered != 0 {
			kind = ordered
		}
		if node.ListFlags&bf.ListTypeDefinition != 0 {
			kind = definition
		}
		if entering {
			r.Tracer(fmt.Sprintf("%v List (entering)", kind),
				fmt.Sprintf("%v", node.ListData))
			//curleftmargin, _, _, _ := r.Pdf.GetMargins()
			r.Pdf.SetLeftMargin(r.cs.peek().leftMargin + r.IndentValue)
			r.Tracer("... Left Margin",
				fmt.Sprintf("set to %v", r.cs.peek().leftMargin+r.IndentValue))
			x := &containerState{containerType: bf.List,
				textStyle: r.Normal, itemNumber: 0,
				listkind:   kind,
				leftMargin: r.cs.peek().leftMargin + r.IndentValue}
			// before pushing check to see if this is a sublist
			// if so, then output a newline
			if r.cs.peek().containerType == bf.Item {
				r.cr()
			}
			r.cs.push(x)
		} else {
			r.Tracer(fmt.Sprintf("%v List (leaving)", kind),
				fmt.Sprintf("%v", node.ListData))
			//curleftmargin, _, _, _ := r.Pdf.GetMargins()
			r.Pdf.SetLeftMargin(r.cs.peek().leftMargin - r.IndentValue)
			r.Tracer("... Left Margin",
				fmt.Sprintf("re-set to %v", r.cs.peek().leftMargin-r.IndentValue))
			r.cs.pop()
			r.cr()
		}
	case bf.Item:
		if entering {
			r.Tracer(fmt.Sprintf("%v Item (entering) #%v",
				r.cs.peek().listkind, r.cs.peek().itemNumber+1),
				fmt.Sprintf("%v", node.ListData))
			x := &containerState{containerType: bf.Item,
				textStyle: r.Normal, itemNumber: r.cs.peek().itemNumber + 1,
				listkind:       r.cs.peek().listkind,
				firstParagraph: true,
				leftMargin:     r.cs.peek().leftMargin}
			r.cs.push(x)
		} else {
			r.Tracer(fmt.Sprintf("%v Item (leaving)",
				r.cs.peek().listkind),
				fmt.Sprintf("%v", node.ListData))
			r.cr()
			r.cs.parent().itemNumber++
			r.cs.pop()
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
		r.w.WriteString(fmt.Sprintf("[%v] %v\n", source, msg))
	}
}
