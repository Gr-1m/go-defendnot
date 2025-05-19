package loader

import (
	"fmt"
	"github.com/Gr-1m/sys/windows"
	"github.com/Gr-1m/sys/windows/extra"
	"unsafe"
)

type AutoRunType uint8

const (
	AsSystemOnBoot       AutoRunType = 0
	AsCurrentUserOnLogin AutoRunType = 1

	ProjName = "go-defendnot"
	RepoUrl  = "https://github.com/Gr-1m/go-defendnot"

	VictimProcess = "Taskmgr.exe"
	DllName       = "defendnot.dll"

	Version = "0.8.1"
)

type ArgsConfig struct {
	Name          string
	Disable       bool
	Verbose       bool
	AutoRun       bool
	AutoRunType   AutoRunType
	EnableAutoRun bool
}

// func AutoRunAdd(t AutoRunType) bool {
func AutoRunAdd(appName, exePath string) error {
	// 注册表路径：HKCU\Software\Microsoft\Windows\CurrentVersion\Run
	keyPath, _ := windows.UTF16PtrFromString("Software\\Microsoft\\Windows\\CurrentVersion\\Run")
	var hKey windows.Handle

	// 创建/打开注册表键
	err := extra.RegCreateKeyEx(windows.HKEY_CURRENT_USER, keyPath, 0, nil, 0, windows.KEY_SET_VALUE, nil, &hKey, nil)
	if err != nil {
		return fmt.Errorf("打开注册表失败: %v", err)
	}
	defer windows.RegCloseKey(hKey)

	// 准备要写入的值
	valueName, _ := windows.UTF16PtrFromString(appName)
	exePathUTF16, _ := windows.UTF16PtrFromString(exePath)
	exePathBytes := unsafe.Slice((*byte)(unsafe.Pointer(exePathUTF16)), (len(exePath)+1)*2)

	// 设置注册表值
	err = extra.RegSetValueEx(hKey, valueName, 0, windows.REG_SZ, &exePathBytes[0], uint32(len(exePathBytes)))
	if err != nil {
		return fmt.Errorf("写入注册表失败: %v", err)
	}

	return nil
}

func AutoRunRemove() error {
	return nil
}
