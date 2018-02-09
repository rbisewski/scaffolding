package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// convertRosewoodToCSV ... turns an array of lines of a Rosewood table into CSV
func convertRosewoodToCSV(lines []string, num int) (string, error) {

	if len(lines) < 1 {
		return "", fmt.Errorf("convertRosewoodToCSV() --> invalid input")
	}

	titleHasNotBeenPrinted := true
	headerHasBeenPrinted := false
	rowNum := 1
	result := ""

	// attempt to obtain the number of columns of the table
	maxLength := 0
	for _, l := range lines {
		pieces := strings.Split(l, "|")
		if len(pieces) > maxLength {
			maxLength = len(pieces) - 1
		}
	}
	if maxLength < 1 {
		return "", fmt.Errorf("convertRosewoodToCSV() --> empty table given")
	}
	lengthAsString := strconv.Itoa(maxLength)

	// for every line in the table...
	maxColumns := 0
	maxRows := 0
	for _, l := range lines {

		// set a table number
		numAsStr := strconv.Itoa(num)

		trimmedLine := strings.TrimSpace(l)

		if trimmedLine == "" {
			continue
		}

		if trimmedLine == "---" {
			continue
		}

		if titleHasNotBeenPrinted {
			result += ":scaffolding-table-title-start:"
			result += "Table " + numAsStr + ": " + trimmedLine + "\n"
			result += ":scaffolding-table-title-end:"
			result += ":scaffolding-table-start-" + numAsStr + ":\n"
			result += ":scaffolding-column-table-" + numAsStr + ".A-len-" + lengthAsString + ":\n"
			titleHasNotBeenPrinted = false
			continue
		} else if !headerHasBeenPrinted {
			headerHasBeenPrinted = true
		} else {
			rowNum = 2
		}

		// generate a row number, for purposes of styling
		rowNumAsString := strconv.Itoa(rowNum)

		// Rosewood instructions are exactly one piece, so check for 2+
		pieces := strings.Split(l, "|")
		if len(pieces) < 2 {
			continue
		}

		// set a starting letter, ISO standard suggests A
		startingLetter := 65

		cleanedLine := ""
		for i, p := range pieces {

			// sometimes for the first element, there is a
			// Rosewood "  " prefix, so this needs to re-append
			// it to the trimmed string
			prefix := ""
			if i == 0 && strings.HasPrefix(p, "  ") && PrintAsCSV {
				prefix = "  "
			} else if i == 0 && strings.HasPrefix(p, "  ") {
				prefix = ":scaffolding-odt-space:"
			}

			cleanedString := strings.TrimSpace(p)

			if cleanedString == "" {
				continue
			}

			if strings.Contains(cleanedString, ",") {
				cleanedString = strings.Replace(cleanedString, ",", " ", -1)
			}

			letterStr := string(byte(startingLetter))
			cellStartStyle := ":scaffolding-cell-start-table-" + numAsStr + "." + letterStr + rowNumAsString + ":"

			if i == 0 {
				cleanedLine = cellStartStyle + prefix + cleanedString + ":scaffolding-cell-end: "
			} else {
				cleanedLine += cellStartStyle + cleanedString + ":scaffolding-cell-end: "
			}

			startingLetter++

			if i > maxColumns {
				maxColumns = i
			}
		}

		result += ":scaffolding-row-start:" + cleanedLine + ":scaffolding-row-end:\n"

		maxRows++
	}

	maxColumnsAsString := strconv.Itoa(maxColumns + 1)
	maxRowsAsString := strconv.Itoa(maxRows)

	result += ":scaffolding-table-cols-" + maxColumnsAsString + "-rows-" + maxRowsAsString + ":\n"
	result += ":scaffolding-table-end:\n"

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
	// Extract styles from the scaffolding elements
	//

	extractedStyles, err := obtainStylesFromScaffolding(data)
	if err != nil {
		return err
	}

	//
	// Append the new document styles
	//

	newContentXML := strings.Replace(odt.content, "<office:automatic-styles/>",
		"<office:automatic-styles>"+

			"<style:style style:name=\"P1\" style:family=\"paragraph\" style:parent-style-name=\"Standard\">"+
			"<style:paragraph-properties fo:break-before=\"page\"/></style:style>"+

			"<style:style style:name=\"P2\" style:family=\"paragraph\" style:parent-style-name=\"Footer\">"+
			"<style:paragraph-properties fo:text-align=\"end\" style:justify-single-word=\"false\"/>"+
			"</style:style>"+

			extractedStyles+

			"</office:automatic-styles>", -1)

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
		newContentXML += fixedLine
	}
	newContentXML += "<text:p text:style-name=\"Standard\"/>" + pieces[1]

	odt.content = newContentXML

	//
	// Handle table title elements
	//

	newContentXML = strings.Replace(odt.content, ":scaffolding-table-title-start:", "<text:p text:style-name=\"P1\">", -1)
	newContentXML = strings.Replace(newContentXML, ":scaffolding-table-title-end:", "</text:p>", -1)

	odt.content = newContentXML
	newContentXML = ""

	//
	// Handle start-of-column spacing
	//

	newContentXML = strings.Replace(odt.content, ":scaffolding-odt-space:", "<text:s text:c=\"6\"/>", -1)
	odt.content = newContentXML
	newContentXML = ""

	//
	// Handle ODT page breaks
	//

	newContentXML = strings.Replace(odt.content, ":scaffolding-page-break:",
		"<text:p text:style-name=\"Standard\"/><text:p text:style-name=\"Standard\"/>", -1)

	odt.content = newContentXML
	newContentXML = ""

	//
	// Convert scaffolding table elements to ODT elements
	//

	newContentXML = odt.content
	re1 := regexp.MustCompile(":scaffolding-table-start-(\\d+):")
	newContentXML = re1.ReplaceAllString(newContentXML, `<table:table table:name="Table$1" table:style-name="Table$1">`)
	newContentXML = strings.Replace(newContentXML, ":scaffolding-table-end:", "</table:table>", -1)

	re2 := regexp.MustCompile(":scaffolding-column-table-(\\d+\\.[A-Z]+)-len-(\\d+):")
	newContentXML = re2.ReplaceAllString(newContentXML, `<table:table-column table:style-name="Table$1" table:number-columns-repeated="$2" />`)

	re3 := regexp.MustCompile(":scaffolding-cell-start-table-(\\d+\\.[A-Z]+\\d+):")
	newContentXML = re3.ReplaceAllString(newContentXML, `<table:table-cell table:style-name="Table$1" office:value-type="string"><text:p text:style-name="Standard">`)

	newContentXML = strings.Replace(newContentXML, ":scaffolding-row-start:", "<table:table-row>", -1)
	newContentXML = strings.Replace(newContentXML, ":scaffolding-cell-end:", "</text:p></table:table-cell>", -1)
	newContentXML = strings.Replace(newContentXML, ":scaffolding-row-end:", "</table:table-row>", -1)

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

// obtainStylesFromScaffolding ... turn scaffolding elements into ODT styles
func obtainStylesFromScaffolding(data string) (string, error) {

	if data == "" {
		return "", fmt.Errorf("obtainStylesFromScaffolding(() --> invalid input")
	}

	regexStyles := regexp.MustCompile(":scaffolding-table-cols-(\\d+)-rows-(\\d+):")
	matches := regexStyles.FindAllStringSubmatch(data, -1)

	// ISO standard for the ODT suggests starting with the ASCII value of "A"
	startingLetter := 65

	styles := ""

	for i, matchArray := range matches {

		// skip if invalid match array
		if len(matchArray) != 3 {
			continue
		}

		tableNumStr := strconv.Itoa(i + 1)
		columns, err := strconv.ParseInt(matchArray[1], 10, 64)
		if err != nil {
			return "", err
		}

		// TODO: re-enable this at some future date for handling single-cell row contexts
		//rows, err := strconv.ParseInt(matchArray[2], 10, 64)
		//if err != nil {
		//	return "", err
		//}

		//
		// handle column styles
		//
		for j := 0; j < int(columns); j++ {

			letterStr := string(byte(startingLetter + j))

			styles += "<style:style style:name=\"Table" + tableNumStr + "." + letterStr + "1\" style:family=\"table-cell\">" +
				"<style:table-cell-properties fo:padding=\"0.049cm\" fo:border-left=\"0.05pt solid #000000\" " +
				"fo:border-right=\"0.05pt solid #000000\" fo:border-top=\"0.05pt solid #000000\" fo:border-bottom=\"0.05pt solid #000000\"/>" +
				"</style:style>"

			styles += "<style:style style:name=\"Table" + tableNumStr + "." + letterStr + "2\" style:family=\"table-cell\">" +
				"<style:table-cell-properties fo:padding=\"0.049cm\" fo:border-left=\"0.05pt solid #000000\" " +
				"fo:border-right=\"0.05pt solid #000000\" fo:border-top=\"0.05pt solid #000000\" fo:border-bottom=\"0.05pt solid #000000\"/>" +
				"</style:style>"
		}
	}

	return styles, nil
}
