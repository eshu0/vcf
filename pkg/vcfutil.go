package main

import (
	"archive/tar"
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"os"
	"path/filepath"

	sl "github.com/eshu0/simplelogger"
)

type VMUtil struct{
	Logger          	 	sl.ISimpleLogger
}

type Fthing struct {
	Name string
	Body string
}

func visit(files *[]Fthing) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		ft := Fthing{}
		ft.Name = path
		dat, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Println("visit ERROR")
			fmt.Println(err)
			return nil
		} else {
			ft.Body = string(dat)
			*files = append(*files, ft)
			return nil
		}
	}
}

func (util VMUtil) TarAndZipFolder(root string, outputfilename string) {

	var files = []Fthing{}

	err := filepath.Walk(root, visit(&files))
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		fmt.Println(file)
	}

	// Create and add some files to the archive.
	//var buf bytes.Buffer
	f, err := os.Create(outputfilename)
	if err != nil {
		slog.LogErrorf("%s", err.Error())
		return
	}
	f.Close()

	file, err := os.OpenFile(outputfilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.LogErrorf("%s", err.Error())
	}

	tw1 := tar.NewWriter(file)

	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}

		if err := tw1.WriteHeader(hdr); err != nil {
			util.Logger.LogErrorf("%s", err.Error())
		}

		if _, err := tw1.Write([]byte(file.Body)); err != nil {
			util.Logger.LogErrorf("%s", err.Error())
		}
		tw1.Flush()
	}

	if err := tw1.Close(); err != nil {
		util.Logger.LogErrorf("%s", err.Error())
	}

}
