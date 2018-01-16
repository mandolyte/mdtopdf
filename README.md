# mdtopdf
## Introduction: Markdown to PDF

This package depends on two other packages:
- The BlackFriday v2 parser to read the markdown source
- The `gofpdf` packace to generate the PDF

Both of the above are documented at Go Docs [http://godocs.org].

The tests included here are from the BlackFriday package.
See the "testdata" folder.
The tests create PDF files and thus while the tests may complete
without errors, visual inspection of the created PDF is the
only way to determine if the tests *really* pass!

The tests create log files that trace the BlackFriday parser
callbacks. This is a valuable debug tool showing each callback 
and data provided in each while the AST is presented.

## Limitations

1. It is common for Markdown to include HTML. HTML is treated as a "code block". *There is no attempt to convert raw HTML to PDF.*

2. Github-flavored Markdown permits strikethough using tildes. This is not supported at present by `gofpdf` as a font style.

3. The markdown link title, which would show when converted to HTML as hover-over text, is not supported. The generated PDF will show the actual URL that will be used if clicked, but this is a function of the PDF viewer.