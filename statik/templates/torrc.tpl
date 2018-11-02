SocksPort {{.SocksPort}}

HiddenServiceDir "{{.RootPath}}/{{.HiddenServiceDir}}"
HiddenServicePort {{.VirtPort}} 127.0.0.1:{{.TargetPort}}

GeoIPFile /settings/geoip
GeoIPv6File /settings/geoip6

#SafeLogging 0
Log notice file {{.RootPath}}/log/tor.log
#Log info file {{.RootPath}}/log/torinfo.log