package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Usage() {
	fmt.Fprintln(os.Stderr, "Usage of ", os.Args[0], " [-h host:port] [-u url] [-f[ramed]] function [arg1 [arg2...]]:")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "\nFunctions:")
	fmt.Fprintln(os.Stderr, "  i32 Add(i32 A, i32 B)")
	fmt.Fprintln(os.Stderr)
	os.Exit(0)
}

type Helper interface {
	ReadArray() ([][]string, error)
	WriteArray(values [][]string) error
}

var newfunc map[string](func(string) (Helper, error))

func aaConvert(idir, odir, itype, otype string) {
	filepath.Walk(idir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}

		if info.Mode().IsRegular() && filepath.Ext(path) == "."+itype {
			ifile, err := newfunc[itype](path)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return nil
			}

			cfile, err := NewCsvHelper(odir + "/" + strings.Replace(info.Name(), "."+itype, "."+otype, -1))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return nil
			}

			data, err := ifile.ReadArray()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return nil
			}

			err = cfile.WriteArray(data)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
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
	}
	idir := flag.String("i", "", "-i dir")
	odir := flag.String("o", "", "-o dir")
	itype := flag.String("it", "xlsx", "-it type")
	otype := flag.String("ot", "csv", "-ot type")
	key := flag.String("k", "", "-k key")

	sheet := flag.String("s", "Sheet1", "-s sheet")
	SetSheetName(*sheet)

	flag.Usage = Usage
	flag.Parse()

	ctype := *itype + "2" + *otype
	switch {
	case ctype == "xlsx2csv" || ctype == "csv2xlsx":
		fallthrough
	case (ctype == "xlsx2lua" || ctype == "csv2lua") && *key == "":
		aaConvert(*idir, *odir, *itype, *otype)
	}

}
