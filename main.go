package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Usage() {
	fmt.Fprintln(os.Stderr, "Usage of ", os.Args[0], " [-i dir] [-o dir] [-it type to convert]] [-ot type convert to] [-k key] [-s sheet]")
	flag.PrintDefaults()
	os.Exit(0)
}

type Helper interface {
	ReadArray() ([][]string, error)
	WriteArray(values [][]string) error
	ReadMap(key string) (interface{}, error)
	WriteMap(values interface{}) error
	WriteMapString(values map[string]map[string]interface{}) error
}

var newfunc map[string](func(string) (Helper, error))

func aaConvert(idir, odir, itype, otype string) {
	filepath.Walk(idir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
			return err
		}

		if info.Mode().IsRegular() && filepath.Ext(path) == "."+itype {
			log.Println(path, itype, otype)
			ifile, err := newfunc[itype](path)
			if err != nil {
				log.Println(path, err)
				return nil
			}

			cfile, err := newfunc[otype](odir + "/" + strings.Replace(info.Name(), "."+itype, "."+otype, -1))
			if err != nil {
				log.Println(path, err)
				return nil
			}

			data, err := ifile.ReadArray()
			if err != nil {
				log.Println(path, err)
				return nil
			}

			err = cfile.WriteArray(data)
			if err != nil {
				log.Println(path, err)
				return nil
			}
		}

		return nil
	})
}

func mmConvert(idir, odir, itype, otype, key string) {
	filepath.Walk(idir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
			return err
		}

		if info.Mode().IsRegular() && filepath.Ext(path) == "."+itype {
			log.Println(path, itype, otype)
			ifile, err := newfunc[itype](path)
			if err != nil {
				log.Println(path, err)
				return nil
			}

			cfile, err := newfunc[otype](odir + "/" + strings.Replace(info.Name(), "."+itype, "."+otype, -1))
			if err != nil {
				log.Println(path, err)
				return nil
			}

			data, err := ifile.ReadMap(key)
			if err != nil {
				log.Println(path, err)
				return nil
			}

			err = cfile.WriteMap(data)
			if err != nil {
				log.Println(path, err)
				return nil
			}
		}

		return nil
	})
}

func mmStringConvert(idir, odir, itype, otype, key string) {
	filepath.Walk(idir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
			return err
		}

		if info.Mode().IsRegular() && filepath.Ext(path) == "."+itype {
			log.Println(path, itype, otype)
			ifile, err := newfunc[itype](path)
			if err != nil {
				log.Println(path, err)
				return nil
			}

			cfile, err := newfunc[otype](odir + "/" + strings.Replace(info.Name(), "."+itype, "."+otype, -1))
			if err != nil {
				log.Println(path, err)
				return nil
			}

			data, err := ifile.ReadMap(key)
			if err != nil {
				log.Println(path, err)
				return nil
			}

			err = cfile.WriteMapString(data.(map[string]map[string]interface{}))
			if err != nil {
				log.Println(path, err)
				return nil
			}
		}

		return nil
	})
}

func main() {
	newfunc = map[string](func(string) (Helper, error)){
		"xlsx": NewXlsxHelper,
		"csv":  NewCsvHelper,
		"lua":  NewLuaHelper,
		"json": NewJsonHelper,
	}
	idir := flag.String("i", "test", "-i dir")
	odir := flag.String("o", "test", "-o dir")
	itype := flag.String("it", "json", "-it type")
	otype := flag.String("ot", "lua", "-ot type")
	key := flag.String("k", "ID", "-k key")

	sheet := flag.String("s", "Sheet1", "-s sheet")
	SetSheetName(*sheet)

	flag.Usage = Usage
	flag.Parse()

	ctype := *itype + "2" + *otype
	switch {
	case ctype == "xlsx2csv" || ctype == "csv2xlsx":
		fallthrough
	case ctype == "lua2xlsx" || ctype == "lua2csv":
		fallthrough
	case ctype == "json2xlsx" || ctype == "json2csv":
		fallthrough
	case (ctype == "xlsx2json" || ctype == "csv2json") && *key == "":
		fallthrough
	case (ctype == "xlsx2lua" || ctype == "csv2lua") && *key == "":
		aaConvert(*idir, *odir, *itype, *otype)
	case (ctype == "xlsx2json" || ctype == "csv2json"):
		fallthrough
	case (ctype == "xlsx2lua" || ctype == "csv2lua"):
		mmStringConvert(*idir, *odir, *itype, *otype, *key)
	case (ctype == "json2lua" || ctype == "lua2json"):
		mmConvert(*idir, *odir, *itype, *otype, *key)
	default:
		log.Fatal("not support convert ", ctype)
	}

}
