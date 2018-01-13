package mdtopdf

import (
	"fmt"
	"strings"

	"github.com/jung-kurt/gofpdf"
	bf "gopkg.in/russross/blackfriday.v2"
)

func (r *PdfRenderer) processCodeblock(node *bf.Node) {
	r.Tracer("Codeblock", fmt.Sprintf("%v", node.CodeBlockData))
	r.Pdf.SetFillColor(200, 220, 255)
	r.setFont(r.Backtick)
	lines := strings.Split(strings.TrimSpace(string(node.Literal)), "\n")
	for n := range lines {
		r.Pdf.CellFormat(0, r.Backtick.Size,
			lines[n], "", 1, "LT", true, 0, "")
	}
}

func (r *PdfRenderer) processList(node *bf.Node, entering bool) {
	kind := unordered
	if node.ListFlags&bf.ListTypeOrdered != 0 {
		kind = ordered
	}
	if node.ListFlags&bf.ListTypeDefinition != 0 {
		kind = definition
	}
	r.setFont(r.Normal)
	if entering {
		r.Tracer(fmt.Sprintf("%v List (entering)", kind),
			fmt.Sprintf("%v", node.ListData))
		r.Pdf.SetLeftMargin(r.cs.peek().leftMargin + r.IndentValue)
		r.Tracer("... List Left Margin",
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
		r.Pdf.SetLeftMargin(r.cs.peek().leftMargin - r.IndentValue)
		r.Tracer("... Reset List Left Margin",
			fmt.Sprintf("re-set to %v", r.cs.peek().leftMargin-r.IndentValue))
		r.cs.pop()
		r.cr()
	}
}

func (r *PdfRenderer) processItem(node *bf.Node, entering bool) {
	if entering {
		r.Tracer(fmt.Sprintf("%v Item (entering) #%v",
			r.cs.peek().listkind, r.cs.peek().itemNumber+1),
			fmt.Sprintf("%v", node.ListData))
		x := &containerState{containerType: bf.Item,
			textStyle: r.Normal, itemNumber: r.cs.peek().itemNumber + 1,
			listkind:       r.cs.peek().listkind,
			firstParagraph: true,
			leftMargin:     r.cs.peek().leftMargin}
		// add bullet or itemnumber; then set left margin for the
		// text/paragraphs in the item
		r.cs.push(x)
		if r.cs.peek().listkind == unordered {
			r.Pdf.CellFormat(3*r.em, r.Normal.Size+r.Normal.Spacing,
				"-",
				"", 0, "RB", false, 0, "")
		} else if r.cs.peek().listkind == ordered {
			r.Pdf.CellFormat(3*r.em, r.Normal.Size+r.Normal.Spacing,
				fmt.Sprintf("%v.", r.cs.peek().itemNumber),
				"", 0, "RB", false, 0, "")
		}
		// with the bullet done, now set the left margin for the text
		r.Pdf.SetLeftMargin(r.cs.peek().leftMargin + (4 * r.em))
		// set the cursor to this point
		r.Pdf.SetX(r.cs.peek().leftMargin + (4 * r.em))
	} else {
		r.Tracer(fmt.Sprintf("%v Item (leaving)",
			r.cs.peek().listkind),
			fmt.Sprintf("%v", node.ListData))
		// before we the new line, reset left margin
		r.Pdf.SetLeftMargin(r.cs.peek().leftMargin)
		r.cr()
		r.cs.parent().itemNumber++
		r.cs.pop()
	}
}

func (r *PdfRenderer) processEmph(node *bf.Node, entering bool) {
	if entering {
		r.Tracer("Emph (entering)", "")
		r.cs.peek().textStyle.Style += "i"
	} else {
		r.Tracer("Emph (leaving)", "")
		r.cs.peek().textStyle.Style = strings.Replace(
			r.cs.peek().textStyle.Style, "i", "", -1)
	}
}

func (r *PdfRenderer) processStrong(node *bf.Node, entering bool) {
	if entering {
		r.Tracer("Strong (entering)", "")
		r.cs.peek().textStyle.Style += "b"
	} else {
		r.Tracer("Strong (leaving)", "")
		r.cs.peek().textStyle.Style = strings.Replace(
			r.cs.peek().textStyle.Style, "b", "", -1)
	}
}

func (r *PdfRenderer) processLink(node *bf.Node, entering bool) {
	if entering {
		x := &containerState{containerType: bf.Link,
			textStyle: r.Link, listkind: notlist,
			leftMargin:  r.cs.peek().leftMargin,
			destination: string(node.LinkData.Destination)}
		r.cs.push(x)
		r.Tracer("Link (entering)",
			fmt.Sprintf("Destination[%v] Title[%v]",
				string(node.LinkData.Destination),
				string(node.LinkData.Title)))
	} else {
		r.Tracer("Link (leaving)", "")
		r.cs.pop()
	}
}

func (r *PdfRenderer) processImage(node *bf.Node, entering bool) {
	// while this has entering and leaving states, it doesn't appear
	// to be useful except for other markup languages to close the tag
	if entering {
		r.Tracer("Image (entering)",
			fmt.Sprintf("Destination[%v] Title[%v]",
				string(node.LinkData.Destination),
				string(node.LinkData.Title)))
		r.Pdf.ImageOptions(string(node.LinkData.Destination),
			-1, 0, 0, 0, true,
			gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
	} else {
		r.Tracer("Image (leaving)", "")
	}
}

func (r *PdfRenderer) processCode(node *bf.Node) {
	r.Tracer("Code", "")
	r.setFont(r.Backtick)
	r.write(r.Backtick, string(node.Literal))
}

func (r *PdfRenderer) processParagraph(node *bf.Node, entering bool) {
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
				}
			}
			return
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
					r.Tracer("Not First Para within a list", "")
					r.cr()
				}
			}
			return
		}
		r.cr()
		r.cr()
	}
}

func (r *PdfRenderer) processBlockQuote(node *bf.Node, entering bool) {
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
}

func (r *PdfRenderer) processHeading(node *bf.Node, entering bool) {
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
}
