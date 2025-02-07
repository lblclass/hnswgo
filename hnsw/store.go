package hnsw

import (
	"encoding/gob"
	"encoding/json"
	"os"
)

func JsonStructLocalStore(val interface{}, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(val); err != nil {
		return err
	}
	return nil
}

func JsonReadStruct(filePath string) (*HNSW, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var res = HNSW{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func GobStructLocalStore(val interface{}, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(val); err != nil {
		return err
	}
	return nil
}

func GobReadStruct(filePath string) (*HNSW, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var res = HNSW{}
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}
