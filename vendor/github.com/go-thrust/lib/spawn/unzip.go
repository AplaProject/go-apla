package spawn

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	fmt.Println("Unzipping", src, "to", dest)

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		filePath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, 0775); err != nil {
				return err
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(filePath), 0775); err != nil {
				return err
			}
			file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(file, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
