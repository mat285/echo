package core

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	exception "github.com/blend/go-sdk/exception"
)

// IsTarExcluded returns if a file is excluded from a tarball.
func IsTarExcluded(name string, excludePatterns []string) (bool, error) {
	for _, exclude := range excludePatterns {
		if !strings.HasPrefix(exclude, string(filepath.Separator)) {
			if matched, err := filepath.Match(exclude, filepath.Base(name)); err != nil {
				return false, err
			} else if matched {
				return true, nil
			}
		}

		if matched, err := filepath.Match(exclude, name); err != nil {
			return false, err
		} else if matched {
			return true, nil
		}
	}
	return false, nil
}

// TarCompress compresses a path into a reader.
func TarCompress(path string, excludePatterns []string, destination io.Writer) error {
	// ensure the src actually exists before trying to tar it
	if _, err := os.Stat(path); err != nil {
		return exception.New(fmt.Sprintf("Unable to tar files - %v", err.Error()))
	}

	gzw := gzip.NewWriter(destination)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	err := filepath.Walk(path, func(file string, fi os.FileInfo, err error) error {
		// return on any error
		if err != nil {
			return exception.New(err)
		}

		filename := strings.TrimPrefix(strings.Replace(file, path, "", -1), string(filepath.Separator))
		isExcluded, err := IsTarExcluded(filename, excludePatterns)
		if err != nil {
			return err
		}
		if isExcluded {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		var link string
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			if link, err = os.Readlink(file); err != nil {
				return exception.New(err)
			}
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, link)
		if err != nil {
			return exception.New(err)
		}

		// update the name to correctly reflect the desired destination when untaring
		header.Name = filename

		// write the header
		if err := tw.WriteHeader(header); err != nil {
			return exception.New(err)
		}

		if !fi.Mode().IsRegular() { //nothing more to do for non-regular
			return nil
		}

		// open files for taring
		f, err := os.Open(file)
		defer f.Close()
		if err != nil {
			return exception.New(err)
		}

		// copy file data into tar writer
		if _, err := io.Copy(tw, f); err != nil {
			return exception.New(err)
		}
		return nil
	})
	if err != nil {
		return exception.New(err)
	}

	return nil
}
