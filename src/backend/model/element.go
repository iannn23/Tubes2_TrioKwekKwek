package model

import (
	"encoding/json"
	"os"
)

type Element struct {
	Name    string     `json:"name"`
	Recipes [][]string `json:"recipes"`
}

func LoadElements(path string) (map[string]*Element, error) {
	var list []Element
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&list); err != nil {
		return nil, err
	}
	result := make(map[string]*Element)
	for _, e := range list {
		copy := e
		result[e.Name] = &copy
	}
	return result, nil
}
