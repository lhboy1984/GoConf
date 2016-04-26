package main

import (
	"encoding/csv"
	"errors"
	"os"
)

type CsvHelper struct {
	name string
}

func NewCsvHelper(name string) (Helper, error) {
	x := &CsvHelper{name: name}
	return x, nil
}

func (x *CsvHelper) ReadArray() ([][]string, error) {
	if fd, err := os.OpenFile(x.name, os.O_RDONLY, os.ModePerm); err == nil {
		defer fd.Close()
		return csv.NewReader(fd).ReadAll()
	} else {
		return nil, err
	}
}

func (x *CsvHelper) WriteArray(values [][]string) error {
	fd, err := os.OpenFile(x.name, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer fd.Close()
	return csv.NewWriter(fd).WriteAll(values)
}

func (x *CsvHelper) ReadMap(key string) (interface{}, error) {
	values, err := x.ReadArray()
	if err != nil {
		return nil, err
	}

	header := map[string]int{}
	for k, v := range values[0] {
		header[v] = k
	}

	if kindex, exist := header[key]; exist {
		result := make(map[string]map[string]interface{})
		for i := 1; i < len(values[1:]); i++ {
			kval := values[i][kindex]
			result[kval] = make(map[string]interface{})

			for j := 0; j < len(values[i]); j++ {
				if j != kindex {
					result[kval][values[0][j]] = values[i][j]
				}
			}
		}
		return result, nil
	} else {
		return nil, errors.New("sheet: " + _sheet + " not has key " + key)
	}
}

func (x *CsvHelper) WriteMap(values interface{}) error {
	panic("xlsx not supported writemap")
}

func (x *CsvHelper) WriteMapString(values map[string]map[string]interface{}) error {
	panic("xlsx not supported WriteMapString")
}
