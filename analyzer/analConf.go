package analyzer

func NewAnalConf(rootDir string) *AnalConf {
	c := new(AnalConf)
	c.RootDir = rootDir
	return c
}
