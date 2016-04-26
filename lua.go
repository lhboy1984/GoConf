package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/yuin/gopher-lua"
)

type LuaHelper struct {
	name string
}

func NewLuaHelper(name string) (Helper, error) {
	return &LuaHelper{name: name}, nil
}

func appendLuaHeader(t *lua.LTable, header map[string]int) (map[string]int, bool) {
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

func luaTableType(t *lua.LTable) (map[string]int, lua.LValueType) {
	var ktype lua.LValueType = lua.LTNil

	header := map[string]int{}

	t.ForEach(func(k, v lua.LValue) {
		if v.Type() != lua.LTTable {
			ktype = lua.LTNil
			return
		}

		var ok bool
		header, ok = appendLuaHeader(v.(*lua.LTable), header)
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

func LValueToString(l lua.LValue) string {
	switch l.Type() {
	case lua.LTNil:
		return "nil"
	case lua.LTBool:
		return l.(lua.LValue).String()
	case lua.LTNumber:
		return l.(lua.LValue).String()
	case lua.LTString:
		return "\"" + l.(lua.LValue).String() + "\""
	case lua.LTTable:
		strbuf := bytes.NewBuffer([]byte{})
		strbuf.WriteString("{")
		l.(*lua.LTable).ForEach(func(k, v lua.LValue) {
			if k.Type() == lua.LTNumber {
				strbuf.WriteString(LValueToString(v))
				strbuf.WriteString(",")
			}
		})
		strbuf.WriteString("}")
		return strbuf.String()
	default:
		panic("not support type")
	}
}

func (helper *LuaHelper) ReadArray() (value [][]string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	L := lua.NewState()
	L.OpenLibs()
	defer L.Close()
	L.DoFile(helper.name)

	basename := filepath.Base(helper.name)
	basename = basename[0 : len(basename)-4]

	t := L.GetGlobal(basename).(*lua.LTable)
	header, ltype := luaTableType(t)
	value = make([][]string, 1)
	value[0] = make([]string, len(header))
	for k, v := range header {
		value[0][v] = k
	}
	if ltype == lua.LTNil {
		return nil, errors.New("kind of table with key number and string not support")
	} else if ltype == lua.LTNumber {
		t.ForEach(func(k, v lua.LValue) {
			row := make([]string, len(header))
			v.(*lua.LTable).ForEach(func(kk, vv lua.LValue) {
				row[header[kk.String()]] = LValueToString(vv)
			})
			value = append(value, row)
		})
	} else {
		row := make([]string, len(header))
		t.ForEach(func(k, v lua.LValue) {
			row[0] = k.String()
			v.(*lua.LTable).ForEach(func(kk, vv lua.LValue) {
				row[header[kk.String()]] = LValueToString(vv)
			})
		})
		value = append(value, row)
	}

	return
}

func (helper *LuaHelper) WriteArray(values [][]string) error {
	basename := filepath.Base(helper.name)
	basename = basename[0 : len(basename)-4]

	f, err := os.OpenFile(helper.name, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(basename + "={\n")

	for i := 1; i < len(values); i++ {
		f.WriteString("\t{\n")
		for j := 0; j < len(values[0]) && j < len(values[i]); j++ {
			f.WriteString("\t\t" + values[0][j] + "=" + values[i][j] + ",\n")
		}
		f.WriteString("\t},\n")
	}
	f.WriteString("}\n")

	return nil
}

func luaValueToInterface(l lua.LValue) interface{} {
	switch l.Type() {
	case lua.LTNil:
		return nil
	case lua.LTBool:
		return (bool)(*(l.(*lua.LBool)))
	case lua.LTNumber:
		return (float64)(*(l.(*lua.LNumber)))
	case lua.LTString:
		return l.(lua.LString).String()
	case lua.LTTable:
		var result interface{}

		l.(*lua.LTable).ForEach(func(k, v lua.LValue) {
			if k.Type() == lua.LTNumber {
				if result == nil {
					result = []interface{}{}
				}
				if inst, ok := result.([]interface{}); ok {
					result = append(inst, luaValueToInterface(v))
				} else {
					panic("not support mix key with number and string")
				}
			} else {
				if result == nil {
					result = make(map[string]interface{})
				}

				if inst, ok := result.(map[string]interface{}); ok {
					inst[k.String()] = luaValueToInterface(v)
				} else {
					panic("not support mix key with number and string")
				}
			}
		})

		return result
	default:
		panic("not supported lua type")
	}
}

func (helper *LuaHelper) ReadMap(key string) (interface{}, error) {
	L := lua.NewState()
	L.OpenLibs()
	defer L.Close()
	err := L.DoFile(helper.name)
	if err != nil {
		return nil, err
	}

	basename := filepath.Base(helper.name)
	basename = basename[0 : len(basename)-4]

	fmt.Println(basename)

	v := L.GetGlobal(basename)
	return luaValueToInterface(v), nil
}

func interfaceToLuaFile(w *os.File, v interface{}, newline bool) {
	t := reflect.TypeOf(v)

	switch t.Kind() {
	case reflect.Bool:
		fallthrough
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		w.Write([]byte(fmt.Sprint(v)))
	case reflect.String:
		w.Write([]byte("\"" + fmt.Sprint(v) + "\""))
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		a, ok := v.([]interface{})
		if !ok {
			panic(t.String() + " not supported")
		}

		w.WriteString("{")
		if newline {
			w.WriteString("\n\t")
		}
		for _, vv := range a {
			interfaceToLuaFile(w, vv, false)
			w.WriteString(",")
			if newline {
				w.WriteString("\n\t")
			}
		}
		w.Seek(-1, 1)
		w.WriteString("}")
	case reflect.Map:
		m, ok := v.(map[string]interface{})
		if !ok {
			panic(t.String() + " not supported")
		}

		w.WriteString("{")
		if newline {
			w.WriteString("\n\t")
		}
		for kk, vv := range m {
			w.WriteString(kk + "=")
			interfaceToLuaFile(w, vv, false)
			w.WriteString(",")
			if newline {
				w.WriteString("\n\t")
			}
		}
		w.Seek(-1, 1)
		w.WriteString("}")
	default:
		fmt.Println("v type of kind is ", t.Kind(), "not supported")
	}
}

func (helper *LuaHelper) WriteMap(values interface{}) error {
	f, err := os.OpenFile(helper.name, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	basename := filepath.Base(helper.name)
	basename = basename[0 : len(basename)-4]
	f.WriteString(basename + "=")

	interfaceToLuaFile(f, values, true)

	return nil
}

func (helper *LuaHelper) WriteMapString(values map[string]map[string]interface{}) error {
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
	f.WriteString("}\n")

	return nil
}
