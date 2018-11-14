SocksPort {{.SocksPort}}

HiddenServiceDir "{{.RootPath}}/{{.HiddenServiceDir}}"
HiddenServicePort {{.VirtPort}} 127.0.0.1:{{.TargetPort}}

DataDirectory "{{.RootPath}}/tor/data"

Log notice file {{.RootPath}}/log/tor.log
