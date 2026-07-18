// Command packagezip creates deterministic release archives.
package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
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
	format := flag.String("format", "zip", "archive format: zip or targz")
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
	var err error
	switch *format {
	case "zip":
		err = createArchive(*source, *output, *binary)
	case "targz":
		err = createTarGzip(*source, *output, *binary)
	default:
		fmt.Fprintln(os.Stderr, "format must be zip or targz")
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "create archive: %v\n", err)
		os.Exit(1)
	}
}

func createTarGzip(source, output, binary string) (returnError error) {
	archive, err := os.Create(output)
	if err != nil {
		return err
	}
	defer func() {
		if err := archive.Close(); returnError == nil && err != nil {
			returnError = err
		}
	}()
	compressed := gzip.NewWriter(archive)
	compressed.ModTime = time.Unix(0, 0).UTC()
	compressed.OS = 255
	defer func() {
		if err := compressed.Close(); returnError == nil && err != nil {
			returnError = err
		}
	}()
	writer := tar.NewWriter(compressed)
	defer func() {
		if err := writer.Close(); returnError == nil && err != nil {
			returnError = err
		}
	}()
	for _, name := range []string{"LICENSE", "THIRD_PARTY_NOTICES.md", binary} {
		path := filepath.Join(source, name)
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		mode := int64(0o644)
		if info.Mode()&0o111 != 0 {
			mode = 0o755
		}
		header := &tar.Header{Name: name, Mode: mode, Size: info.Size(), ModTime: archiveTime, AccessTime: archiveTime, ChangeTime: archiveTime, Uid: 0, Gid: 0, Typeflag: tar.TypeReg}
		if err := writer.WriteHeader(header); err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(writer, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
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
