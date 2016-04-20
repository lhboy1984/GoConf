package main

import (
	"errors"

	"github.com/tealeg/xlsx"
)

type XlsxHelper struct {
	file *xlsx.File
	name string
}

var _sheet string = ""

func SetSheetName(s string) {
	_sheet = s
}

func NewXlsxHelper(name string) (Helper, error) {
	x := &XlsxHelper{name: name}

	var err error
	x.file, err = xlsx.OpenFile(name)
	if err != nil {
		x.file, err = xlsx.NewFile(), nil
	}
	return x, err
}

func (x *XlsxHelper) ReadMap(key string) (map[string]map[string]string, error) {
	if s, ok := x.file.Sheet[_sheet]; ok {
		header, err := x.HeaderIndex()
		if err != nil {
			return nil, err
		}

		if kindex, exist := header[key]; exist {
			result := make(map[string]map[string]string)
			for i := 1; i < len(s.Rows); i++ {
				kval := s.Rows[i].Cells[kindex].String()
				result[kval] = make(map[string]string)

				for j := 0; j < len(s.Rows[i].Cells); j++ {
					if j != kindex {
						result[kval][s.Rows[0].Cells[j].String()] = s.Rows[i].Cells[j].String()
					}
				}
			}
			return result, nil
		} else {
			return nil, errors.New("sheet: " + _sheet + " not has key " + key)
		}
	} else {
		return nil, errors.New("sheet: " + _sheet + " not exists")
	}
}

func (x *XlsxHelper) ReadArray() ([][]string, error) {
	if s, ok := x.file.Sheet[_sheet]; ok {
		var result [][]string = make([][]string, len(s.Rows))
		for i := 0; i < len(s.Rows); i++ {
			result[i] = make([]string, len(s.Rows[i].Cells))
			for j := 0; j < len(s.Rows[i].Cells); j++ {
				result[i][j] = s.Rows[i].Cells[j].String()
			}
		}
		return result, nil
	} else {
		return nil, errors.New("sheet: " + _sheet + " not exists")
	}
}

func (x *XlsxHelper) WriteArray(values [][]string) error {
	s, ok := x.file.Sheet[_sheet]
	if !ok {
		s = x.file.AddSheet(_sheet)
	} else {
		s.Rows = nil
		s.Cols = nil
		s.MaxCol = 0
		s.MaxRow = 0
		s.Hidden = false
		s.SheetViews = nil
	}

	for i := 0; i < len(values); i++ {
		row := s.AddRow()
		for j := 0; j < len(values[i]); j++ {
			cell := row.AddCell()
			cell.SetString(values[i][j])
		}
	}

	x.file.Save(x.name)
	return nil
}

func (x *XlsxHelper) HeaderIndex() (map[string]int, error) {
	if s, ok := x.file.Sheet[_sheet]; ok {
		var header map[string]int = map[string]int{}
		for i := 0; i < len(s.Rows[0].Cells); i++ {
			h := s.Rows[0].Cells[i].String()
			if _, exist := header[h]; exist {
				return nil, errors.New("sheet: " + _sheet + " has duplicate header " + h)
			}
			header[h] = i
		}

		return header, nil
	} else {
		return nil, errors.New("sheet: " + _sheet + " not exists")
	}
}
