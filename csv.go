package main

import (
	"encoding/csv"
	"os"
)

type CsvHelper struct {
	value [][]string
	name  string
}

func NewCsvHelper(name string) (Helper, error) {
	x := &CsvHelper{name: name}

	var fd *os.File
	var err error
	if fd, err = os.OpenFile(name, os.O_RDONLY, os.ModePerm); err != nil {
		x.value, err = csv.NewReader(fd).ReadAll()
	}

	return x, err
}

func (x *CsvHelper) ReadArray() ([][]string, error) {
	return x.value, nil
}

func (x *CsvHelper) WriteArray(values [][]string) error {
	x.value = values
	fd, err := os.OpenFile(x.name, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	return csv.NewWriter(fd).WriteAll(values)
}
