package uploader

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
)

const (
	JS     string = "js"
	CSS           = "css"
	IMG           = "image"
	UNKNOW        = "unknow"
)

type File struct {
	Name     string
	FileType string
	Content  io.Reader
}

func NewFile(filePath string) (*File, error) {
	file := &File{}
	fileData, err := os.ReadFile(filePath)

	if err != nil {
		return file, errors.New("Falha ao criar o arquivo.")
	}

	_, fileName := filepath.Split(filePath)

	file.Name = fileName
	file.FileType = getFileType(fileName)
	file.Content = bytes.NewReader(fileData)

	return file, nil
}

func getFileType(filePath string) string {
	ext := filepath.Ext(filePath)

	switch ext {
	case ".json":
		return JS
	case ".css":
		return CSS
	case ".png":
		return IMG
	case ".jpg":
		return IMG
	case ".gif":
		return IMG
	default:
		return UNKNOW
	}
}
