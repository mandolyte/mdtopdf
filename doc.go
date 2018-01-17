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

* The gofpdf packace to generate the PDF

The tests included here are from the BlackFriday package.
See the "testdata" folder.
The tests create PDF files and thus while the tests may complete
without errors, visual inspection of the created PDF is the
only way to determine if the tests *really* pass!

The tests create log files that trace the BlackFriday parser
callbacks. This is a valuable debug tool showing each callback
and data provided in each while the AST is presented.

Installation

To install the package:

	go get github.com/mandolyte/mdtopdf


Quick start

In the cmd folder is an example using the package. It demonstrates
a number of features. The test PDF was created with this command:

	go run convert.go -i test.md -o test.pdf

See README for limitations and known issues
*/

package mdtopdf
