package product

type installConfig struct {
	CoreDapp bool
	Install  []command
	Update   []command
	Remove   []command
	Start    []command
	Stop     []command
}

type command struct {
	Admin   bool
	Command string
}
