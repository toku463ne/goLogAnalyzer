package main

import (
	"bufio"
	"flag"
	"fmt"
	"goLogAnalyzer/internal/logan"
	"goLogAnalyzer/pkg/utils"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var (
	configPath          string
	debug               bool
	silent              bool
	readOnly            bool
	dataDir             string
	logPath             string
	searchString        string
	excludeString       string
	searchRegex         []string
	excludeRegex        []string
	logFormat           string
	timestampLayout     string
	maxBlocks           int
	blockSize           int
	keepPeriod          int64
	unitSecs            int64
	minMatchRate        float64
	termCountBorderRate float64
	termCountBorder     int
	line                string
	outputFile          string
	_keywords           string
	keywords            []string
	_ignorewords        string
	ignorewords         []string
	customLogGroups     []string
	N                   int
	cmd                 string
	useUtcTime          bool
	separators          string
)

type config struct {
	DataDir             string   `yaml:"dataDir"`
	LogPath             string   `yaml:"logPath"`
	SearchRegex         []string `yaml:"searchRegex"`
	ExcludeRegex        []string `yaml:"excludeRegex"`
	LogFormat           string   `yaml:"logFormat"`
	TimestampLayout     string   `yaml:"timestampLayout"`
	KeepPeriod          int64    `yaml:"keepPeriod"`
	UnitSecs            int64    `yaml:"unitSecs"`
	MaxBlocks           int      `yaml:"maxBlocks"`
	BlockSize           int      `yaml:"blockSize"`
	MinMatchRate        float64  `yaml:"minMatchRate"`
	TermCountBorderRate float64  `yaml:"termCountBorderRate"`
	TermCountBorder     int      `yaml:"termCountBorder"`
	Keywords            []string `yaml:"keywords"`
	Ignorewords         []string `yaml:"ignorewords"`
	CustomLogGroups     []string `yaml:"phrases"`
	UseUtcTime          bool     `yaml:"useUtcTime"`
	OutputFile          string   `yaml:"outputFile"`
	Separators          string   `yaml:"separators"`
}

func init() {
	// Set up command line flags
	flag.StringVar(&configPath, "c", "", "Path to the configuration file")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.BoolVar(&silent, "silent", false, "Enable silent mode")
	flag.BoolVar(&readOnly, "r", false, "Read only mode. Do not update data directory.")
	flag.StringVar(&dataDir, "d", "", "Path to the data directory")
	flag.StringVar(&logPath, "f", "", "Log file")

	flag.Int64Var(&unitSecs, "u", 0, "time unit in seconds")
	flag.Int64Var(&keepPeriod, "p", 0, "Number of unit secs to keep data")
	flag.StringVar(&searchString, "s", "", "Search string")
	flag.StringVar(&excludeString, "x", "", "Exclude string")
	flag.Float64Var(&minMatchRate, "m", 0, "It is considered 2 log lines 'match', if more than matchRate number of terms in a log line matches.")
	flag.Float64Var(&termCountBorderRate, "R", 0, "Words with less appearance will be replaced by '*'. The border is calculated by this rate.")
	flag.IntVar(&termCountBorder, "b", 0, "Words with less appearance than this number will be replaced by '*'. If 0, it will be calculated by termCountBorderRate")
	flag.StringVar(&line, "line", "", "Log line to analyze")

	flag.StringVar(&_keywords, "keys", "", "List of terms to include in all phrases. Comma separated")
	flag.StringVar(&_ignorewords, "ignores", "", "List of terms to ignore in all phrases. Comma separated")
	flag.StringVar(&separators, "sep", "", "separators of words")

	flag.StringVar(&outputFile, "o", "", "Output file")
	flag.IntVar(&N, "N", 100, "Number of top items")

	// Parse command line flags
	//flag.Parse()

	// Set up logging format
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})

	// Set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else if silent {
		logrus.SetLevel(logrus.ErrorLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

}

func loadConfig(path string) error {
	logrus.WithField("path", path).Info("Loading configuration")

	replaceEnvVars := func(content string) string {
		// Regex to find placeholders of the form {{ VAR }}
		re := regexp.MustCompile(`\{\{\s*(\w+)\s*\}\}`)
		return re.ReplaceAllStringFunc(content, func(placeholder string) string {
			// Extract the variable name from the placeholder
			varName := re.FindStringSubmatch(placeholder)[1]
			// Return the environment variable value or the original placeholder if not found
			return os.Getenv(varName)
		})
	}

	// Read the YAML file
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.Fatalf("Error reading YAML file: %v", err)
	}
	// Replace all placeholders with environment variable values
	yamlContent := replaceEnvVars(string(yamlFile))

	var c config
	err = yaml.Unmarshal([]byte(yamlContent), &c)
	if err != nil {
		logrus.Fatalf("Error unmarshalling YAML: %v", err)
	}
	if dataDir == "" {
		dataDir = c.DataDir
	}
	if logPath == "" {
		logPath = c.LogPath
	}
	if searchRegex == nil {
		searchRegex = c.SearchRegex
	}
	if excludeRegex == nil {
		excludeRegex = c.ExcludeRegex
	}
	if logFormat == "" {
		logFormat = c.LogFormat
	}
	if timestampLayout == "" {
		timestampLayout = c.TimestampLayout
	}
	if keepPeriod == 0 {
		keepPeriod = c.KeepPeriod
	}
	if unitSecs == 0 {
		unitSecs = c.UnitSecs
	}
	if blockSize == 0 {
		blockSize = c.BlockSize
	}
	if maxBlocks == 0 {
		maxBlocks = c.MaxBlocks
	}
	if minMatchRate == 0 {
		minMatchRate = c.MinMatchRate
	}
	if termCountBorderRate == 0 {
		termCountBorderRate = c.TermCountBorderRate
	}
	if termCountBorder == 0 {
		termCountBorder = c.TermCountBorder
	}
	if keywords == nil {
		keywords = c.Keywords
	}
	if ignorewords == nil {
		ignorewords = c.Ignorewords
	}
	if customLogGroups == nil {
		customLogGroups = c.CustomLogGroups
	}
	if outputFile == "" {
		outputFile = c.OutputFile
	}
	if separators == "" {
		separators = c.Separators
	}

	return nil
}

func clean() {
	// Check if the directory exists
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		fmt.Printf("Directory '%s' does not exist.\n", dataDir)
		return
	}

	if !silent {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Are you sure you want to remove the directory '%s'? (y/N): ", dataDir)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			return
		}

		response = strings.TrimSpace(response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Directory removal canceled.")
			return
		}
	}

	// Remove the directory
	err := os.RemoveAll(dataDir)
	if err != nil {
		fmt.Printf("Failed to remove directory '%s': %v\n", dataDir, err)
		return
	}

	fmt.Printf("Directory '%s' removed successfully.\n", dataDir)
}

func run() error {
	logrus.Debug("Starting")
	var err error
	var a *logan.Analyzer

	if len(os.Args) < 2 {
		return errors.New("usage: logan feed|history|groups|clean")
	}

	cmd = os.Args[1]
	flag.CommandLine.Parse(os.Args[2:])

	// Load configuration
	if configPath != "" {
		if err := loadConfig(configPath); err != nil {
			logrus.Info(fmt.Sprintf("Failed to load configuration: %s", configPath))
		}
	}
	if cmd == "clean" {
		clean()
		return nil
	}

	if len(searchRegex) == 0 && searchString != "" {
		searchRegex = []string{searchString}
	}
	if len(excludeRegex) == 0 && excludeString != "" {
		excludeRegex = []string{excludeString}
	}
	keywords = strings.Split(_keywords, ",")
	ignorewords = strings.Split(_ignorewords, ",")

	tblDir := fmt.Sprintf("%s/config.tbl.ini", dataDir)
	if utils.PathExist(tblDir) {
		logrus.Infof("Loading config from %s\n", tblDir)
		a, err = logan.LoadAnalyzer(dataDir, logPath,
			termCountBorderRate,
			termCountBorder,
			minMatchRate,
			customLogGroups,
			readOnly)
	} else {
		a, err = logan.NewAnalyzer(dataDir, logPath, logFormat, timestampLayout, useUtcTime,
			searchRegex, excludeRegex,
			maxBlocks, blockSize,
			keepPeriod, unitSecs,
			termCountBorderRate,
			termCountBorder,
			minMatchRate,
			keywords, ignorewords, customLogGroups,
			separators,
			readOnly)
	}
	if err != nil {
		return err
	}

	switch cmd {
	case "feed":
		err = a.Feed(0)
	case "history":
		err = a.OutputLogGroups(N, outputFile, true)
	case "groups":
		err = a.OutputLogGroups(N, outputFile, false)
	default:
		err = errors.New("must be one of feed|history|groups|clean")
	}
	if err != nil {
		return err
	}
	return nil
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 1024)
			n := runtime.Stack(buf, false)
			logrus.WithFields(logrus.Fields{
				"panic": r,
				"stack": string(buf[:n]),
			}).Error("A panic occurred")
		}
	}()

	if err := run(); err != nil {
		fmt.Printf("%+v\n", err)
	} else {
		logrus.Debug("Finished successfully")
	}
}
