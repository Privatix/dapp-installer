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

// Installation is TOR installation details.
type Installation struct {
	Hostname         string
	HiddenServiceDir string
	RootPath         string
	Config           string
	Addr             string
	InstallUID       string
	SocksPort        int
	VirtPort         int
	TargetPort       int
}

// NewInstallation return new instance with default values.
func NewInstallation() *Installation {
	return &Installation{
		Addr:             "127.0.0.1",
		HiddenServiceDir: "tor/hidden_service",
		SocksPort:        9999,
		VirtPort:         80,
		TargetPort:       3452,
		Config:           "torrc",
	}
}

// The private key and hostname are generated according to the instruction
// https://trac.torproject.org/projects/tor/wiki/doc/HiddenServiceNames
func (inst *Installation) createHiddenService() error {
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
	inst.Hostname = strings.ToLower(onion) + ".onion"

	path := filepath.Join(inst.RootPath, inst.HiddenServiceDir)
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

func (inst *Installation) writeTorrc() error {
	path := filepath.Join(inst.RootPath, "tor", "settings")
	if runtime.GOOS == "linux" {
		path = filepath.Join(inst.RootPath, "etc", "tor")
	}
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err := os.MkdirAll(path, 0644); err != nil {
			return err
		}
	}

	file, err := os.Create(filepath.Join(path, inst.Config))
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

	p, _ := util.FreePort(inst.Addr, strconv.Itoa(inst.SocksPort))
	inst.SocksPort, _ = strconv.Atoi(p)

	p, _ = util.FreePort(inst.Addr, strconv.Itoa(inst.TargetPort))
	inst.TargetPort, _ = strconv.Atoi(p)

	return templ.Execute(file, inst)
}

// InstallAgent prepares hidden service, writes torrc and install TOR as a service.
func (inst *Installation) InstallAgent(autostart bool) error {
	if err := inst.createHiddenService(); err != nil {
		return err
	}
	if err := inst.writeTorrc(); err != nil {
		return err
	}
	descr := fmt.Sprintf("Privatix Agent Tor transport %s", util.Hash(inst.RootPath))
	return installService(inst.ServiceName(), inst.RootPath, descr, autostart)
}

// InstallClient prepares hidden service, writes torrc and install TOR as a service.
func (inst *Installation) InstallClient(autostart bool) error {
	if err := inst.writeTorrc(); err != nil {
		return err
	}
	descr := fmt.Sprintf("Privatix Client Tor transport %s", util.Hash(inst.RootPath))
	return installService(inst.ServiceName(), inst.RootPath, descr, autostart)
}

// Start starts tor process.
func (inst *Installation) Start(ctx context.Context) (err error) {
	for {
		select {
		case <-ctx.Done():
			if err != nil {
				return err
			}
			return ctx.Err()
		default:
			if !util.IsServiceStopped(inst.ServiceName(), inst.InstallUID) {
				return nil
			}
			err = startService(inst.ServiceName(), inst.InstallUID)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Stop stops tor process.
func (inst *Installation) Stop(ctx context.Context, installUID string) (err error) {
	for {
		select {
		case <-ctx.Done():
			if err != nil {
				return err
			}
			return ctx.Err()
		default:
			if util.IsServiceStopped(inst.ServiceName(), installUID) {
				return nil
			}
			err = stopService(inst.ServiceName(), installUID)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Remove removes tor process.
func (inst *Installation) Remove() error {
	return removeService(inst.ServiceName())
}
