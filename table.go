package main

import (
	"errors"
	"regexp"
	"strconv"
	"util"
)

type column struct {
	Index int
	Type  string
	Name  string
	ExVal []string
}

type TableConfig struct {
	key  column
	cols map[string][]column
}

func matchone(str string, re map[string]*regexp.Regexp) ([]string, string) {
	for k, v := range re {
		rr := v.FindAllStringSubmatch(str, -1)
		if len(rr) == 1 {
			if k == "N" {
				return rr[0], rr[0][2]
			} else {
				return rr[0], k
			}
		}
	}

	return nil, ""
}

func (t *TableConfig) addColumn(c column) error {
	if col, ok := t.cols[c.Name]; ok {
		if col[0].Type != c.Type {
			return errors.New("duplicate column name with different types")
		} else if c.Type == "L" || c.Type == "S" || c.Type == "N" || c.Type == "B" {
			return errors.New("duplicate column name with simple type")
		} else {
			col = append(col, c)
			t.cols[c.Name] = col
			return nil
		}
	} else {
		t.cols[c.Name] = []column{c}
		return nil
	}
}

func (t *TableConfig) init(row []string) error {
	remap := make(map[string]*regexp.Regexp)

	atreg := "(@\\w+(\\.\\w*){0,1}){0,1}"

	remap["ATA"] = regexp.MustCompile("^([a-zA-Z][a-z0-9A-Z]*)_A_(\\d+)_T_([a-zA-Z][a-z0-9A-Z]*)_(\\d+)" + atreg + "$")
	remap["AT"] = regexp.MustCompile("^([a-zA-Z][a-z0-9A-Z]*)_A_(\\d+)_T_([a-zA-Z][a-z0-9A-Z]*)" + atreg + "$")
	remap["A"] = regexp.MustCompile("^([a-zA-Z][a-z0-9A-Z]*)_A_(\\d+)" + atreg + "$")
	remap["T"] = regexp.MustCompile("^([a-zA-Z][a-z0-9A-Z]*)_T_([a-zA-Z][a-z0-9A-Z]*)" + atreg + "$")
	remap["TA"] = regexp.MustCompile("^([a-zA-Z][a-z0-9A-Z]*)_T_([a-zA-Z][a-z0-9A-Z]*)_(\\d+)" + atreg + "$")
	remap["N"] = regexp.MustCompile("^([a-zA-Z][a-z0-9A-Z]*)_([LSNB])" + atreg + "$")

	exp := regexp.MustCompile("^([a-zA-Z][a-z0-9A-Z]*)_K([NS])$")

	t.key = column{Index: -1}
	t.cols = make(map[string][]column)
	for i := 0; i < len(row); i++ {
		ids := exp.FindAllStringSubmatch(row[i], -1)
		if len(ids) == 1 {
			if t.key.Index != -1 {
				return errors.New("multiple key not supported")
			}
			t.key.Index = i
			t.key.Type = ids[0][2]
			t.key.Name = ids[0][1]
			if err := t.addColumn(t.key); err != nil {
				return err
			}
		} else {
			rr, tt := matchone(row[i], remap)
			if rr != nil {
				if err := t.addColumn(column{Index: i, Type: tt, Name: rr[1], ExVal: rr}); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (t *TableConfig) ParseRow(row []string, includekey bool) map[string]string {
	result := make(map[string]string)
	for name, cols := range t.cols {
		if !includekey && name == t.key.Name {
			continue
		}

		switch cols[0].Type {
		case "B", "N":
			result[name] = row[cols[0].Index]
		case "S":
			result[name] = "\"" + row[cols[0].Index] + "\""
		case "L":
		case "A":
		case "AT":
		case "T":
		case "TA":
		case "ATA":
		default:
			panic("invalid type")
		}
	}
	for i, val := range t.cols {
		if val.Type == "L" {
			base := row.Cells[i].String()
			if base == "" {
				continue
			} else if len(val.ExVal) == 2 || val.ExVal[3] == "" {
				v.NVal = name + "." + row.Cells[i].String()
			} else {
				v.NVal = FormatAtString(row.Cells[i].String(), val.ExVal[3:])
			}
		} else if val.Type == "A" {
			if row.Cells[i].String() == "" {
				continue
			}
			idx, err := strconv.Atoi(val.ExVal[2])
			util.IfErrPanic(err)
			v.Reserve(idx)

			v.AVal[idx].NVal, v.AVal[idx].Type = RealValue(row.Cells[i], val.ExVal, 3)
		} else if val.Type == "AT" {
			if row.Cells[i].String() == "" {
				continue
			}
			idx, err := strconv.Atoi(val.ExVal[2])
			util.IfErrPanic(err)
			v.Reserve(idx)
			v.AVal[idx].Type = "T"
			cstr, ctype := RealValue(row.Cells[i], val.ExVal, 4)
			v.AVal[idx].TVal[val.ExVal[3]] = &ToLuaValue{Type: ctype, NVal: cstr}
		} else if val.Type == "T" {
			if row.Cells[i].String() == "" {
				continue
			}
			cstr, ctype := RealValue(row.Cells[i], val.ExVal, 3)
			v.TVal[val.ExVal[2]] = &ToLuaValue{Type: ctype, NVal: cstr}
		} else if val.Type == "TA" {
			if row.Cells[i].String() == "" {
				continue
			}
			tval, ok := v.TVal[val.ExVal[2]]
			if !ok {
				tval = &ToLuaValue{Type: "A"}
				v.TVal[val.ExVal[2]] = tval
			}
			idx, err := strconv.Atoi(val.ExVal[3])
			util.IfErrPanic(err)
			tval.Reserve(idx)
			tval.AVal[idx].NVal, tval.AVal[idx].Type = RealValue(row.Cells[i], val.ExVal, 4)
		} else if val.Type == "ATA" {
			if row.Cells[i].String() == "" {
				continue
			}
			idx, err := strconv.Atoi(val.ExVal[2])
			util.IfErrPanic(err)
			v.Reserve(idx)
			v.AVal[idx].Type = "T"

			tval, ok := v.AVal[idx].TVal[val.ExVal[3]]
			if !ok {
				tval = &ToLuaValue{Type: "T"}
				v.TVal[val.ExVal[3]] = tval
			}

			idx, err = strconv.Atoi(val.ExVal[4])
			util.IfErrPanic(err)
			tval.Reserve(idx)
			tval.AVal[idx].NVal, tval.AVal[idx].Type = RealValue(row.Cells[i], val.ExVal, 5)
		}

		v.Type = val.Type
	}

	return result
}

func (t *TableConfig) Parse(data [][]string) (interface{}, error) {
	if err := t.init(data[0]); err != nil {
		return nil, err
	}

	var result interface{}
	if t.key.Type == "N" {
		result = make([]interface{}, len(data)-1)
	} else {
		result = make(map[string]interface{})
	}

	for i := 1; i < len(data); i++ {
		if t.key.Type == "N" {
			row := t.ParseRow(data[i], true)
			result = append(result.([]interface{}), row)
		} else {
			result.(map[string]interface{})[data[0][t.key.Index]] = t.ParseRow(data[i], false)
		}
	}

	return nil, nil
}
