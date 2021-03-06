SocksPort {{.SocksPort}}

HiddenServiceDir "{{if not .IsLinux}}{{.RootPath}}{{end}}/{{.HiddenServiceDir}}"
HiddenServicePort {{.VirtPort}} 127.0.0.1:{{.TargetPort}}

{{if not .IsLinux}}
DataDirectory "{{.RootPath}}/tor/data"

Log notice file {{.RootPath}}/log/tor.log
{{end}}
