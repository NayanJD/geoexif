package filesystem

import (
	"errors"
	"fmt"
	"geoexif/pkg/extractor"
	"log"
	"os"
	"strings"
)

type TypeResultWriter int

const (
	TypeCSVWriter TypeResultWriter = iota
	TypeHTMLWriter
)

type ResultWriter interface {
	WriteResult(*extractor.ExtractedResult) (int, error)
	WriteHeader() (int, error)
	Sync() error
	Close() error
}

func NewResultWriter(path string, t TypeResultWriter) (ResultWriter, error) {
	outputFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)

	if err != nil {
		return nil, err
		log.Fatalf("Could not create or open output.csv")
	}

	switch t {
	case TypeCSVWriter:
		return &CSVResultWriter{outputFile}, nil
	case TypeHTMLWriter:
		return &HTMLResultWriter{outputFile}, nil
	default:
		return nil, errors.New("Unknown output file formated specified")
	}
}

type CSVResultWriter struct {
	*os.File
}

func (cfw *CSVResultWriter) WriteResult(r *extractor.ExtractedResult) (int, error) {

	errStr := ""

	if r.Error != nil {
		errStr = r.Error.Error()
	}

	return cfw.WriteString(strings.Join([]string{r.ImagePath, r.Data, errStr, "\n"}, ","))
}

func (cfw *CSVResultWriter) WriteHeader() (int, error) {

	return cfw.WriteString("Image Path, Lat/Long, Comment\n")
}

type HTMLResultWriter struct {
	*os.File
}

func (cfw *HTMLResultWriter) WriteResult(r *extractor.ExtractedResult) (int, error) {

	errStr := ""

	if r.Error != nil {
		errStr = r.Error.Error()

		return cfw.WriteString(fmt.Sprintf(`
  <tr>
    <td>%v</td>
    <td>%v</td>
    <td>%v</td>
  </tr>
`, r.ImagePath, "", `<font color="red">`+errStr+`</font>`))
		// return cfw.WriteString(strings.Join([]string{"<li>", r.ImagePath, `<font color="red">`, errStr, `</font>`, "</li>\n"}, " "))
	} else {

		return cfw.WriteString(fmt.Sprintf(`
  <tr>
    <td>%v</td>
    <td>%v</td>
    <td>%v</td>
  </tr>
`, r.ImagePath, r.Data, ""))

		// return cfw.WriteString(strings.Join([]string{"<li>", r.ImagePath, `<font color="green">`, r.Data, `</font>`, "</li>\n"}, " "))

	}

}

func (cfw *HTMLResultWriter) WriteHeader() (int, error) {
	return cfw.WriteString(`
<head>
<style>
table {
  font-family: arial, sans-serif;
  border-collapse: collapse;
  width: 100%;
}

td, th {
  border: 1px solid #dddddd;
  text-align: left;
  padding: 8px;
}

tr:nth-child(even) {
  background-color: #dddddd;
}
</style>
</head>
<table>
  <tr>
    <th>Image Path</th>
    <th>Lat/Long</th>
    <th>Comment</th>
  </tr>
    `)
	// return cfw.WriteString("<h3>GPS Details of Images</h3>\n")
}
