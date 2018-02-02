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

			cleanedString := strings.TrimSpace(p)

			if cleanedString == "" {
				continue
			}

			if strings.Contains(cleanedString, ",") {
				cleanedString = strings.Replace(cleanedString, ",", " ", -1)
			}

			if i == 0 {
				cleanedLine = cleanedString
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

	newContentXML := ""

	pieces := strings.Split(odt.content, "<text:p text:style-name=\"Standard\"/>")

	if len(pieces) != 2 {
		return fmt.Errorf("AppendString() --> malformed template, consider replacing the ODT template")
	}

	lines := strings.Split(data, "\n")

	// no lines mean nothing to do
	if len(lines) == 0 {
		return nil
	}

	newContentXML += pieces[0]
	for _, line := range lines {

		// adjust the line to handle certain special characters
		fixedLine := line
		fixedLine = strings.Replace(fixedLine, ">", "&gt;", -1)
		fixedLine = strings.Replace(fixedLine, "<", "&lt;", -1)

		newContentXML += "<text:p text:style-name=\"Standard\">" + fixedLine + "</text:p>"
	}
	newContentXML += "<text:p text:style-name=\"Standard\"/>" + pieces[1]

	// replace the old content.xml with the newly generated content
	odt.content = newContentXML
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
