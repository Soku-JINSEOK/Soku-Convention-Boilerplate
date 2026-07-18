// Command packagezip creates the deterministic Windows release archive.
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

var archiveTime = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)

func main() {
	source := flag.String("source", "", "directory containing release files")
	output := flag.String("output", "", "output zip archive")
	binary := flag.String("binary", "", "binary filename")
	list := flag.String("list", "", "list an existing archive")
	flag.Parse()
	if *list != "" {
		if err := listArchive(*list); err != nil {
			fmt.Fprintf(os.Stderr, "list zip archive: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if *source == "" || *output == "" || *binary == "" {
		fmt.Fprintln(os.Stderr, "source, output, and binary are required")
		os.Exit(2)
	}
	if err := createArchive(*source, *output, *binary); err != nil {
		fmt.Fprintf(os.Stderr, "create zip archive: %v\n", err)
		os.Exit(1)
	}
}

func listArchive(path string) (returnError error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := reader.Close(); returnError == nil && err != nil {
			returnError = err
		}
	}()
	for _, file := range reader.File {
		fmt.Println(file.Name)
	}
	return nil
}

func createArchive(source, output, binary string) (returnError error) {
	archive, err := os.Create(output)
	if err != nil {
		return err
	}
	defer func() {
		if err := archive.Close(); returnError == nil && err != nil {
			returnError = err
		}
	}()

	writer := zip.NewWriter(archive)
	defer func() {
		if err := writer.Close(); returnError == nil && err != nil {
			returnError = err
		}
	}()

	for _, name := range []string{"LICENSE", "THIRD_PARTY_NOTICES.md", binary} {
		if err := addFile(writer, source, name); err != nil {
			return err
		}
	}
	return nil
}

func addFile(writer *zip.Writer, source, name string) error {
	path := filepath.Join(source, name)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = name
	header.Method = zip.Deflate
	header.Modified = archiveTime

	destination, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}
	sourceFile, err := os.Open(path)
	if err != nil {
		return err
	}
	_, copyError := io.Copy(destination, sourceFile)
	closeError := sourceFile.Close()
	if copyError != nil {
		return copyError
	}
	return closeError
}
