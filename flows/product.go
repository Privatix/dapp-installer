package flows

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/privatix/dapp-installer/dapp"
	"github.com/privatix/dapp-installer/product"
	"github.com/privatix/dapp-installer/util"
)

func installProducts(d *dapp.Dapp) error {
	if d.OnlyCore {
		return nil
	}

	conn := d.DBEngine.DB.ConnectionString()
	if err := product.Install(d.Role, d.Path, conn, d.Product); err != nil {
		return fmt.Errorf("failed to install products: %v", err)
	}

	return nil
}

func removeProducts(d *dapp.Dapp) error {
	if d.OnlyCore {
		return nil
	}

	if err := product.Remove(d.Role, d.Path, d.Product); err != nil {
		return fmt.Errorf("failed to remove products: %v", err)
	}

	return nil
}

func startProducts(d *dapp.Dapp) error {
	if err := product.Start(d.Role, d.Path, d.Product); err != nil {
		return fmt.Errorf("failed to start products: %v", err)
	}

	return nil
}

func updateProducts(d *dapp.Dapp) error {
	if len(d.Source) == 0 {
		return fmt.Errorf("product path not set")
	}

	err := os.Setenv("PRIVATIX_TEMP_PRODUCT", d.Source)
	if err != nil {
		return fmt.Errorf("failed to set env variables: %v", err)
	}

	defer os.Setenv("PRIVATIX_TEMP_PRODUCT", "")

	err = product.Update(d.Role, d.Path, d.Source, d.Product)
	if err != nil {
		return fmt.Errorf("failed to update products: %v", err)
	}

	util.CopyDir(filepath.Join(d.Source, "product"),
		filepath.Join(d.Path, "product"))

	return nil
}

func stopProducts(d *dapp.Dapp) error {
	if err := product.Stop(d.Role, d.Path, d.Product); err != nil {
		return fmt.Errorf("failed to stop products: %v", err)
	}

	return nil
}
