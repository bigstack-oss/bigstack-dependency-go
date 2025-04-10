package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func DecompressFromTo(src, dest string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}

	defer reader.Close()
	for _, file := range reader.File {
		path := filepath.Join(dest, file.Name)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("zip: illegal file path: %s", path)
		}

		if file.FileInfo().IsDir() {
			err := os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}

		err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
		if err != nil {
			return err
		}

		readCloser, err := file.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			readCloser.Close()
			return err
		}

		_, err = io.Copy(outFile, readCloser)
		outFile.Close()
		readCloser.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
