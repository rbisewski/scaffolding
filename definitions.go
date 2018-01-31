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
	headers  map[string]string
	footers  map[string]string
}

// ReplaceOdt ... structure for handling ODT files content replacement
type ReplaceOdt struct {
	zipReader *zip.ReadCloser
	content   string
	settings  string
	styles    string
	headers   map[string]string
	footers   map[string]string
}
