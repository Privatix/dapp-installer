package util

import (
	"encoding/json"
	"os"
)

// Recursivly remove all data from directory
func ClearDirectory(path string) {
	if err := os.RemoveAll(path); err != nil {
		panic(err)
	}
}

// ReadJSONFile reads and parses a JSON file filling a given data instance.
func ReadJSONFile(name string, data interface{}) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(data)
}

// Version get installer versions
func Version() string {
	return "0.0.0.0"
}
