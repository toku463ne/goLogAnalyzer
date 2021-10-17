package csvdb

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

func ensureTestDir(testname string) (string, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	rootDir := fmt.Sprintf("%s/goCsvDb/%s", userDir, testname)
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		os.MkdirAll(rootDir, 0755)
	} else if os.IsExist(err) {
		os.RemoveAll(rootDir)
		os.MkdirAll(rootDir, 0755)
	}

	return rootDir, nil
}

func getGotExpErr(title string, got interface{}, exp interface{}) error {
	if got == exp {
		return nil
	}
	return errors.New(fmt.Sprintf("%s got=%v expected=%v", title, got, exp))
}
