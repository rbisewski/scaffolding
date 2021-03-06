/*~

# Summary
Scaffolding, a utility to convert rosewood tables into CSVs.

# Usage
Use the -h flag when running this program for basic usage information, and
consider reading the included doc.go for further details.

# Authors
Robert Bisewski <robert.bisewski@umanitoba.ca>

TODO: add a flag option to allow for overwriting, as right now the default is
      to refuse to overwrite an existing file

~*/

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

//
// Globals
//
var (
	// The version and build number are appended via Makefile, else default to
	// the below in the event that fails.
	Year    = "??"
	Version = "0.0"
	Build   = "unknown"

	// Whether or not to print the version + build information
	PrintVersionArgument = false

	// Whether to print CSV or ODT output
	PrintAsCSV = false

	// Default CSV output file name
	DefaultCSVOutputFilename = "rosewood.csv"

	// Default ODT output file name
	DefaultODTOutputFilename = "rosewood.odt"

	// Default templates directory
	DefaultTemplatesDir = "templates"
)

//
// Program Main
//
func main() {

	var config = Config{
		tables:    "",
		inputDir:  ".",
		outputDir: ".",
	}

	err := setupArguments(&config)
	if err != nil {
		fatal(err)
	}

	// if the version flag has been set to true, print the version
	// information and quit
	if PrintVersionArgument {
		fmt.Printf("Scaffolding ODT Generator v%s, Build: %s\n", Version, Build)
		os.Exit(0)
	}

	// by default, use the current directory if none is specified
	if config.inputDir == "" {
		config.inputDir = "."
	}
	if config.outputDir == "" {
		config.outputDir = "."
	}

	// validate input
	if err := validArgument(&config); err != nil {
		fmt.Println(usageMessage)
		fatal(err)
	}

	// split the table list into a string[]
	tables := strings.Split(config.tables, ",")

	tablePaths := make([]string, 0)
	for _, t := range tables {
		newPath := filepath.Join(config.inputDir, t)
		tablePaths = append(tablePaths, newPath)
	}

	if len(tablePaths) == 0 {
		fmt.Println("Error: unable to alloc mem for table filepaths.")
		os.Exit(1)
	}

	// create output paths to for where the files will be generated
	defaultOutputFilename := DefaultODTOutputFilename
	if PrintAsCSV {
		defaultOutputFilename = DefaultCSVOutputFilename
	}
	outputFilepath := filepath.Join(config.outputDir, defaultOutputFilename)

	// cycle thru every file
	contents := ""
	for i, path := range tablePaths {

		byteContents, err := ioutil.ReadFile(path)
		if err != nil {
			fatal(err)
		}

		rosewoodFileContents := string(byteContents)

		if rosewoodFileContents == "" {
			continue
		}

		rosewoodLines := strings.Split(rosewoodFileContents, "\n")

		csvOutput, err := convertRosewoodToCSV(rosewoodLines, i+1)
		if err != nil {
			fatal(err)
		}

		// If CSV then just newlines, else if ODT then do page-breaks after each table
		if PrintAsCSV {

			// plain-text CSV
			contents += csvOutput + "\n"

		} else if contents == "" {

			// first table of ODT
			contents += csvOutput

		} else {

			// subsequent tables of ODT
			contents += SCAFFOLDING_PAGE_BREAK + csvOutput
		}
	}

	// Print a plaint-text CSV file with the rosewood file contents
	if PrintAsCSV {
		err = ioutil.WriteFile(outputFilepath, []byte(contents), 0644)
		if err != nil {
			fatal(err)
		}
		os.Exit(0)
	}

	// Print an ODT file with the rosewood file contents
	odtTemplate, err := ReadOdtFile("odt_blank_template")
	newOdtFile := odtTemplate.New()
	newOdtFile.AppendStrings(contents)
	newOdtFile.Write(outputFilepath)
	os.Exit(0)
}

// Setup the program arguments
func setupArguments(config *Config) error {

	// input validation
	if config == nil {
		return fmt.Errorf("setupArguments() --> invalid config")
	}

	flag.Usage = func() {
		fmt.Println(usageMessage)
	}

	flag.BoolVar(&PrintAsCSV, "csv", false, "")
	flag.StringVar(&config.tables, "tables", "", "")
	flag.StringVar(&config.inputDir, "indir", ".", "")
	flag.StringVar(&config.outputDir, "outdir", ".", "")
	flag.BoolVar(&PrintVersionArgument, "version", false, "")

	flag.Parse()

	return nil
}

//validArgument returns an error if a necessary argument is missing
func validArgument(config *Config) error {

	if config.tables == "" {
		return fmt.Errorf("Invalid table names. Please enter a valid list of tables.")
	}

	// validation to ensure that inputDir actually corresponds to a valid path
	if config.inputDir == "" {
		return fmt.Errorf("Invalid input directory. Please enter a valid input directory.")
	}
	_, err := ioutil.ReadDir(config.inputDir)
	if err != nil {
		fatal(fmt.Errorf("Warning: the following is an invalid directory path --> " + config.inputDir))
	}

	// validation to ensure that outputDir actually corresponds to a valid path
	if config.outputDir == "" {
		return fmt.Errorf("Invalid output directory. Please enter a valid output directory.")
	}
	_, err = ioutil.ReadDir(config.outputDir)
	if err != nil {
		fatal(fmt.Errorf("Warning: the following is an invalid directory path --> " + config.outputDir))
	}

	return nil
}
