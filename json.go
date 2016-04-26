package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitly/go-simplejson"
	"github.com/yuin/gopher-lua"
)

type JsonHelper struct {
	name string
}

func NewJsonHelper(name string) (Helper, error) {
	return &JsonHelper{name: name}, nil
}

func appendHeader(t *lua.LTable, header map[string]int) (map[string]int, bool) {
	is := true
	t.ForEach(func(k, v lua.LValue) {
		if k.Type() != lua.LTString {
			is = false
			return
		}

		if _, ok := header[k.String()]; !ok {
			index := len(header)
			header[k.String()] = index
		}
	})
	return header, is
}

func tableType(t *lua.LTable) (map[string]int, lua.LValueType) {
	var ktype lua.LValueType = lua.LTNil

	header := map[string]int{}

	t.ForEach(func(k, v lua.LValue) {
		if v.Type() != lua.LTTable {
			ktype = lua.LTNil
			return
		}

		var ok bool
		header, ok = appendHeader(v.(*lua.LTable), header)
		if !ok {
			ktype = lua.LTNil
			return
		}

		if ktype == lua.LTNil {
			ktype = k.Type()
		} else if ktype != k.Type() {
			ktype = lua.LTNil
			return
		}

	})

	if ktype == lua.LTString {
		for k, v := range header {
			header[k] = v + 1
		}

		id := "ID"
		for {
			if _, ok := header[id]; ok {
				id += "_K"
			} else {
				header[id] = 0
			}
		}
	}

	return header, ktype
}

func jsonReadFromArray(jdata *simplejson.Json) ([][]string, error) {
	if arr, err := jdata.Array(); err == nil {
		header := map[string]int{}
		for _, v := range arr {
			if m, err := v.(*simplejson.Json).Map(); err == nil {
				for hk, _ := range m {
					if _, ok := header[hk]; !ok {
						index := len(header)
						header[hk] = index
					}
				}
			} else {
				return nil, err
			}
		}

		values := make([][]string, 1)
		values[0] = make([]string, len(header))
		for k, v := range header {
			values[0][v] = k
		}

		for _, v := range arr {
			row := make([]string, len(header))
			m, _ := v.(*simplejson.Json).Map()
			for kk, vv := range m {
				row[header[kk]] = fmt.Sprint(vv)
			}
			values = append(values, row)
		}

		return values, nil
	} else {
		return nil, os.ErrInvalid
	}
}

func jsonReadFromMap(jdata *simplejson.Json) ([][]string, error) {
	if m, err := jdata.Map(); err == nil {
		header := map[string]int{}
		for _, v := range m {
			if mm, err := v.(*simplejson.Json).Map(); err == nil {
				for hk, _ := range mm {
					if _, ok := header[hk]; !ok {
						index := len(header)
						header[hk] = index
					}
				}
			} else {
				return nil, err
			}
		}

		values := make([][]string, 1)
		values[0] = make([]string, len(header)+1)
		values[0][0] = "ID"
		for {
			if _, ok := header[values[0][0]]; ok {
				values[0][0] = "ID" + "_K"
			} else {
				break
			}
		}
		for k, v := range header {
			values[0][v+1] = k
		}

		for k, v := range m {
			row := make([]string, len(header)+1)
			row[0] = k
			mm, _ := v.(*simplejson.Json).Map()
			for kk, vv := range mm {
				row[header[kk]] = fmt.Sprint(vv)
			}
			values = append(values, row)
		}

		return values, nil
	} else {
		return nil, os.ErrInvalid
	}
}

func (helper *JsonHelper) ReadArray() ([][]string, error) {
	f, err := os.OpenFile(helper.name, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	jdata, err := simplejson.NewFromReader(f)
	if err != nil {
		return nil, err
	}

	if value, err := jsonReadFromArray(jdata); err != os.ErrInvalid {
		return value, err
	} else if value, err := jsonReadFromMap(jdata); err != os.ErrInvalid {
		return value, err
	} else {
		return nil, errors.New("kind of table with key number and string not support")
	}
}

func (helper *JsonHelper) WriteArray(values [][]string) error {
	basename := filepath.Base(helper.name)
	basename = basename[0 : len(basename)-5]

	f, err := os.OpenFile(helper.name, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(basename + "[")

	for i := 1; i < len(values); i++ {
		f.WriteString("{")
		for j := 0; j < len(values[0]) && j < len(values[i]); j++ {
			f.WriteString(values[0][j] + ":" + values[i][j])
			if j != len(values[i])-1 {
				f.WriteString(",")
			}
		}
		if i != len(values)-1 {
			f.WriteString("},")
		} else {
			f.WriteString("}")
		}

	}
	f.WriteString("]")

	return nil
}

func (helper *JsonHelper) ReadMap(key string) (map[string]map[string]interface{}, error) {
	panic("lua not support readmap")
}

func (helper *JsonHelper) WriteMap(values map[string]map[string]interface{}) error {
	basename := filepath.Base(helper.name)
	basename = basename[0 : len(basename)-4]

	f, err := os.OpenFile(helper.name, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(basename + "={\n")
	for k, v := range values {
		f.WriteString("\t" + k + "={\n")
		for kk, vv := range v {
			f.WriteString("\t\t" + kk + "=")
			f.WriteString(fmt.Sprint(vv))
			f.WriteString(",\n")
		}
		f.WriteString("\t},\n")
	}

	return nil
}

func (helper *JsonHelper) WriteMapString(values map[string]map[string]interface{}) error {
	basename := filepath.Base(helper.name)
	basename = basename[0 : len(basename)-4]

	f, err := os.OpenFile(helper.name, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(basename + "={\n")
	for k, v := range values {
		f.WriteString("\t" + k + "={\n")
		for kk, vv := range v {
			f.WriteString("\t\t" + kk + "=")
			f.WriteString(fmt.Sprint(vv))
			f.WriteString(",\n")
		}
		f.WriteString("\t},\n")
	}

	return nil
}
