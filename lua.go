package main

import (
	"errors"
	"path/filepath"

	"github.com/yuin/gopher-lua"
)

type LuaHelper struct {
	name string
}

func NewLuaHelper(name string) (Helper, error) {
	return nil, nil
}

func tableType(t *lua.LTable) lua.LValueType {
	var ltype lua.LValueType = lua.LTNil
	t.ForEach(func(k, v lua.LValue) {
		if ltype == lua.LTNil {
			ltype = k.Type()
		} else if ltype != k.Type() {
			ltype = lua.LTNil
			return
		}
	})

	return ltype
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
	//var header map[string]int = map[string]int{}
	ltype := tableType(t)
	if ltype == lua.LTNil {
		return nil, errors.New("table with key number and string not support")
	} else if ltype == lua.LTNumber {
		t.ForEach(func(k, v lua.LValue) {

		})
	} else {
		t.ForEach(func(k, v lua.LValue) {

		})
	}

	return
}

func (helper *LuaHelper) WriteArray(values [][]string) error {
	basename := filepath.Base(helper.name)
	basename = basename[0 : len(basename)-4]

	str := basename + "={\n"
	for i := 1; i < len(values); i++ {
		str += "\t{\n"
		for j := 0; j < len(values[0]) && j < len(values[i]); j++ {
			str += "\t\t" + values[0][j] + "=" + values[i][j] + ",\n"
		}
		str += "\t},\n"
	}
	str += "}\n"

	return nil
}
