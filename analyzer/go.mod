module github.com/toku463ne/goLogAnalyzer/analyzer

go 1.16

replace github.com/toku463ne/goLogAnalyzer/analyzer/csvdb => ../csvdb

require (
	github.com/pkg/errors v0.9.1
	github.com/toku463ne/goLogAnalyzer/csvdb v0.0.0-20211017125145-cafef12f73e3
)
