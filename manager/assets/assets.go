// Package assets
//
//	Author:		龙泽坤
//	Company:	奇富科技
//	Email:		longzekun-jk@360shuke.com
//				lzk1342325850@163.com
//	Date:		2025-05-21
//	File:		assets.go
//
//	Description(获取当前的工作路径):
//
//
//	Change:
package assets

import (
	"os"
	"path/filepath"
)

const (
	ManagerRootDir = "ManagerRootDir"
)

// GetManagerRootAppDir - 获取当前工作的根路径
func GetManagerRootAppDir() string {
	value := os.Getenv(ManagerRootDir)
	var dir string

	if len(value) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		dir = filepath.Join(wd, ".manager")
	} else {
		dir = value
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			panic(err)
		}
	}
	return dir
}
