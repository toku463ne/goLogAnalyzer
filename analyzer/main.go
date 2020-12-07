package analyzer

import "fmt"

// Destroy ... Drop all tables
func Destroy(iniFile1 string) error {
	a, err := newLogAnalyzerByIni(iniFile1)
	if err != nil {
		return err
	}
	if err := a.destroy(); err != nil {
		return err
	}
	return nil
}

// Run ...
func Run(iniFile1 string, verbose1 bool, pathRegex string) error {
	verbose = verbose1
	var a *logAnalyzer
	var err error

	if iniFile1 != "" {
		a, err = newLogAnalyzerByIni(iniFile1)
		logInfo(fmt.Sprintf("Starting goLogAnalyzer with ini=%s", iniFile1))
	}
	if pathRegex != "" {
		a, err = newLogAnalyzerByDefaults(pathRegex)
		logInfo(fmt.Sprintf("Starting goLogAnalyzer for target=%s", pathRegex))
	}
	if err != nil {
		return err
	}

	if a.useDB {
		if err := a.loadDB(); err != nil {
			return err
		}
	}

	defer a.close()
	if err := a.run(0); err != nil {
		return err
	}
	logInfo("Finished goLogAnalyzer")

	return nil
}
