package ipmi

import "strings"

type FRU struct {
	Board   `json:"board"`
	Product `json:"product"`
}

type Board struct {
	ManufacturingDate string `json:"manufacturingDate"`
	Manufacturer      string `json:"manufacturer"`
	Product           string `json:"product"`
	Serial            string `json:"serial"`
	PartNumber        string `json:"partNumber"`
}

type Product struct {
	Manufacturer string `json:"manufacturer"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	Serial       string `json:"serial"`
}

func (h *Helper) parseFRU(out []byte) (*FRU, error) {
	fru := FRU{}
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Board Mfg Date":
			fru.Board.ManufacturingDate = value
		case "Board Mfg":
			fru.Board.Manufacturer = value
		case "Board Product":
			fru.Board.Product = value
		case "Board Serial":
			fru.Board.Serial = value
		case "Board Part Number":
			fru.Board.PartNumber = value
		case "Product Manufacturer":
			fru.Product.Manufacturer = value
		case "Product Name":
			fru.Product.Name = value
		case "Product Version":
			fru.Product.Version = value
		case "Product Serial":
			fru.Product.Serial = value
		}
	}

	return &fru, nil
}
