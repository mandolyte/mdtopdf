// testing
package mdtopdf

import (
	"io/ioutil"
	"path"
	"strings"
	"testing"
)

func TestAutoLinks(t *testing.T) {
	inputd := "./testdata/"
	inputf := "Auto links.text"
	input := path.Join(inputd, inputf)

	tracerfile := path.Join(inputd, strings.TrimRight(path.Base(input), ".text"))
	tracerfile += ".log"

	pdffile := path.Join(inputd, strings.TrimRight(path.Base(input), ".text"))
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

func TestAmpersandEncoding(t *testing.T) {
	inputd := "./testdata/"
	inputf := "Amps and angle encoding.text"
	input := path.Join(inputd, inputf)

	tracerfile := path.Join(inputd, strings.TrimRight(path.Base(input), ".text"))
	tracerfile += ".log"

	pdffile := path.Join(inputd, strings.TrimRight(path.Base(input), ".text"))
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

func TestInlineLinks(t *testing.T) {
	inputd := "./testdata/"
	inputf := "Links, inline style.text"
	input := path.Join(inputd, inputf)

	tracerfile := path.Join(inputd, strings.TrimRight(path.Base(input), ".text"))
	tracerfile += ".log"

	pdffile := path.Join(inputd, strings.TrimRight(path.Base(input), ".text"))
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

func TestLists(t *testing.T) {
	inputd := "./testdata/"
	inputf := "Ordered and unordered lists.text"
	input := path.Join(inputd, inputf)

	tracerfile := path.Join(inputd, strings.TrimRight(path.Base(input), ".text"))
	tracerfile += ".log"

	pdffile := path.Join(inputd, strings.TrimRight(path.Base(input), ".text"))
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

func TestStringEmph(t *testing.T) {
	inputd := "./testdata/"
	inputf := "Strong and em together.text"
	input := path.Join(inputd, inputf)

	tracerfile := path.Join(inputd, strings.TrimRight(path.Base(input), ".text"))
	tracerfile += ".log"

	pdffile := path.Join(inputd, strings.TrimRight(path.Base(input), ".text"))
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

func TestTabs(t *testing.T) {
	inputd := "./testdata/"
	inputf := "Tabs.text"
	input := path.Join(inputd, inputf)

	tracerfile := path.Join(inputd, strings.TrimRight(path.Base(input), ".text"))
	tracerfile += ".log"

	pdffile := path.Join(inputd, strings.TrimRight(path.Base(input), ".text"))
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
