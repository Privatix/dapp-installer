package metadata

// Installation is metadata of installation.
type Installation struct {
	Schema     string
	Role       string
	WorkDir    string
	UserID     string
	Version    string
	Dapp       dapp
	DB         service
	Tor        service
	Supervisor service
}

type service struct {
	Service string
}

type dapp struct {
	service
	Controller string
	Gui        string
}
