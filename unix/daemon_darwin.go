// +build darwin

package unix

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"text/template"
	"time"
)

func (d *Daemon) name() string {
	// By convention names are written in reverse domain name notation.
	return "io.privatix." + d.ID
}

func (d *Daemon) path() string {
	usr, _ := user.Current()
	dir := filepath.Join(usr.HomeDir, "Library/LaunchAgents")

	return filepath.Join(dir, d.name()+".plist")
}

// Install installs a daemon.
func (d *Daemon) Install() error {
	dir, _ := filepath.Split(d.path())
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	file, err := os.Create(d.path())
	if err != nil {
		return err
	}
	defer file.Close()

	templ, err := template.New("daemonTemplate").Parse(daemonTemplate)
	if err != nil {
		return err
	}

	d.Name = d.name()
	return templ.Execute(file, &d)
}

// Start starts the daemon.
func (d *Daemon) Start() error {
	cmd := exec.Command("launchctl", "load", d.path())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to load: %v", err)
	}
	time.Sleep(time.Millisecond)
	cmd = exec.Command("launchctl", "start", d.name())
	return cmd.Run()
}

// Stop stops the daemon.
func (d *Daemon) Stop() error {
	cmd := exec.Command("launchctl", "unload", d.path())
	return cmd.Run()
}

// Remove removes the daemon.
func (d *Daemon) Remove() error {
	if !d.IsStopped() {
		return errors.New("can't remove a running daemon")
	}
	return os.Remove(d.path())
}

// IsStopped returns the daemon stopped status.
func (d *Daemon) IsStopped() bool {
	cmd := exec.Command("launchctl", "list", d.name())
	output, err := cmd.Output()

	if err != nil {
		return true
	}

	ok, err := regexp.MatchString(d.name(), string(output))
	if !ok || err != nil {
		return true
	}

	reg := regexp.MustCompile("PID\" = ([0-9]+);")
	data := reg.FindStringSubmatch(string(output))
	return len(data) < 2
}

var daemonTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Disabled</key>
	<false/>
	<key>Label</key>
	<string>{{.Name}}</string>
	<key>ProgramArguments</key>
	<array>
		<string>{{.Command}}</string>
		{{range .Args}}<string>{{.}}</string>
		{{end}}
	</array>
	<key>KeepAlive</key>
	{{if .AutoStart}}<true/>{{else}}<false/>{{end}}
	<key>StandardErrorPath</key>
	<string>{{.Command}}.err</string>
	<key>StandardOutPath</key>
	<string>{{.Command}}.log</string>
</dict>
</plist>
`
