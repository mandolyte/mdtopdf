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
	"io/ioutil"
	"path"
	"strings"
	"testing"
)

func testit(inputf string, t *testing.T) {
	inputd := "./testdata/"
	input := path.Join(inputd, inputf)

	tracerfile := path.Join(inputd, strings.TrimSuffix(path.Base(input), ".text"))
	tracerfile += ".log"

	pdffile := path.Join(inputd, strings.TrimSuffix(path.Base(input), ".text"))
	pdffile += ".pdf"

	content, err := ioutil.ReadFile(input)
	if err != nil {
		t.Errorf("%v:%v", input, err)
	}

	r := NewPdfRenderer(pdffile, tracerfile)
	err = r.Process(content)
	if err != nil {
		t.Error(err)
	}
}

func TestMarkdownDocumenationBasic(t *testing.T) {
	testit("Markdown Documentation - Basics.text", t)
}

func TestMarkdownDocumenationSyntax(t *testing.T) {
	testit("Markdown Documentation - Syntax.text", t)
}

func TestImage(t *testing.T) {
	testit("Image.text", t)
}

func TestAutoLinks(t *testing.T) {
	testit("Auto links.text", t)
}

func TestAmpersandEncoding(t *testing.T) {
	testit("Amps and angle encoding.text", t)
}

func TestInlineLinks(t *testing.T) {
	testit("Links, inline style.text", t)
}

func TestLists(t *testing.T) {
	testit("Ordered and unordered lists.text", t)
}

func TestStringEmph(t *testing.T) {
	testit("Strong and em together.text", t)
}

func TestTabs(t *testing.T) {
	testit("Tabs.text", t)
}

func TestBackslashEscapes(t *testing.T) {
	testit("Backslash escapes.text", t)
}

func TestBackquotes(t *testing.T) {
	testit("Blockquotes with code blocks.text", t)
}

func TestCodeBlocks(t *testing.T) {
	testit("Code Blocks.text", t)
}

func TestCodeSpans(t *testing.T) {
	testit("Code Spans.text", t)
}

func TestHardWrappedPara(t *testing.T) {
	testit("Hard-wrapped paragraphs with list-like lines no empty line before block.text", t)
}

func TestHardWrappedPara2(t *testing.T) {
	testit("Hard-wrapped paragraphs with list-like lines.text", t)
}

func TestHorizontalRules(t *testing.T) {
	testit("Horizontal rules.text", t)
}

func TestInlineHtmlSimple(t *testing.T) {
	testit("Inline HTML (Simple).text", t)
}

func TestInlineHtmlAdvanced(t *testing.T) {
	testit("Inline HTML (Advanced).text", t)
}

func TestInlineHtmlComments(t *testing.T) {
	testit("Inline HTML comments.text", t)
}

func TestTitleWithQuotes(t *testing.T) {
	testit("Literal quotes in titles.text", t)
}

func TestNestedBlockquotes(t *testing.T) {
	testit("Nested blockquotes.text", t)
}

func TestLinksReference(t *testing.T) {
	testit("Links, reference style.text", t)
}

func TestLinksShortcut(t *testing.T) {
	testit("Links, shortcut references.text", t)
}

func TestTidyness(t *testing.T) {
	testit("Tidyness.text", t)
}
