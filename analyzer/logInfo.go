package analyzer

import (
	"encoding/json"
	"io/ioutil"
)

func newLogInfoMap(jsonFile string) (*logInfoMap, error) {
	bytes, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return nil, err
	}
	var lm *logInfoMap
	if err := json.Unmarshal(bytes, &lm); err != nil {
		return nil, err
	}
	return lm, nil
}
