package main

import (
	"fmt"
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
func ReadOdtFile(path string) (*ReplaceOdt, error) {

	// TODO: implement the below logic
	/*
		reader, err := zipOpenReader(path)
		if err != nil {
			return nil, err
		}

		content, err := readText(reader.File)
		if err != nil {
			return nil, err
		}

		settings, err := readSettings(reader.File)
		if err != nil {
			return nil, err
		}

		styles, err := readStyles(reader.File)
		if err != nil {
			return nil, err
		}
	*/

	return &ReplaceOdt{}, nil
}
