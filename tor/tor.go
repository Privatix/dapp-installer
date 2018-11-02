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
	Addr             string
	SocksPort        int
	VirtPort         int
	TargetPort       int
}

// NewTor creates a default Tor configuration.
func NewTor() *Tor {
	return &Tor{
		Addr:             "127.0.0.1",
		HiddenServiceDir: "tor/hidden_service",
		SocksPort:        9999,
		VirtPort:         80,
		TargetPort:       80,
	}
}

func (t *Tor) configurate() error {
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
	file, err := os.Create(filepath.Join(t.RootPath, "tor/settings/torrc"))
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

	return templ.Execute(file, &t)
}

// Install installs tor process.
func (t *Tor) Install() error {
	if err := t.configurate(); err != nil {
		return err
	}

	return installService(t.ServiceName(), t.RootPath)
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
