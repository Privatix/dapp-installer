// +build windows

package windows

import (
	"syscall"
	"unsafe"
)

var (
	CLSIDShellLink = syscall.GUID{0x00021401, 0x0000, 0x0000,
		[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
	IIDIShellLink = syscall.GUID{0x000214F9, 0x0000, 0x0000,
		[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
	IIDIPersistFile = syscall.GUID{0x0000010B, 0x0000, 0x0000,
		[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
)

// IPersistFileVtbl is used for saving shortcuts to files
type IPersistFileVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	GetClassID     uintptr
	IsDirty        uintptr
	Load           uintptr
	Save           uintptr
	SaveCompleted  uintptr
	GetCurFile     uintptr
}

// IPersistFile is used for saving shortcuts to files
type IPersistFile struct {
	LpVtbl *IPersistFileVtbl
}

// QueryInterface retrieves pointers to the supported interfaces on an object.
func (obj *IPersistFile) QueryInterface(riid *syscall.GUID,
	ppvObject unsafe.Pointer) (uint32, error) {
	ret, _, err := syscall.Syscall(obj.LpVtbl.QueryInterface, 3,
		uintptr(unsafe.Pointer(obj)),
		uintptr(unsafe.Pointer(riid)),
		uintptr(ppvObject))

	return uint32(ret), err
}

// AddRef increments the reference count for an interface on an object.
func (obj *IPersistFile) AddRef() (uint32, error) {
	ret, _, err := syscall.Syscall(obj.LpVtbl.AddRef, 1,
		uintptr(unsafe.Pointer(obj)),
		0,
		0)

	return uint32(ret), err
}

// Release decrements the reference count for an interface on an object.
func (obj *IPersistFile) Release() (uint32, error) {
	ret, _, err := syscall.Syscall(obj.LpVtbl.Release, 1,
		uintptr(unsafe.Pointer(obj)),
		0,
		0)

	return uint32(ret), err
}

// Save creates new lnk file.
func (obj *IPersistFile) Save(pszName *uint16, fRemember int) (uint32, error) {
	ret, _, err := syscall.Syscall(obj.LpVtbl.Save, 3,
		uintptr(unsafe.Pointer(obj)),
		uintptr(unsafe.Pointer(pszName)),
		uintptr(fRemember))

	return uint32(ret), err
}

// IShellLinkVtbl provides all of the methods for shortcut files.
type IShellLinkVtbl struct {
	QueryInterface      uintptr
	AddRef              uintptr
	Release             uintptr
	GetArguments        uintptr
	GetDescription      uintptr
	GetHotkey           uintptr
	GetIconLocation     uintptr
	GetIDList           uintptr
	GetPath             uintptr
	GetShowCmd          uintptr
	GetWorkingDirectory uintptr
	Resolve             uintptr
	SetArguments        uintptr
	SetDescription      uintptr
	SetHotkey           uintptr
	SetIconLocation     uintptr
	SetIDList           uintptr
	SetPath             uintptr
	SetRelativePath     uintptr
	SetShowCmd          uintptr
	SetWorkingDirectory uintptr
}

// IShellLink provides all of the methods for shortcut files.
type IShellLink struct {
	LpVtbl *IShellLinkVtbl
}

// SetPath sets the path and file name.
func (obj *IShellLink) SetPath(pszFile *uint16) (uint32, error) {
	ret, _, err := syscall.Syscall(obj.LpVtbl.SetPath, 2,
		uintptr(unsafe.Pointer(obj)),
		uintptr(unsafe.Pointer(pszFile)),
		0)

	return uint32(ret), err
}

// SetShowCmd sets the show command.
func (obj *IShellLink) SetShowCmd(iShowCmd int) (uint32, error) {
	ret, _, err := syscall.Syscall(obj.LpVtbl.SetShowCmd, 2,
		uintptr(unsafe.Pointer(obj)),
		uintptr(iShowCmd),
		0)

	return uint32(ret), err
}

// SetDescription sets the descriptions.
func (obj *IShellLink) SetDescription(pszName *uint16) (uint32, error) {
	ret, _, err := syscall.Syscall(obj.LpVtbl.SetDescription, 2,
		uintptr(unsafe.Pointer(obj)),
		uintptr(unsafe.Pointer(pszName)),
		0)

	return uint32(ret), err
}

// SetArguments sets the arguments.
func (obj *IShellLink) SetArguments(pszArgs *uint16) (uint32, error) {
	ret, _, err := syscall.Syscall(obj.LpVtbl.SetArguments, 2,
		uintptr(unsafe.Pointer(obj)),
		uintptr(unsafe.Pointer(pszArgs)),
		0)

	return uint32(ret), err
}

// SetWorkingDirectory sets the working directory.
func (obj *IShellLink) SetWorkingDirectory(pszDir *uint16) (uint32, error) {
	ret, _, err := syscall.Syscall(obj.LpVtbl.SetWorkingDirectory, 2,
		uintptr(unsafe.Pointer(obj)),
		uintptr(unsafe.Pointer(pszDir)),
		0)

	return uint32(ret), err
}

// QueryInterface retrieves pointers to the supported interfaces on an object.
func (obj *IShellLink) QueryInterface(riid *syscall.GUID,
	ppvObject unsafe.Pointer) (uint32, error) {
	ret, _, err := syscall.Syscall(obj.LpVtbl.QueryInterface, 3,
		uintptr(unsafe.Pointer(obj)),
		uintptr(unsafe.Pointer(riid)),
		uintptr(ppvObject))

	return uint32(ret), err
}

// Release decrements the reference count for an interface on an object.
func (obj *IShellLink) Release() (uint32, error) {
	ret, _, err := syscall.Syscall(obj.LpVtbl.Release, 1,
		uintptr(unsafe.Pointer(obj)),
		0,
		0)

	return uint32(ret), err
}

// AddRef increments the reference count for an interface on an object.
func (obj *IShellLink) AddRef() (uint32, error) {
	ret, _, err := syscall.Syscall(obj.LpVtbl.AddRef, 1,
		uintptr(unsafe.Pointer(obj)),
		0,
		0)

	return uint32(ret), err
}
