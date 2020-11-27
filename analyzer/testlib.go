package analyzer

import (
	"fmt"
	"os"
)

func ensureTestDir(testname string) (string, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	rootDir := fmt.Sprintf("%s/loganal/%s", userDir, testname)
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		os.Mkdir(rootDir, 0755)
	}
	return rootDir, nil
}
