package flows

import (
	"fmt"

	"github.com/privatix/dapp-installer/dapp"
	"github.com/privatix/dapp-installer/product"
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
	if d.OnlyCore {
		return nil
	}

	if err := product.Start(d.Role, d.Path, d.Product); err != nil {
		return fmt.Errorf("failed to start products: %v", err)
	}

	return nil
}

func stopProducts(d *dapp.Dapp) error {
	if d.OnlyCore {
		return nil
	}

	if err := product.Stop(d.Role, d.Path, d.Product); err != nil {
		return fmt.Errorf("failed to stop products: %v", err)
	}

	return nil
}
