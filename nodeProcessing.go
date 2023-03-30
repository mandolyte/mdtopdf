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

package mdtopdf

import (
	"fmt"
	"os"
	"strings"

	"github.com/jung-kurt/gofpdf"
	bf "github.com/russross/blackfriday/v2"
)

func (r *PdfRenderer) processText(node *bf.Node) {
	currentStyle := r.cs.peek().textStyle
	r.setStyler(currentStyle)
	s := string(node.Literal)
	if !r.NeedBlockquoteStyleUpdate {
		s = strings.Replace(s, "\n", " ", -1)
	}
	r.tracer("Text", s)

	if r.cs.peek().containerType == bf.Link {
		r.writeLink(currentStyle, s, r.cs.peek().destination)
	} else if r.cs.peek().containerType == bf.Heading {
		//r.cr() // add space before heading
		r.write(currentStyle, s)
	} else if r.cs.peek().containerType == bf.TableCell {
		if r.cs.peek().isHeader {
			r.setStyler(currentStyle)
			// get the string width of header value
			hw := r.Pdf.GetStringWidth(s) + (2 * r.em)
			// now append it
			cellwidths = append(cellwidths, hw)
			// now write it...
			h, _ := r.Pdf.GetFontSize()
			h += currentStyle.Spacing
			r.tracer("... table header cell",
				fmt.Sprintf("Width=%v, height=%v", hw, h))

			r.Pdf.CellFormat(hw, h, s, "1", 0, "C", true, 0, "")
		} else {
			r.setStyler(currentStyle)
			hw := cellwidths[curdatacell]
			h := currentStyle.Size + currentStyle.Spacing
			r.tracer("... table body cell",
				fmt.Sprintf("Width=%v, height=%v", hw, h))
			r.Pdf.CellFormat(hw, h, s, "LR", 0, "", fill, 0, "")
		}
	} else if r.NeedBlockquoteStyleUpdate && r.cs.peek().containerType == bf.BlockQuote {
		r.tracer("Text BlockQuote", s)
		r.multiCell(currentStyle, s)	
	} else {
		r.write(currentStyle, s)
	}
}

func (r *PdfRenderer) processCodeblock(node *bf.Node) {
	r.tracer("Codeblock", fmt.Sprintf("%v", node.CodeBlockData))
	r.setStyler(r.Backtick)
	r.cr() // start on next line!
	r.multiCell(r.Backtick, string(node.Literal))
	/*
		lines := strings.Split(strings.TrimSpace(string(node.Literal)), "\n")
		for n := range lines {
			r.Pdf.CellFormat(0, r.Backtick.Size,
				lines[n], "", 1, "LT", true, 0, "")
		}
	*/
}

func (r *PdfRenderer) processList(node *bf.Node, entering bool) {
	kind := unordered
	if node.ListFlags&bf.ListTypeOrdered != 0 {
		kind = ordered
	}
	if node.ListFlags&bf.ListTypeDefinition != 0 {
		kind = definition
	}
	r.setStyler(r.Normal)
	if entering {
		r.tracer(fmt.Sprintf("%v List (entering)", kind),
			fmt.Sprintf("%v", node.ListData))
		r.Pdf.SetLeftMargin(r.cs.peek().leftMargin + r.IndentValue)
		r.tracer("... List Left Margin",
			fmt.Sprintf("set to %v", r.cs.peek().leftMargin+r.IndentValue))
		x := &containerState{containerType: bf.List,
			textStyle: r.Normal, itemNumber: 0,
			listkind:   kind,
			leftMargin: r.cs.peek().leftMargin + r.IndentValue}
		// before pushing check to see if this is a sublist
		// if so, then output a newline
		/*
			if r.cs.peek().containerType == bf.Item {
				r.cr()
			}
		*/
		r.cs.push(x)
	} else {
		r.tracer(fmt.Sprintf("%v List (leaving)", kind),
			fmt.Sprintf("%v", node.ListData))
		r.Pdf.SetLeftMargin(r.cs.peek().leftMargin - r.IndentValue)
		r.tracer("... Reset List Left Margin",
			fmt.Sprintf("re-set to %v", r.cs.peek().leftMargin-r.IndentValue))
		r.cs.pop()
		if len(r.cs.stack) < 2 {
			r.cr()
		}
	}
}

func (r *PdfRenderer) processItem(node *bf.Node, entering bool) {
	if entering {
		r.tracer(fmt.Sprintf("%v Item (entering) #%v",
			r.cs.peek().listkind, r.cs.peek().itemNumber+1),
			fmt.Sprintf("%v", node.ListData))
		r.cr() // newline before getting started
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
		r.tracer(fmt.Sprintf("%v Item (leaving)",
			r.cs.peek().listkind),
			fmt.Sprintf("%v", node.ListData))
		// before we output the new line, reset left margin
		r.Pdf.SetLeftMargin(r.cs.peek().leftMargin)
		//r.cr()
		r.cs.parent().itemNumber++
		r.cs.pop()
	}
}

func (r *PdfRenderer) processEmph(node *bf.Node, entering bool) {
	if entering {
		r.tracer("Emph (entering)", "")
		r.cs.peek().textStyle.Style += "i"
	} else {
		r.tracer("Emph (leaving)", "")
		r.cs.peek().textStyle.Style = strings.Replace(
			r.cs.peek().textStyle.Style, "i", "", -1)
	}
}

func (r *PdfRenderer) processStrong(node *bf.Node, entering bool) {
	if entering {
		r.tracer("Strong (entering)", "")
		r.cs.peek().textStyle.Style += "b"
	} else {
		r.tracer("Strong (leaving)", "")
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
		r.tracer("Link (entering)",
			fmt.Sprintf("Destination[%v] Title[%v]",
				string(node.LinkData.Destination),
				string(node.LinkData.Title)))
	} else {
		r.tracer("Link (leaving)", "")
		r.cs.pop()
	}
}

func (r *PdfRenderer) processImage(node *bf.Node, entering bool) {
	// while this has entering and leaving states, it doesn't appear
	// to be useful except for other markup languages to close the tag
	if entering {
		r.tracer("Image (entering)",
			fmt.Sprintf("Destination[%v] Title[%v]",
				string(node.LinkData.Destination),
				string(node.LinkData.Title)))
		// following changes suggested by @sirnewton01, issue #6
		// does file exist?
		var imgPath = string(node.LinkData.Destination)
		_, err := os.Stat(imgPath)
		if err == nil {
			r.Pdf.ImageOptions(string(node.LinkData.Destination),
				-1, 0, 0, 0, true,
				gofpdf.ImageOptions{ImageType: "", ReadDpi: true}, 0, "")
		} else {
			r.tracer("Image (file error)", err.Error())
		}
	} else {
		r.tracer("Image (leaving)", "")
	}
}

func (r *PdfRenderer) processCode(node *bf.Node) {
	r.tracer("processCode", fmt.Sprintf("%s", string(node.Literal)))
	if r.NeedCodeStyleUpdate {
		r.tracer("Code (entering)", "")
		r.setStyler(r.Code)
		s := string(node.Literal)
		hw := r.Pdf.GetStringWidth(s) + (1 * r.em)
		h := r.Code.Size
		r.Pdf.CellFormat(hw, h, s, "", 0, "C", true, 0, "")
	} else {
		r.tracer("Backtick (entering)", "")
		r.setStyler(r.Backtick)
		r.write(r.Backtick, string(node.Literal))
	}
}

func (r *PdfRenderer) processParagraph(node *bf.Node, entering bool) {
	r.setStyler(r.Normal)
	if entering {
		r.tracer("Paragraph (entering)", "")
		lm, tm, rm, bm := r.Pdf.GetMargins()
		r.tracer("... Margins (left, top, right, bottom:",
			fmt.Sprintf("%v %v %v %v", lm, tm, rm, bm))
		if r.cs.peek().containerType == bf.Item {
			t := r.cs.peek().listkind
			if t == unordered || t == ordered || t == definition {
				if r.cs.peek().firstParagraph {
					r.tracer("First Para within a list", "breaking")
				} else {
					r.tracer("Not First Para within a list", "indent etc.")
					r.cr()
				}
			}
			return
		}
		r.cr()
		//r.cr()
	} else {
		r.tracer("Paragraph (leaving)", "")
		lm, tm, rm, bm := r.Pdf.GetMargins()
		r.tracer("... Margins (left, top, right, bottom:",
			fmt.Sprintf("%v %v %v %v", lm, tm, rm, bm))
		if r.cs.peek().containerType == bf.Item {
			t := r.cs.peek().listkind
			if t == unordered || t == ordered || t == definition {
				if r.cs.peek().firstParagraph {
					r.cs.peek().firstParagraph = false
				} else {
					r.tracer("Not First Para within a list", "")
					r.cr()
				}
			}
			return
		}
		//r.cr()
		r.cr()
	}
}

func (r *PdfRenderer) processBlockQuote(node *bf.Node, entering bool) {
	if entering {
		r.tracer("BlockQuote (entering)", "")
		curleftmargin, _, _, _ := r.Pdf.GetMargins()
		x := &containerState{containerType: bf.BlockQuote,
			textStyle: r.Blockquote, listkind: notlist,
			leftMargin: curleftmargin + r.IndentValue}
		r.cs.push(x)
		r.Pdf.SetLeftMargin(curleftmargin + r.IndentValue)
	} else {
		r.tracer("BlockQuote (leaving)", "")
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
			r.tracer("Heading (1, entering)", fmt.Sprintf("%v", node.HeadingData))
			x := &containerState{containerType: bf.Heading,
				textStyle: r.H1, listkind: notlist,
				leftMargin: r.cs.peek().leftMargin}
			r.cs.push(x)
		case 2:
			r.tracer("Heading (2, entering)", fmt.Sprintf("%v", node.HeadingData))
			x := &containerState{containerType: bf.Heading,
				textStyle: r.H2, listkind: notlist,
				leftMargin: r.cs.peek().leftMargin}
			r.cs.push(x)
		case 3:
			r.tracer("Heading (3, entering)", fmt.Sprintf("%v", node.HeadingData))
			x := &containerState{containerType: bf.Heading,
				textStyle: r.H3, listkind: notlist,
				leftMargin: r.cs.peek().leftMargin}
			r.cs.push(x)
		case 4:
			r.tracer("Heading (4, entering)", fmt.Sprintf("%v", node.HeadingData))
			x := &containerState{containerType: bf.Heading,
				textStyle: r.H4, listkind: notlist,
				leftMargin: r.cs.peek().leftMargin}
			r.cs.push(x)
		case 5:
			r.tracer("Heading (5, entering)", fmt.Sprintf("%v", node.HeadingData))
			x := &containerState{containerType: bf.Heading,
				textStyle: r.H5, listkind: notlist,
				leftMargin: r.cs.peek().leftMargin}
			r.cs.push(x)
		case 6:
			r.tracer("Heading (6, entering)", fmt.Sprintf("%v", node.HeadingData))
			x := &containerState{containerType: bf.Heading,
				textStyle: r.H6, listkind: notlist,
				leftMargin: r.cs.peek().leftMargin}
			r.cs.push(x)
		}
	} else {
		r.tracer("Heading (leaving)", "")
		r.cr()
		r.cs.pop()
	}
}

func (r *PdfRenderer) processHorizontalRule(node *bf.Node) {
	r.tracer("HorizontalRule", "")
	// do a newline
	r.cr()
	// get the current x and y (assume left margin in ok)
	x, y := r.Pdf.GetXY()
	// get the page margins
	lm, _, _, _ := r.Pdf.GetMargins()
	// get the page size
	w, _ := r.Pdf.GetPageSize()
	// now compute the x value of the right side of page
	newx := w - lm
	r.tracer("... From X,Y", fmt.Sprintf("%v,%v", x, y))
	r.Pdf.MoveTo(x, y)
	r.tracer("...   To X,Y", fmt.Sprintf("%v,%v", newx, y))
	r.Pdf.LineTo(newx, y)
	r.Pdf.SetLineWidth(3)
	r.Pdf.SetFillColor(200, 200, 200)
	r.Pdf.DrawPath("F")
	// another newline
	r.cr()
}

func (r *PdfRenderer) processHTMLBlock(node *bf.Node) {
	r.tracer("HTMLBlock", string(node.Literal))
	r.cr()
	r.setStyler(r.Backtick)
	r.Pdf.CellFormat(0, r.Backtick.Size,
		string(node.Literal), "", 1, "LT", true, 0, "")
	r.cr()
}

func (r *PdfRenderer) processTable(node *bf.Node, entering bool) {
	if entering {
		r.tracer("Table (entering)", "")
		x := &containerState{containerType: bf.Table,
			textStyle: r.THeader, listkind: notlist,
			leftMargin: r.cs.peek().leftMargin}
		r.cr()
		r.cs.push(x)
		fill = false
	} else {
		wSum := 0.0
		for _, w := range cellwidths {
			wSum += w
		}
		r.Pdf.CellFormat(wSum, 0, "", "T", 0, "", false, 0, "")

		r.cs.pop()
		r.tracer("Table (leaving)", "")
		r.cr()
	}
}

func (r *PdfRenderer) processTableHead(node *bf.Node, entering bool) {
	if entering {
		r.tracer("TableHead (entering)", "")
		x := &containerState{containerType: bf.TableHead,
			textStyle: r.THeader, listkind: notlist,
			leftMargin: r.cs.peek().leftMargin}
		r.cs.push(x)
		cellwidths = make([]float64, 0)
	} else {
		r.cs.pop()
		r.tracer("TableHead (leaving)", "")
	}
}

func (r *PdfRenderer) processTableBody(node *bf.Node, entering bool) {
	if entering {
		r.tracer("TableBody (entering)", "")
		x := &containerState{containerType: bf.TableBody,
			textStyle: r.TBody, listkind: notlist,
			leftMargin: r.cs.peek().leftMargin}
		r.cs.push(x)
	} else {
		r.cs.pop()
		r.tracer("TableBody (leaving)", "")
		r.Pdf.Ln(-1)
	}
}

func (r *PdfRenderer) processTableRow(node *bf.Node, entering bool) {
	if entering {
		r.tracer("TableRow (entering)", "")
		x := &containerState{containerType: bf.TableRow,
			textStyle: r.TBody, listkind: notlist,
			leftMargin: r.cs.peek().leftMargin}
		if r.cs.peek().isHeader {
			x.textStyle = r.THeader
		}
		r.Pdf.Ln(-1)

		// initialize cell widths slice; only one table at a time!
		curdatacell = 0
		r.cs.push(x)
	} else {
		r.cs.pop()
		r.tracer("TableRow (leaving)", "")
		fill = !fill
	}
}

func (r *PdfRenderer) processTableCell(node *bf.Node, entering bool) {
	if entering {
		r.tracer("TableCell (entering)", "")
		x := &containerState{containerType: bf.TableCell,
			textStyle: r.Normal, listkind: notlist,
			leftMargin: r.cs.peek().leftMargin}
		if node.TableCellData.IsHeader {
			x.isHeader = true
			x.textStyle = r.THeader
			r.setStyler(r.THeader)
		} else {
			x.textStyle = r.TBody
			r.setStyler(r.TBody)
			x.isHeader = false
		}
		r.cs.push(x)
	} else {
		r.cs.pop()
		r.tracer("TableCell (leaving)", "")
		curdatacell++
	}
}
