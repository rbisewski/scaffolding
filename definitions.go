package main

import (
	"archive/zip"
)

// Config holds user-provided and other settings
type Config struct {

	// CSV list of tables to convert
	tables string

	// location of said tables
	inputDir string

	// location to write the converted output
	outputDir string
}

// Odt ... Structure for handling ODT files
type Odt struct {
	files    []*zip.File
	content  string
	settings string
	styles   string
}

// CachedOdtTemplate ... structure for handling ODT files content replacement
type CachedOdtTemplate struct {
	zipReader *zip.ReadCloser
	content   string
	settings  string
	styles    string
}
