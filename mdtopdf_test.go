// testing
package mdtopdf

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"testing"
)

func TestStringEmph(t *testing.T) {
	inputd := "./testdata/"
	inputf := "Strong and em together.text"
	input := path.Join(inputd, inputf)

	tracerfile := path.Join(inputd, strings.TrimRight(path.Base(input), ".text"))
	tracerfile += ".log"

	pdffile := path.Join(inputd, strings.TrimRight(path.Base(input), ".text"))
	pdffile += ".pdf"
	fmt.Printf("pdffile is: %v\n", pdffile)
	fmt.Printf("tracerfile is: %v\n", tracerfile)

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
	fmt.Printf("pdffile is: %v\n", pdffile)
	fmt.Printf("tracerfile is: %v\n", tracerfile)

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
