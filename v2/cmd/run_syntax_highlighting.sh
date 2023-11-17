#!/bin/sh
cd md2pdf || exit
go run md2pdf.go -i test_syntax_highlighting.md -o test_syntax_highlighting.pdf -s ../gohighlight/syntax_files
