package constants

import "errors"

// Thermal conductivity values in W/(m·K)
var ThermalConductivity = map[string]float64{
	"wood":       0.12,
	"brick":      0.8,
	"concrete":   1.8,
	"fiberglass": 0.04,
	"ps_foam":    0.035,
	"pe_foam":    0.04,
}

// Average cost per cubic meter in Thai Baht (THB)
var materialCosts = map[string]float64{
	"wood":       22_500, // Average of 15,000 - 30,000 THB
	"brick":      5_250,  // Average of 3,500 - 7,000 THB
	"concrete":   4_000,  // Average of 3,000 - 5,000 THB
	"fiberglass": 1_500,  // Average of 1,000 - 2,000 THB
	"ps_foam":    2_250,  // Average of 1,500 - 3,000 THB
	"pe_foam":    3_000,  // Average of 2,000 - 4,000 THB
}

// GetMaterialCost retrieves the average cost per cubic meter for a given material.
func GetMaterialCost(material string) (float64, error) {
	cost, exists := materialCosts[material]
	if !exists {
		return 0, errors.New("material not found")
	}
	return cost, nil
}
