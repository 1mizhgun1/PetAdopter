package utils

import (
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"strings"
)

func GetFormat(choice map[string]string, content []byte) string {
	fileFormat := http.DetectContentType(content)

	for mimeType, format := range choice {
		if strings.HasPrefix(fileFormat, mimeType) {
			return format
		}
	}

	return ""
}

func WriteFileOnDisk(path string, extension string, resource io.ReadSeeker) error {
	file, err := os.Create(path + extension)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = resource.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, resource)
	if err != nil {
		return err
	}

	return nil
}
