// +build windows

//go:generate go build -o ./wrapper/winsvc.exe ../tool/winsvc/
//go:generate cmd /C IF EXIST "statik.go" (del /F /Q statik.go)
//go:generate statik -f -src=. -dest=..

package statik
