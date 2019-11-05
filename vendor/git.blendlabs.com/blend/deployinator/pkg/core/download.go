package core

import (
	"io"
	"os"

	exception "github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/request"
)

// DownloadFile will download a url to a local file
func DownloadFile(file io.Writer, url string) error {
	// Get the data
	req := request.Get(url)
	resp, err := req.Response()
	if err != nil {
		return exception.New(err)
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(file, resp.Body)
	return exception.New(err)
}

// InstallExecutable installs an executable to the specified location
func InstallExecutable(location string, url string) error {
	err := os.Remove(location)
	if err != nil && !os.IsNotExist(err) {
		return exception.New(err)
	}
	file, err := os.Create(location)
	if err != nil {
		return exception.New(err)
	}
	err = DownloadFile(file, url)
	if err != nil {
		file.Close()
		os.Remove(file.Name())
		return exception.New(err)
	}
	defer file.Close()
	return exception.New(file.Chmod(os.FileMode(0777)))
}
