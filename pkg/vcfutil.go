package vcf

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type VMUtil struct {
	Session VCFSession
}

type Fthing struct {
	Name string
	Body string
}

func visit(files *[]Fthing, util *VMUtil) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			util.Session.Logger.LogErrorE("Visit", err)
		}
		ft := Fthing{}
		ft.Name = path
		dat, err := ioutil.ReadFile(path)
		if err != nil {
			util.Session.Logger.LogErrorE("Visit", err)
			return nil
		} else {
			ft.Body = string(dat)
			*files = append(*files, ft)
			return nil
		}
	}
}

func (util *VMUtil) TarAndZipFolder(roots []string, outputfilename string) {

	var files = []Fthing{}
	for _, root := range roots {
		err := filepath.Walk(root, visit(&files, util))
		if err != nil {
			panic(err)
		}
	}

	for _, file := range files {
		fmt.Println(file)
	}

	// Create and add some files to the archive.
	f, err := os.Create(outputfilename)
	if err != nil {
		util.Session.Logger.LogErrorE("TarAndZipFolder", err)
		return
	}
	f.Close()

	file, err := os.OpenFile(outputfilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
	}

	var writer *gzip.Writer

	if writer, err = gzip.NewWriterLevel(file, gzip.BestCompression); err != nil {
		util.Session.Logger.LogErrorE("TarAndZipFolder", err)
	}

	defer writer.Close()

	tw1 := tar.NewWriter(writer)

	for _, file := range files {

		util.Session.Logger.LogInfo("TarAndZipFolder", fmt.Sprintf("Adding %s", file.Name))
		unixpath := switchtoUnix(file.Name)
		util.Session.Logger.LogInfo("TarAndZipFolder", fmt.Sprintf("switched to unix %s", unixpath))

		hdr := &tar.Header{
			Name: unixpath,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}

		if err := tw1.WriteHeader(hdr); err != nil {
			util.Session.Logger.LogErrorE("TarAndZipFolder", err)
		}

		if _, err := tw1.Write([]byte(file.Body)); err != nil {
			util.Session.Logger.LogErrorE("TarAndZipFolder", err)
		}
		tw1.Flush()
	}

	if err := tw1.Close(); err != nil {
		util.Session.Logger.LogErrorE("TarAndZipFolder", err)
	}

}

func switchtoUnix(filname string) string {
	return strings.Replace(filname, "\\", "/", -1)
}
