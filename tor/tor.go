package tor

import (
	"context"
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
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"

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
		TargetPort:       3452,
		Config:           "torrc",
		IsLinux:          runtime.GOOS == "linux",
	}
}

func (t *Tor) configure(installUID string) error {
	if err := t.generateKey(installUID); err != nil {
		return err
	}

	return t.createSettings()
}

// The private key and hostname are generated according to the instruction
// https://trac.torproject.org/projects/tor/wiki/doc/HiddenServiceNames
func (t *Tor) generateKey(installUID string) error {
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
		if err := os.MkdirAll(path, 0700); err != nil {
			return err
		}
		if installUID != "" {
			if err := util.ChownToUID(path, installUID); err != nil {
				return err
			}
		}
	}

	pkeyFilePath := filepath.Join(path, "private_key")
	keyOut, err := os.OpenFile(pkeyFilePath,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	err = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	if err != nil {
		return err
	}

	if installUID == "" {
		return nil
	}

	return util.ChownToUID(pkeyFilePath, installUID)
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
func (t *Tor) Install(role string, autostart bool, installUID string) error {
	if err := t.configure(installUID); err != nil {
		return err
	}

	descr := fmt.Sprintf("Privatix %s Tor transport %s", role,
		util.Hash(t.RootPath))
	return installService(t.ServiceName(), t.RootPath, descr, installUID, autostart)
}

// Start starts tor process.
func (t *Tor) Start(ctx context.Context, installUID string) (err error) {
	for {
		select {
		case <-ctx.Done():
			if err != nil {
				return err
			}
			return ctx.Err()
		default:
			if !util.IsServiceStopped(t.ServiceName(), installUID) {
				return nil
			}
			err = startService(t.ServiceName(), installUID)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Stop stops tor process.
func (t *Tor) Stop(ctx context.Context, installUID string) (err error) {
	for {
		select {
		case <-ctx.Done():
			if err != nil {
				return err
			}
			return ctx.Err()
		default:
			if util.IsServiceStopped(t.ServiceName(), installUID) {
				return nil
			}
			err = stopService(t.ServiceName(), installUID)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Remove removes tor process.
func (t *Tor) Remove(installUID string) error {
	return removeService(t.ServiceName(), installUID)
}
