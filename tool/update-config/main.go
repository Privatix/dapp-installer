package main

import (
	"encoding/json"
	"flag"

	"github.com/privatix/dapp-installer/util"
)

func main() {
	src := flag.String("source", "source.json", "source config path to update from")
	dst := flag.String("dest", "dest.json", "destination config to update")
	copyItemsJSON := flag.String("copyItems", "[['foo', 'subfoo'], ['bar', 'subbar']]", "copy items paths")
	flag.Parse()

	copyItems := make([][]string, 0)

	if err := json.Unmarshal([]byte(*copyItemsJSON), &copyItems); err != nil {
		panic("invalid copyItems: " + err.Error())
	}

	if err := util.UpdateConfig(copyItems, *src, *dst); err != nil {
		panic(err)
	}
}
