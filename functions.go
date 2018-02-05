package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// convertRosewoodToCSV ... turns an array of lines of a Rosewood table into CSV
func convertRosewoodToCSV(lines []string, num int) (string, error) {

	if len(lines) < 1 {
		return "", fmt.Errorf("convertRosewoodToCSV() --> invalid input")
	}

	titleHasNotBeenPrinted := true
	result := ""

	for _, l := range lines {

		trimmedLine := strings.TrimSpace(l)

		if trimmedLine == "" {
			continue
		}

		if trimmedLine == "---" {
			continue
		}

		if titleHasNotBeenPrinted {
			numAsStr := strconv.Itoa(num)
			result += "Table " + numAsStr + ": " + trimmedLine + "\n"
			titleHasNotBeenPrinted = false
			continue
		}

		pieces := strings.Split(l, "|")

		// Rosewood instructions are exactly one piece, so check for 2+
		if len(pieces) < 2 {
			continue
		}

		cleanedLine := ""
		for i, p := range pieces {

			// sometimes for the first element, there is a
			// Rosewood "  " prefix, so this needs to re-append
			// it to the trimmed string
			prefix := ""
			if i == 0 && strings.HasPrefix(p, "  ") && PrintAsCSV {
				prefix = "  "
			} else if i == 0 && strings.HasPrefix(p, "  ") {
				prefix = ":rosewood-odt-space:"
			}

			cleanedString := strings.TrimSpace(p)

			if cleanedString == "" {
				continue
			}

			if strings.Contains(cleanedString, ",") {
				cleanedString = strings.Replace(cleanedString, ",", " ", -1)
			}

			if i == 0 {
				cleanedLine = prefix + cleanedString
			} else {
				cleanedLine += "," + cleanedString
			}
		}

		result += cleanedLine + "\n"
	}

	return result, nil
}

// ReadOdtFile ... read contents of Odt file
func ReadOdtFile(templateName string) (*CachedOdtTemplate, error) {

	if templateName == "" {
		return nil, fmt.Errorf("ReadOdtFile() --> invalid input")
	}

	path := filepath.Join(DefaultTemplatesDir, templateName)

	//
	// decompress the ODT file as it is in Zip format
	//
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}

	//
	// obtain the strings from content.xml, settings.xml, and styles.xml
	//

	content, err := readFile(reader.File, "content.xml")
	if err != nil {
		return nil, err
	}

	settings, err := readFile(reader.File, "settings.xml")
	if err != nil {
		return nil, err
	}

	styles, err := readFile(reader.File, "styles.xml")
	if err != nil {
		return nil, err
	}

	return &CachedOdtTemplate{zipReader: reader, content: content, settings: settings, styles: styles}, nil
}

// New ... pass back an new editable instance of the ODT file
func (r *CachedOdtTemplate) New() *Odt {
	return &Odt{
		files:    r.zipReader.File,
		content:  r.content,
		settings: r.settings,
		styles:   r.styles,
	}
}

// readContent ... open content.xml from the cached ODT file
func readFile(files []*zip.File, filename string) (string, error) {

	if files == nil || len(files) == 0 || filename == "" {
		return "", fmt.Errorf("readContentFile --> invalid input")
	}

	var fileOfInterest *zip.File
	var documentReader io.ReadCloser

	for _, f := range files {
		if f.Name == filename {
			fileOfInterest = f
			break
		}
	}

	if fileOfInterest == nil {
		return "", fmt.Errorf("readContentFile --> content.xml not found")
	}

	documentReader, err := fileOfInterest.Open()
	if err != nil {
		return "", err
	}

	bytes, err := ioutil.ReadAll(documentReader)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// AppendStrings ... attach plain-text word to the document in question
func (odt *Odt) AppendStrings(data string) error {

	// no data means nothing to do
	if data == "" {
		return nil
	}

	//
	// Append the new document styles
	//

	newContentXML := strings.Replace(odt.content, "<office:automatic-styles/>", "<office:automatic-styles>"+
		"<style:style style:name=\"P1\" style:family=\"paragraph\" style:parent-style-name=\"Standard\">"+
		"<style:paragraph-properties fo:break-before=\"page\"/></style:style>"+
		"<style:style style:name=\"P2\" style:family=\"paragraph\" style:parent-style-name=\"Footer\">"+
		"<style:paragraph-properties fo:text-align=\"end\" style:justify-single-word=\"false\"/>"+
		"</style:style></office:automatic-styles>", -1)

	// replace the old content.xml with the newly generated content
	odt.content = newContentXML
	newContentXML = ""

	//
	// Append the new footer styles
	//

	newStylesXML := strings.Replace(odt.styles, "</style:style><text:outline-style style:name=\"Outline\">",
		"</style:style>"+
			"<style:style style:name=\"Footer\" style:family=\"paragraph\" style:parent-style-name=\"Standard\" style:class=\"extra\">"+
			"<style:paragraph-properties text:number-lines=\"false\" text:line-number=\"0\">"+
			"<style:tab-stops>"+
			"<style:tab-stop style:position=\"8.795cm\" style:type=\"center\"/>"+
			"<style:tab-stop style:position=\"17.59cm\" style:type=\"right\"/>"+
			"</style:tab-stops>"+
			"</style:paragraph-properties>"+
			"</style:style>"+
			"<text:outline-style style:name=\"Outline\">", -1)

	newStylesXML = strings.Replace(newStylesXML, "<office:automatic-styles><style:page-layout style:name=\"Mpm1\">",
		"<office:automatic-styles><style:style style:name=\"MP1\" style:family=\"paragraph\" style:parent-style-name=\"Footer\">"+
			"<style:paragraph-properties fo:text-align=\"end\" style:justify-single-word=\"false\"/>"+
			"</style:style><style:page-layout style:name=\"Mpm1\">", -1)

	newStylesXML = strings.Replace(newStylesXML, "<style:footer-style/>",
		"<style:footer-style>"+
			"<style:header-footer-properties fo:min-height=\"0cm\" fo:margin-top=\"0.499cm\"/>"+
			"</style:footer-style>", -1)

	newStylesXML = strings.Replace(newStylesXML, "<style:master-page style:name=\"Standard\" style:page-layout-name=\"Mpm1\"/>",
		"<style:master-page style:name=\"Standard\" style:page-layout-name=\"Mpm1\">"+
			"<style:footer>"+
			"<text:p text:style-name=\"MP1\">"+
			"<text:page-number text:select-page=\"current\">1</text:page-number>"+
			"</text:p>"+
			"</style:footer>"+
			"</style:master-page>", -1)

	// replace the old styles.xml with the newly generated content
	odt.styles = newStylesXML
	newStylesXML = ""

	//
	// Append the string text
	//

	pieces := strings.Split(odt.content, "<text:p text:style-name=\"Standard\"/>")
	if len(pieces) != 2 {
		return fmt.Errorf("AppendString() --> malformed template, consider replacing the ODT template")
	}

	lines := strings.Split(data, "\n")
	if len(lines) == 0 {
		return nil
	}

	newContentXML += pieces[0]
	for _, line := range lines {

		// adjust the line to handle certain special characters
		fixedLine := line
		fixedLine = strings.Replace(fixedLine, ">", "&gt;", -1)
		fixedLine = strings.Replace(fixedLine, "<", "&lt;", -1)

		if strings.Contains(fixedLine, ":rosewood-page-break:") {
			fixedLine = strings.Replace(fixedLine, ":rosewood-page-break:",
				"<text:p text:style-name=\"Standard\"/><text:p text:style-name=\"Standard\"/><text:p text:style-name=\"P1\">", -1)
			newContentXML += fixedLine + "</text:p>"
		} else {
			newContentXML += "<text:p text:style-name=\"Standard\">" + fixedLine + "</text:p>"
		}
	}
	newContentXML += "<text:p text:style-name=\"Standard\"/>" + pieces[1]

	// replace the old content.xml with the newly generated content
	odt.content = newContentXML

	//
	// Handle start-of-column spacing
	//

	newContentXML = strings.Replace(odt.content, ":rosewood-odt-space:", "<text:s text:c=\"2\"/>", -1)

	// replace the old content.xml with the newly generated content
	odt.content = newContentXML
	newContentXML = ""

	return nil
}

// Write ... take the modified ODT file in memory and write it to a file
func (odt *Odt) Write(path string) error {

	if odt.files == nil || odt.content == "" || path == "" {
		return fmt.Errorf("Write() --> invalid input")
	}

	newFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Write() --> unable to create ODT file: " + path)
	}
	defer newFile.Close()

	zipWriter := zip.NewWriter(newFile)

	for _, file := range odt.files {

		var writer io.Writer
		var readCloser io.ReadCloser

		writer, err = zipWriter.Create(file.Name)
		if err != nil {
			return err
		}

		readCloser, err = file.Open()
		if err != nil {
			return err
		}

		//
		// Handle each of the subfiles of interest
		//

		switch file.Name {

		case "content.xml":
			writer.Write([]byte(odt.content))
		case "styles.xml":
			writer.Write([]byte(odt.styles))
		case "settings.xml":
			writer.Write([]byte(odt.settings))
		default:
			writer.Write(streamToByte(readCloser))
		}
	}
	zipWriter.Close()

	return nil
}

// streamToByte ... convert a string stream to a byte array
func streamToByte(stream io.Reader) []byte {

	if stream == nil {
		return []byte{}
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)

	return buf.Bytes()
}
