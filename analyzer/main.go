package analyzer

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
func Run(iniFile1 string, verbose1 bool) error {
	verbose = verbose1
	a, err := newLogAnalyzerByIni(iniFile1)
	if err != nil {
		return err
	}

	defer a.close()
	if err := a.run(); err != nil {
		return err
	}

	/*
		if a.useDB {
			if _, err := a.runBlock(-1); err != nil {
				return err
			}
		}
	*/

	return nil
}
