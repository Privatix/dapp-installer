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

// SetUID sets uid daemon is running on.
func (d *Daemon) SetUID(uid string) {
	d.UID = uid
}

func (d *Daemon) path() string {
	var homedir string
	if d.UID == "" {
		usr, _ := user.Current()
		homedir = usr.HomeDir
	} else {
		usr, _ := user.LookupId(d.UID)
		homedir = usr.HomeDir
	}
	dir := filepath.Join(homedir, "Library/LaunchAgents")

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
	cmd := d.buildLaunchctlCommand("load", d.path())
	if err := cmd.Run(); err != nil {
		out, errout := cmd.Output()
		return fmt.Errorf("failed to load %s: %v, Out: `%s`, Error: %v,", d.path(), err, string(out), errout)
	}
	time.Sleep(100 * time.Millisecond)
	cmd = d.buildLaunchctlCommand("start", d.name())
	if err := cmd.Run(); err != nil {
		out, errout := cmd.Output()
		return fmt.Errorf("failed to start %s: %v, Out: `%s`, Error: %v", d.path(), err, string(out), errout)
	}
	return nil
}

// Stop stops the daemon.
func (d *Daemon) Stop() error {
	cmd := d.buildLaunchctlCommand("unload", d.path())
	return cmd.Run()
}

func (d *Daemon) buildLaunchctlCommand(args ...string) *exec.Cmd {
	if d.UID != "" {
		args = append([]string{"asuser", d.UID, "launchctl"}, args...)
	}
	return exec.Command("launchctl", args...)
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
	cmd := d.buildLaunchctlCommand("list", d.name())
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
	<key>RunAtLoad</key>
	{{ if .AutoStart }}
	<true/>
	{{ else }}
	<false/>
	{{end}}
	<key>KeepAlive</key>
	<dict>
		<key>AfterInitialDemand</key>
		{{ if .AutoStart }}	
		<false/>
		{{ else }}
		<true/>
		{{ end }}
		<key>SuccessfulExit</key>
		<false/>
		<key>Crashed</key>
		<true/>
	</dict>
	<key>StandardErrorPath</key>
	<string>{{.Command}}.err</string>
	<key>StandardOutPath</key>
	<string>{{.Command}}.log</string>
</dict>
</plist>
`
