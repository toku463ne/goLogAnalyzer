package analyzer

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

func initTestDir(testname string) (string, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	rootDir := fmt.Sprintf("%s/loganal/%s", userDir, testname)
	if _, err := os.Stat(rootDir); err == nil {
		Clean(rootDir)
	}
	ensureDir(rootDir)

	return rootDir, nil
}
func removeTestDir(testname string) error {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	rootDir := fmt.Sprintf("%s/loganal/%s", userDir, testname)
	if _, err := os.Stat(rootDir); err == nil {
		os.RemoveAll(rootDir)
	}
	return nil
}

func getGotExpErr(title string, got interface{}, exp interface{}) error {
	if got == exp {
		return nil
	}
	return errors.New(fmt.Sprintf("%s got=%v expected=%v", title, got, exp))
}
