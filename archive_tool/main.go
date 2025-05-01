package main

import (
	"archive/zip"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	var (
		output      string
		modulePath  string
		version     string
		goMod       string
		srcFiles    arrayFlags
		stripPrefix string
	)
	flag.StringVar(&output, "output", "", "Path to output .zip file")
	flag.StringVar(&modulePath, "module-path", "", "Module path (e.g., github.com/my_project)")
	flag.StringVar(&version, "version", "", "Module version (e.g., v0.1.0)")
	flag.StringVar(&goMod, "go-mod", "", "Path to go.mod file")
	flag.Var(&srcFiles, "src", "Path to a .go source file (can be repeated)")
	flag.StringVar(&stripPrefix, "strip-prefix", "", "Prefix to strip from source file paths")
	flag.Parse()

	if output == "" || modulePath == "" || version == "" || goMod == "" || len(srcFiles) == 0 {
		log.Fatal("Missing required flags: --output, --module-path, --version, --go-mod, --src")
	}

	zipFile, err := os.Create(output)
	if err != nil {
		log.Fatalf("Failed to create %s: %v", output, err)
	}
	defer zipFile.Close()

	zw := zip.NewWriter(zipFile)
	defer zw.Close()

	moduleDir := modulePath + "@" + version

	if err := addFileToZip(zw, goMod, filepath.Join(moduleDir, "go.mod")); err != nil {
		log.Fatalf("Failed to add go.mod to zip: %v", err)
	}

	for _, src := range srcFiles {
		relPath := src
		if stripPrefix != "" && strings.HasPrefix(src, stripPrefix) {
			relPath = strings.TrimPrefix(src, stripPrefix)
			relPath = strings.TrimPrefix(relPath, string(filepath.Separator)) // avoid leading slash
		}
		zipPath := filepath.Join(moduleDir, relPath)
		if err := addFileToZip(zw, src, zipPath); err != nil {
			log.Fatalf("Failed to add %s to zip: %v", src, err)
		}
	}

	if err := zw.Flush(); err != nil {
		log.Fatalf("Failed to flush zip: %v", err)
	}
}

func addFileToZip(zw *zip.Writer, srcPath, zipPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	zipEntry, err := zw.Create(zipPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(zipEntry, srcFile)
	return err
}
