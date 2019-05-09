package product

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sethvargo/go-password/password"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

const (
	templatePath = "templates"
	productPath  = "products"

	offeringTemplate = "offering.json"
	accessTemplate   = "access.json"

	serverProduct = "server.json"
	clientProduct = "client.json"

	adapterConfig       = "adapter.config.json"
	agentAdapterConfig  = "adapter.agent.config.json"
	clientAdapterConfig = "adapter.client.config.json"

	jsonIdent = "    "

	passwordLength = 12
	saltLength     = 9 * 1e18
)

// Errors.
var (
	ErrNotAssociated = fmt.Errorf("product is not associated with the template")
	ErrNotFile       = fmt.Errorf("object is not file")
	ErrNotAllItems   = fmt.Errorf("some required items not found")
)

type item struct {
	product *data.Product
	config  string
}

func processor(dir string, adjust bool, tx *reform.TX) error {
	srvProduct, cliProduct, err := handler(dir, tx)
	if err != nil {
		return err
	}

	if adjust {
		items := []*item{
			{srvProduct, filepath.Join(dir, agentAdapterConfig)},
			{cliProduct, filepath.Join(dir, clientAdapterConfig)},
		}

		for k := range items {
			err = adjustment(items[k].product, items[k].config)
			if err != nil {
				return err
			}
		}
	}

	for _, product := range []*data.Product{srvProduct, cliProduct} {
		err = importProduct(tx, product)
		if err != nil {
			return err
		}
	}

	return nil
}

func handler(dir string, tx *reform.TX) (srvProduct,
	cliProduct *data.Product, err error) {
	err = validateRoot(dir)
	if err != nil {
		return nil, nil, err
	}

	offerTplFile := filepath.Join(dir, templatePath, offeringTemplate)
	accessTplFile := filepath.Join(dir, templatePath, accessTemplate)

	offerTpl, accessTpl, err := templates(tx, offerTplFile, accessTplFile)
	if err != nil {
		return nil, nil, err
	}

	serverProductFile := filepath.Join(dir, productPath, serverProduct)
	clientProductFile := filepath.Join(dir, productPath, clientProduct)

	return products(serverProductFile, clientProductFile, offerTpl.ID, accessTpl.ID)
}

func adjustment(product *data.Product, configFile string) error {
	pass, err := setProductAuth(product)
	if err != nil {
		return err
	}

	cfg := make(map[string]interface{})

	err = util.ReadJSONFile(configFile, &cfg)
	if err != nil {
		return err
	}

	sess, ok := cfg["Sess"]
	if !ok {
		sess = make(map[string]interface{})
		cfg["Sess"] = sess
	}
	sess.(map[string]interface{})["Product"] = product.ID
	sess.(map[string]interface{})["Password"] = pass

	return util.WriteJSONFile(configFile, "", jsonIdent, &cfg)
}

func templates(tx *reform.TX, offer,
	access string) (offerTpl, accessTpl *data.Template, err error) {
	offerTpl, err = importTemplate(offer, data.TemplateOffer, tx)
	if err != nil {
		return nil, nil, err
	}

	accessTpl, err = importTemplate(access, data.TemplateAccess, tx)
	if err != nil {
		return nil, nil, err
	}
	return offerTpl, accessTpl, err
}

func products(serverFile, clientFile, templateID, accessID string) (srvProduct,
	cliProduct *data.Product, err error) {
	srvProduct, err = productFromFile(serverFile, templateID, accessID)
	if err != nil {
		return nil, nil, err
	}

	if !productConcord(srvProduct, templateID) {
		return nil, nil, ErrNotAssociated
	}

	cliProduct, err = productFromFile(clientFile, templateID, accessID)
	if err != nil {
		return nil, nil, err
	}

	if !productConcord(cliProduct, templateID) {
		return nil, nil, ErrNotAssociated
	}
	return srvProduct, cliProduct, err
}

func validateDir(name string, expect map[string]bool) error {
	info, err := os.Stat(name)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("%s - is not a directory", name)
	}

	dir, err := os.Open(name)
	if err != nil {
		return err
	}
	defer dir.Close()

	items, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}

	isFile := func(name string) bool {
		stat, err := os.Stat(name)
		if err != nil {
			return false
		}
		return !stat.IsDir()
	}

	for _, v := range items {
		if expect[v] && !isFile(filepath.Join(name, v)) {
			return ErrNotFile
		}

		delete(expect, v)
	}

	if len(expect) != 0 {
		return ErrNotAllItems
	}
	return nil
}

func validateRoot(dir string) error {
	rootItems := map[string]bool{
		templatePath:        false,
		productPath:         false,
		agentAdapterConfig:  true,
		clientAdapterConfig: true,
	}

	err := validateDir(dir, rootItems)
	if err != nil {
		return err
	}

	tplItems := map[string]bool{
		offeringTemplate: true,
		accessTemplate:   true,
	}

	err = validateDir(filepath.Join(dir, templatePath), tplItems)
	if err != nil {
		return err
	}

	productItems := map[string]bool{
		serverProduct: true,
		clientProduct: true,
	}

	return validateDir(filepath.Join(dir, productPath), productItems)
}

func productConcord(product *data.Product, tplID string) bool {
	if product.OfferTplID == nil {
		return false
	}
	return *product.OfferTplID == tplID
}

func importTemplate(file, kind string, tx *reform.TX) (*data.Template, error) {
	// Reading to map[string]interface{} and marshaling it compreses
	// initial json by removing spaces and newlines which is what should be used
	// to produce correct hash.
	var schema map[string]interface{}

	err := util.ReadJSONFile(file, &schema)
	if err != nil { 
		return nil, err
	}

	schemaB, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}

	template := new(data.Template)
	template.ID = util.NewUUID()
	template.Raw = schemaB
	template.Hash = data.HexFromBytes(crypto.Keccak256([]byte(schemaB)))
	template.Kind = kind

	err = tx.Insert(template)
	if err != nil {
		return nil, err
	}
	return template, err
}

func importProduct(tx *reform.TX, product *data.Product) error {
	return tx.Insert(product)
}

func productFromFile(file, offerID, accessID string) (product *data.Product, err error) {
	err = util.ReadJSONFile(file, &product)
	product.OfferTplID = &offerID
	product.OfferAccessID = &accessID
	return product, err
}

func setProductAuth(product *data.Product) (string, error) {
	product.ID = util.NewUUID()

	salt, err := rand.Int(rand.Reader, big.NewInt(saltLength))
	if err != nil {
		return "", err
	}

	n, _ := rand.Int(rand.Reader, big.NewInt(10))
	pass, _ := password.Generate(passwordLength, int(n.Int64()), 0, false, false)

	passwordHash, err := data.HashPassword(pass, string(salt.Uint64()))
	if err != nil {
		return "", err
	}

	product.Password = passwordHash
	product.Salt = salt.Uint64()

	return pass, nil
}
