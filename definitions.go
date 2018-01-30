package main

// Config holds user-provided and other settings
type Config struct {

	// CSV list of tables to convert
	tables string

	// location of said tables
	inputDir string

	// location to write the converted output
	outputDir string
}
