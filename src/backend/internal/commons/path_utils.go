package commons

import (
	"fmt"
	"os"
	"path/filepath"
)

// 取得執行檔的路徑
func GetBinPath() string {
	// 取得執行檔的絕對路徑
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	binPath := filepath.Dir(exePath)
	return binPath
}

// 取得應用程式的路徑 (bin 的上一層)
func GetHome() string {
	binPath := GetBinPath()
	home := filepath.Dir(binPath)
	return home
}

// 以 HOME 為基礎的相對路徑取絕對路徑
func GetPathFromHome(configItem string) string {
	path := fmt.Sprintf("%s/%s", GetHome(), configItem)
	return filepath.FromSlash(path)
}
