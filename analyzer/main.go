package analyzer

import "fmt"

// Destroy ... Drop all tables
func Destroy(iniFile1 string) error {
	a, err := newLogAnalyzerByIni(iniFile1, false)
	if err != nil {
		return err
	}
	if err := a.destroy(); err != nil {
		return err
	}
	return nil
}

// Run ...
func Run(iniFile1 string, debug bool, pathRegex string) error {
	var a *logAnalyzer
	var err error

	if iniFile1 != "" {
		a, err = newLogAnalyzerByIni(iniFile1, debug)
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

// Frq ... Get Closed Frequent Item Sets by DCI Closed Algorithm
func Frq(path string,
	minSupport int,
	filterRe, xFilterRe string) error {
	a := newFileAnalyzer(path, filterRe, xFilterRe)
	if err := a.tokenizeFile(); err != nil {
		return err
	}
	matrix := tran2BitMatrix(a.trans, a.items)
	dci, err := newDCIClosed(matrix, minSupport, true)
	if err != nil {
		return err
	}
	if err = dci.run(); err != nil {
		return err
	}
	dci.output(a.items, len(a.trans.getSlice()), a.trans.mask)

	return nil
}
