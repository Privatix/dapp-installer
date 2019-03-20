package util

import (
	"encoding/json"
	"os"
	"reflect"
)

type merger struct {
	exceptions []string
}

// MergeJSONFile merges two json files.
func MergeJSONFile(dstFile, srcFile string, exceptions ...string) error {
	dstRead, err := os.Open(dstFile)
	if err != nil {
		return err
	}
	defer dstRead.Close()

	dstMap := make(map[string]interface{})
	json.NewDecoder(dstRead).Decode(&dstMap)

	srcRead, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer srcRead.Close()

	srcMap := make(map[string]interface{})
	json.NewDecoder(srcRead).Decode(&srcMap)

	m := &merger{exceptions: exceptions}
	m.merge(dstMap, srcMap)

	write, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer write.Close()

	return json.NewEncoder(write).Encode(dstMap)
}

func (m *merger) merge(dstMap, srcMap map[string]interface{}) {
	for key := range dstMap {
		m.mergeKey(key, dstMap[key], srcMap[key], dstMap)
	}
}

func (m *merger) mergeKey(key string, dst interface{}, src interface{},
	result map[string]interface{}) {
	if !reflect.DeepEqual(dst, src) {
		switch dst.(type) {
		case map[string]interface{}:
			if _, ok := src.(map[string]interface{}); ok {
				dstMap := dst.(map[string]interface{})
				srcMap := src.(map[string]interface{})
				for k := range dstMap {
					m.mergeKey(k, dstMap[k], srcMap[k], dstMap)
				}
			}
		default:
			if src != nil && !contains(m.exceptions, key) {
				result[key] = src
			}
		}
	}
}

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
