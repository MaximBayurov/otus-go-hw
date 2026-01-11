package configuration

import (
	"path/filepath"
	"strings"
)

type configPathInfo struct {
	Name string
	Path string
}

func getConfigPathInfoFor(path string) configPathInfo {
	fileName := filepath.Base(path)

	return configPathInfo{
		Name: strings.TrimSuffix(fileName, filepath.Ext(path)),
		Path: strings.TrimSuffix(path, fileName),
	}
}
