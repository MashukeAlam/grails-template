package helpers

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type ModelsJSON map[string][]Field

const jsonFilePath = "models.json"

func ReadModelsFromJSON() (ModelsJSON, error) {
	file, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return ModelsJSON{}, nil // If the file doesn't exist, return an empty map
		}
		return nil, err
	}

	var models ModelsJSON
	err = json.Unmarshal(file, &models)
	if err != nil {
		return nil, err
	}
	return models, nil
}

func WriteModelsToJSON(models ModelsJSON) error {
	file, err := json.MarshalIndent(models, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(jsonFilePath, file, 0644)
	if err != nil {
		return err
	}
	return nil
}

func GetModelNames() ([]string, error) {
	models, err := ReadModelsFromJSON()
	if err != nil {
		return nil, err
	}

	var modelNames []string
	for modelName := range models {
		modelNames = append(modelNames, modelName)
	}
	return modelNames, nil
}
