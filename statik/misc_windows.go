// +build windows

//go:generate go build -o ./wrapper/winsvc.exe ../tool/winsvc/
//go:generate statik -f -src=. -dest=..

package statik
