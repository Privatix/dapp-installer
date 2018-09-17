// +build windows

package windows

import (
	"syscall"
	"unsafe"
)

// CreateShortcut creates Windows shortcut.
func CreateShortcut(target, descr, args, link string) error {
	ole32, err := syscall.LoadLibrary("ole32.dll")
	if err != nil {
		return err
	}
	defer syscall.FreeLibrary(ole32)

	coInitialize, _ := syscall.GetProcAddress(ole32, "CoInitialize")
	coCreateInstance, _ := syscall.GetProcAddress(ole32, "CoCreateInstance")

	if _, _, err := syscall.Syscall(coInitialize, 0, 0, 0, 0); err != 0 {
		return err
	}

	var psl *IShellLink

	if _, _, err := syscall.Syscall6(coCreateInstance, 5,
		uintptr(unsafe.Pointer(&CLSIDShellLink)), uintptr(0), uintptr(0x1),
		uintptr(unsafe.Pointer(&IIDIShellLink)),
		uintptr(unsafe.Pointer(&psl)), 0); err != 0 {
		return err
	}
	defer psl.Release()

	t, _ := syscall.UTF16PtrFromString(target)
	psl.SetPath(t)
	d, _ := syscall.UTF16PtrFromString(descr)
	psl.SetDescription(d)
	a, _ := syscall.UTF16PtrFromString(args)
	psl.SetArguments(a)
	psl.SetWorkingDirectory(t)
	psl.SetShowCmd(1)

	var ppf *IPersistFile
	psl.QueryInterface(&IIDIPersistFile, unsafe.Pointer(&ppf))
	defer ppf.Release()

	l, _ := syscall.UTF16PtrFromString(link)
	ppf.Save(l, 1)
	return nil
}
