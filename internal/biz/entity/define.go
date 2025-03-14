package entity

import (
	"path/filepath"
	"time"
)

func GetFileStoragePath(constPath string, runtime bool) string {
	if runtime {
		return filepath.Join(constPath, "runtime", time.Now().Format("20060102"))
	}
	return filepath.Join(constPath, "static", time.Now().Format("20060102"))
}
