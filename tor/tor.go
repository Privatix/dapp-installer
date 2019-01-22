package tor

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base32"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/privatix/dapp-installer/statik"
	"github.com/privatix/dapp-installer/util"
)

// Tor has a tor configuration.
type Tor struct {
	Hostname         string
	HiddenServiceDir string
	RootPath         string
	Config           string
	Addr             string
	SocksPort        int
	VirtPort         int
	TargetPort       int
	IsLinux          bool
}

// NewTor creates a default Tor configuration.
func NewTor() *Tor {
	return &Tor{
		Addr:             "127.0.0.1",
		HiddenServiceDir: "tor/hidden_service",
		SocksPort:        9999,
		VirtPort:         80,
		TargetPort:       5555,
		Config:           "torrc",
	}
}

func (t *Tor) configurate() error {
	if t.IsLinux {
		return t.createSettings()
	}

	if err := t.generateKey(); err != nil {
		return err
	}

	return t.createSettings()
}

// The private key and hostname are generated according to the instruction
// https://trac.torproject.org/projects/tor/wiki/doc/HiddenServiceNames
func (t *Tor) generateKey() error {
	priv, err := rsa.GenerateKey(rand.Reader, 1024)

	if err != nil {
		return err
	}

	bytes, err := asn1.Marshal(priv.PublicKey)
	if err != nil {
		return err
	}

	h := sha1.New()
	h.Write(bytes)

	src := h.Sum(nil)
	length := hex.EncodedLen(len(src))

	onion := base32.StdEncoding.EncodeToString(src[:length/4])
	t.Hostname = strings.ToLower(onion) + ".onion"

	path := filepath.Join(t.RootPath, t.HiddenServiceDir)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0700)
	}

	keyOut, err := os.OpenFile(filepath.Join(path, "private_key"),
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	return pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv)})
}

func (t *Tor) createSettings() error {
	path := filepath.Join(t.RootPath, "tor", "settings")
	if t.IsLinux {
		path = filepath.Join(t.RootPath, "etc", "tor")
	}
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err := os.MkdirAll(path, 0644); err != nil {
			return err
		}
	}

	file, err := os.Create(filepath.Join(path, t.Config))
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := statik.ReadFile("/templates/torrc.tpl")
	if err != nil {
		return err
	}

	templ, err := template.New("torTemplate").Parse(string(data))
	if err != nil {
		return err
	}

	p, _ := util.FreePort(t.Addr, strconv.Itoa(t.SocksPort))
	t.SocksPort, _ = strconv.Atoi(p)

	p, _ = util.FreePort(t.Addr, strconv.Itoa(t.TargetPort))
	t.TargetPort, _ = strconv.Atoi(p)

	return templ.Execute(file, &t)
}

// Install installs tor process.
func (t *Tor) Install(role string) error {
	if err := t.configurate(); err != nil {
		return err
	}

	descr := fmt.Sprintf("Privatix %s Tor transport %s", role,
		util.Hash(t.RootPath))
	return installService(t.ServiceName(), t.RootPath, descr)
}

// Start starts tor process.
func (t *Tor) Start() error {
	return startService(t.ServiceName())
}

// Stop stops tor process.
func (t *Tor) Stop() error {
	if util.IsServiceStopped(t.ServiceName()) {
		return nil
	}
	return stopService(t.ServiceName())
}

// Remove removes tor process.
func (t *Tor) Remove() error {
	return removeService(t.ServiceName())
}
