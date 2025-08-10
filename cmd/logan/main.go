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

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const (
	usageStr = "usage: logan feed|history|groups|clean|test"
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
	msgFormats          []string
	timestampLayout     string
	maxBlocks           int
	blockSize           int
	keepPeriod          int64
	unitSecs            int64
	minMatchRate        float64
	termCountBorderRate float64
	termCountBorder     int
	line                string
	outDir              string
	_keywords           string
	keywords            []string
	_ignorewords        string
	ignorewords         []string
	_keyRegexes         string
	keyRegexes          []string
	_ignoreRegexes      string
	ignoRegexes         []string
	customLogGroups     []string
	N                   int
	B                   int
	cmd                 string
	useUtcTime          bool
	separators          string
	_flagSet            *flag.FlagSet
	loaded              bool
	ascOrder            bool
	analLogPath         string
	ignoreNumbers       bool
	minLastUpdate       int64
	minLogCount         int
	maxLogCount         int
	stdThreshold        float64
	minOccurrences      float64
	lastFileEpoch       int64
)

type config struct {
	DataDir             string   `yaml:"dataDir"`
	LogPath             string   `yaml:"logPath"`
	SearchRegex         []string `yaml:"searchRegex"`
	ExcludeRegex        []string `yaml:"excludeRegex"`
	LogFormat           string   `yaml:"logFormat"`
	MsgFormats          []string `yaml:"msgFormats"`
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
	KeyRegexes          []string `yaml:"keyRegexes"`
	IgnoreRegexes       []string `yaml:"ignoreRegexes"`
	CustomLogGroups     []string `yaml:"phrases"`
	UseUtcTime          bool     `yaml:"useUtcTime"`
	OutDir              string   `yaml:"outDir"`
	Separators          string   `yaml:"separators"`
	IgnoreNumbers       bool     `yaml:"ignoreNumbers"`
	DaysToShow          int64    `yaml:"daysToShow"`
	MinLogCount         int      `yaml:"minLogCount"`
	MaxLogCount         int      `yaml:"maxLogCount"`
	StdThreshold        float64  `yaml:"stdThreshold"`
	MinOccurrences      float64  `yaml:"minOccurrences"`
}

func setCommonFlag(fs *flag.FlagSet) {
	fs.BoolVar(&debug, "debug", false, "Enable debug mode")
	fs.BoolVar(&silent, "silent", false, "Enable silent mode")

	// Set up command line flags
	fs.StringVar(&dataDir, "d", "", "Path to the data directory")
	fs.StringVar(&configPath, "c", "", "Path to the configuration file")
	fs.StringVar(&logPath, "f", "", "Log file")
	fs.Int64Var(&unitSecs, "u", 0, "time unit in seconds")
	fs.Int64Var(&keepPeriod, "p", 0, "Number of unit secs to keep data")
	fs.StringVar(&searchString, "s", "", "Search string")
	fs.StringVar(&excludeString, "x", "", "Exclude string")
	fs.Float64Var(&minMatchRate, "m", 0, "It is considered 2 log lines 'match', if more than matchRate number of terms in a log line matches.")
	fs.Float64Var(&termCountBorderRate, "R", 0, "Words with less appearance will be replaced by '*'. The border is calculated by this rate.")
	fs.IntVar(&termCountBorder, "b", 0, "Words with less appearance than this number will be replaced by '*'. If 0, it will be calculated by termCountBorderRate")
	//fs.StringVar(&line, "line", "", "Log line to analyze")
	fs.StringVar(&_keywords, "keys", "", "List of terms to include in all phrases. Comma separated")
	fs.StringVar(&_ignorewords, "ignores", "", "List of terms to ignore in all phrases. Comma separated")
	fs.StringVar(&separators, "sep", "", "separators of words")
	fs.BoolVar(&ascOrder, "asc", false, "list up logGroups in ascending order or not")
	fs.BoolVar(&ignoreNumbers, "ignoreNumbers", false, "ignore all numbers")
	fs.StringVar(&analLogPath, "anallog", "", "File to output logs of this application")
	fs.Int64Var(&lastFileEpoch, "lastEpoch", 0, "last epoch of the log file")

}

func setNonFeedFlag(fs *flag.FlagSet) {
	setCommonFlag(fs)
	fs.BoolVar(&readOnly, "r", false, "Read only mode. Do not update data directory.")
}

func setOutFlag(fs *flag.FlagSet) {
	setNonFeedFlag(fs)
	fs.StringVar(&outDir, "o", "", "Output file")
	fs.IntVar(&N, "N", 0, "Number of items to output")
	fs.IntVar(&B, "B", 0, "Number of days to show the results")
	fs.IntVar(&minLogCount, "minCount", 0, "Minimum of logGroup size")
	fs.IntVar(&maxLogCount, "maxCount", 0, "Maximum of logGroup size")
	fs.Int64Var(&minLastUpdate, "lastepoch", 0, "minimum of the last updated epoch to show in output")
}

func setParseLineFlag(fs *flag.FlagSet) {
	fs.BoolVar(&debug, "debug", false, "Enable debug mode")
	fs.BoolVar(&silent, "silent", false, "Enable silent mode")
	fs.StringVar(&configPath, "c", "", "Path to the configuration file")
	fs.StringVar(&line, "line", "", "Log line to analyze")
	fs.StringVar(&analLogPath, "anallog", "", "File to output logs of this application")

}

func init() {
	loaded = false
}

func loadConfig(path string) error {
	logrus.WithField("path", path).Info("Loading configuration")

	// Read the YAML file
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.Fatalf("Error reading YAML file: %v", err)
	}

	// Load YAML content into a temporary map to process environment replacements
	var tempConfig map[string]interface{}
	if err := yaml.Unmarshal(yamlFile, &tempConfig); err != nil {
		logrus.Fatalf("Error unmarshalling YAML: %v", err)
	}

	// Recursively replace environment variables in the map
	replaceEnvVarsInMap(tempConfig)

	// Marshal modified map back to YAML to load it into the actual config struct
	modifiedYaml, err := yaml.Marshal(tempConfig)
	if err != nil {
		logrus.Fatalf("Error re-marshalling YAML: %v", err)
	}

	// Now unmarshal into the actual config struct
	var c config
	if err := yaml.Unmarshal(modifiedYaml, &c); err != nil {
		logrus.Fatalf("Error unmarshalling modified YAML: %v", err)
	}

	// Set defaults
	applyConfigValues(&c)

	return nil
}

// replaceEnvVarsInMap replaces {{ VAR }} placeholders with environment variables recursively in maps
func replaceEnvVarsInMap(data map[string]interface{}) {
	re := regexp.MustCompile(`\{\{\s*(\w+)\s*\}\}`)
	for key, value := range data {
		switch v := value.(type) {
		case string:
			// Replace environment variables in string values
			data[key] = re.ReplaceAllStringFunc(v, func(placeholder string) string {
				varName := re.FindStringSubmatch(placeholder)[1]
				return os.Getenv(varName)
			})
		case map[string]interface{}:
			// Recursively replace in nested maps
			replaceEnvVarsInMap(v)
		}
	}
}

// applyConfigValues applies default values for config fields if they are unset
func applyConfigValues(c *config) {
	if dataDir == "" {
		dataDir = c.DataDir
	}
	if logPath == "" {
		logPath = c.LogPath
	}
	if outDir == "" {
		outDir = c.OutDir
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
	if len(msgFormats) == 0 {
		msgFormats = c.MsgFormats
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
	if keyRegexes == nil {
		keyRegexes = c.KeyRegexes
	}
	if ignoRegexes == nil {
		ignoRegexes = c.IgnoreRegexes
	}
	ignoreNumbers = c.IgnoreNumbers
	if customLogGroups == nil {
		customLogGroups = c.CustomLogGroups
	}
	if separators == "" {
		separators = c.Separators
	}
	if minLastUpdate == 0 && c.DaysToShow > 0 {
		minLastUpdate = utils.GetNdaysBefore(int(c.DaysToShow))
	}
	if minLogCount == 0 {
		minLogCount = c.MinLogCount
	}
	if maxLogCount == 0 {
		maxLogCount = c.MaxLogCount
	}
	if stdThreshold == 0 {
		stdThreshold = c.StdThreshold
	}
	if minOccurrences == 0 {
		minOccurrences = c.MinOccurrences
	}

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

func checkCommonFlag() string {
	if logPath == "" {
		return "logPath is mandatory"
	}
	return ""
}

func checkTestFlag() string {
	if line == "" {
		return "line is mandatory"
	}
	if configPath == "" {
		return "configPath is mandatory"
	}
	return ""
}

func run() error {
	logrus.Debug("Starting")
	var err error
	var a *logan.Analyzer

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
	if _keywords != "" {
		keywords = strings.Split(_keywords, ",")
	}
	if _ignorewords != "" {
		ignorewords = strings.Split(_ignorewords, ",")
	}
	if _keyRegexes != "" {
		keyRegexes = strings.Split(_keyRegexes, ",")
	}
	if _ignoreRegexes != "" {
		ignoRegexes = strings.Split(_ignoreRegexes, ",")
	}
	if minLastUpdate == 0 && B > 0 {
		minLastUpdate = utils.GetNdaysBefore(B)
	}

	msg := ""
	testMode := false
	switch cmd {
	case "feed":
		msg = checkCommonFlag()
	case "test":
		msg = checkTestFlag()
		readOnly = true
		testMode = true
	}
	if msg != "" {
		fmt.Printf("%s for '%s' option\n", msg, cmd)
		return nil
	}

	tblDir := fmt.Sprintf("%s/config.json", dataDir)
	if utils.PathExist(tblDir) && !testMode {
		logrus.Infof("Loading config from %s\n", tblDir)
		a, err = logan.LoadAnalyzer(dataDir, logPath,
			termCountBorderRate,
			termCountBorder,
			minMatchRate,
			customLogGroups,
			readOnly, debug, testMode, ignoreNumbers)
	} else {
		conf := new(logan.AnalConfig)
		conf.DataDir = dataDir
		conf.LogPath = logPath
		conf.SearchRegex = searchRegex
		conf.ExludeRegex = excludeRegex
		conf.LogFormat = logFormat
		conf.MsgFormats = msgFormats
		conf.TimestampLayout = timestampLayout
		conf.KeepPeriod = keepPeriod
		conf.UnitSecs = unitSecs
		conf.MaxBlocks = maxBlocks
		conf.BlockSize = blockSize
		conf.TermCountBorderRate = termCountBorderRate
		conf.TermCountBorder = termCountBorder
		conf.MinMatchRate = minMatchRate
		conf.Keywords = keywords
		conf.Ignorewords = ignorewords
		conf.KeyRegexes = keyRegexes
		conf.IgnoreRegexes = ignoRegexes
		conf.CustomLogGroups = customLogGroups
		conf.UseUtcTime = useUtcTime
		conf.Separators = separators
		conf.IgnoreNumbers = ignoreNumbers

		a, err = logan.NewAnalyzer(conf,
			lastFileEpoch,
			readOnly, testMode)
	}
	if err != nil {
		return err
	}

	switch cmd {
	case "feed":
		err = a.Feed(0)
	case "history":
		err = a.OutputLogGroups(N, outDir, searchString, excludeString, minLastUpdate, minLogCount, maxLogCount, true, ascOrder)
	case "groups":
		err = a.OutputLogGroups(N, outDir, searchString, excludeString, minLastUpdate, minLogCount, maxLogCount, false, ascOrder)
	case "test":
		a.ParseLogLine(line)
	default:
		println(usageStr)
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}

func main() {

	if len(os.Args) < 2 {
		println(usageStr)
		return
	}

	cmd = os.Args[1]
	if !loaded {
		_flagSet = flag.NewFlagSet(fmt.Sprintf("logan %s", cmd), flag.ExitOnError)

		switch cmd {
		case "clean":
			setCommonFlag(_flagSet)
		case "feed":
			setCommonFlag(_flagSet)
		case "history":
			setOutFlag(_flagSet)
		case "groups":
			setOutFlag(_flagSet)
		case "test":
			setParseLineFlag(_flagSet)
		default:
			println(usageStr)
			return
		}
		if len(os.Args) < 3 {
			_flagSet.Usage()
			return
		}

		loaded = true
	}
	_flagSet.Parse(os.Args[2:])

	// log setting
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

	if analLogPath != "" {
		logFile, err := os.OpenFile(analLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Failed to open log file: %v\n", err)
			return
		}
		defer logFile.Close()
		logrus.SetOutput(logFile)
	}

	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 1024)
			n := runtime.Stack(buf, false)
			stackLines := strings.Split(string(buf[:n]), "\n")
			logrus.WithField("panic", r).Error("A panic occurred")
			logrus.SetFormatter(&logrus.TextFormatter{
				DisableTimestamp: true,
				DisableColors:    true,
				DisableQuote:     true,
			})
			for _, line := range stackLines {
				logrus.Errorln(line)
				fmt.Println(line)
			}

		}
	}()

	if err := run(); err != nil {
		fmt.Printf("%+v\n", err)
	} else {
		logrus.Debug("Finished successfully")
	}

}
