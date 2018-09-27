// +build linux darwin

//go:generate rm -f statik.go
//go:generate statik -f -src=. -dest=..

package statik
