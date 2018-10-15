package dapp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/privatix/dapp-installer/data"
	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dappctrl/util/log"
)

func setConfigurationValues(jsonMap map[string]interface{},
	settings map[string]interface{}) error {
	for key, value := range settings {
		path := strings.Split(key, ".")
		length := len(path) - 1
		m := jsonMap
		if length > 0 {
			for i := 0; i < length; i++ {
				item, ok := m[path[i]]
				if !ok || reflect.TypeOf(m) != reflect.TypeOf(item) {
					return fmt.Errorf("failed to set config params: %s", key)
				}
				m, _ = item.(map[string]interface{})
			}
		}
		m[path[length]] = value
	}
	return nil
}

func setDynamicPorts(configFile string, logger log.Logger) error {
	read, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	contents := string(read)
	addrs := util.MatchAddr(contents)
	reserves := make(map[string]bool)

	for _, addr := range addrs {
		port, err := util.FreePort(addr.Host, addr.Port)
		if err != nil {
			return err
		}

		// If the available port is already reserved,
		// the search for another unreserved available port in the loop.
		_, ok := reserves[port]
		for ok {
			p, _ := strconv.Atoi(port)
			port, _ := util.FreePort(addr.Host, strconv.Itoa(p+1))
			_, ok = reserves[port]
		}

		reserves[port] = true

		if port == addr.Port {
			continue
		}

		newAddress := strings.Replace(addr.Address,
			fmt.Sprintf(":%s", addr.Port),
			fmt.Sprintf(":%s", port), -1)
		contents = strings.Replace(contents, addr.Address,
			newAddress, -1)
		logger.Info(fmt.Sprintf("%s -> %s", addr.Address, newAddress))
	}
	return ioutil.WriteFile(configFile, []byte(contents), 0)
}

func roleAndDBConnFromConfig(configFile string) (string, *data.DB, error) {
	read, err := os.Open(configFile)
	if err != nil {
		return "", nil, err
	}
	defer read.Close()

	jsonMap := make(map[string]interface{})

	json.NewDecoder(read).Decode(&jsonMap)

	db, ok := jsonMap["DB"].(map[string]interface{})
	if !ok {
		return "", nil, fmt.Errorf("DB params not found")
	}
	conn, ok := db["Conn"].(map[string]interface{})
	if !ok {
		return "", nil, fmt.Errorf("Conn params not found")
	}

	res := data.NewConfig()

	if dbname, ok := conn["dbname"]; ok {
		res.DBName = dbname.(string)
	}

	if user, ok := conn["user"]; ok {
		res.User = user.(string)
	}

	if pwd, ok := conn["password"]; ok {
		res.Password = pwd.(string)
	}

	if port, ok := conn["port"]; ok {
		res.Port = port.(string)
	}

	role, ok := jsonMap["Role"]
	if !ok {
		return "", nil, fmt.Errorf("Role not found")
	}

	return role.(string), res, nil
}
