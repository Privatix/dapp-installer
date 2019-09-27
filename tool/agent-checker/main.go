package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/antchfx/htmlquery"
)

func main() {
	torsocks := flag.String("torsocks", "", "Tor socks port number")
	torhs := flag.String("torhs", "", "Tor hidden service onion")
	payPort := flag.String("payport", "", "Payment server's port number")
	ovpnPort := flag.String("ovpnport", "", "OpenVPN's port number")

	flag.Parse()

	if *torsocks == "" || *torhs == "" || *ovpnPort == "" || *payPort == "" {
		flag.Usage()
		return
	}

	if err := checkTorHiddenService(*torhs, *torsocks); err != nil {
		log.Panicf("could not reach tor hidden service: %v", err)
	}
	log.Println("TOR Hidden Service is OK.")

	if err := checkPort(*payPort, "TCP"); err != nil {
		log.Panicf("Payment Server's Port is NOT OK: %v", err)
	}
	log.Println("Payment Server's Port is OK.")

	if err := checkPort(*ovpnPort, "UDP"); err != nil {
		log.Panicf("OpenVPN's Port is NOT OK: %v", err)
	}
	log.Println("OpenVPN's Port is OK.")
}

func checkTorHiddenService(hs, socksN string) error {
	httpclient, err := torClient(socksN)
	if err != nil {
		return err
	}

	res, err := httpclient.Head("http://" + hs)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return err
	}

	return nil
}

func checkPort(port, proto string) error {
	doc, err := htmlquery.LoadURL("https://privatix.network/portcheck/?ports=" + port)
	if err != nil {
		return fmt.Errorf("could not connect to port checker: %v", err)
	}
	query := `//td[3][contains(., "Not reachable")]`
	if proto == "TCP" {
		query = `//td[2][contains(., "Not reachable")]`
	}
	list := htmlquery.Find(doc, query)
	if len(list) > 0 {
		return errors.New("Not reachable")
	}
	return nil
}

func torClient(sock string) (*http.Client, error) {
	torProxyURL, err := url.Parse(fmt.Sprint("socks5://127.0.0.1:", sock))
	if err != nil {
		return nil, err
	}

	// Set up a custom HTTP transport to use the proxy and create the client
	torTransport := &http.Transport{Proxy: http.ProxyURL(torProxyURL)}
	return &http.Client{
		Transport: torTransport,
		Timeout:   time.Second * 10,
	}, nil
}
