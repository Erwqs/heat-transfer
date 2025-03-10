package calc

import (
	"math"
	"time"
)

type ElectricityRate struct {
	Blocks     []float64
	BlockRates []float64 // bath/kwh
	ServiceFee float64
	FtRate     float64
	VatPercent float64
}

func GetResidentialRate() ElectricityRate {
	// 2024 data
	return ElectricityRate{
		Blocks:     []float64{15, 25, 35, 100, 150, 400, math.Inf(1)},
		BlockRates: []float64{2.3488, 2.9882, 3.2405, 3.6237, 3.7171, 4.2218, 4.4217},
		ServiceFee: 38.22,
		FtRate:     0.6889,
		VatPercent: 7.0, // tax
	}
}

func CalculateACElectricityCost(
	acParams *ACParams,
	daysInMonth int,
	existingUsage float64,
	acRunningProfile []bool,
) (float64, []bool) {
	if acParams == nil || !acParams.Enabled || len(acRunningProfile) == 0 {
		return 0.0, nil
	}

	if daysInMonth <= 0 {
		currentTime := time.Now()
		currentYear, currentMonth, _ := currentTime.Date()
		daysInMonth = time.Date(currentYear, currentMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()
	}

	totalACMinutes := 0
	for _, running := range acRunningProfile {
		if running {
			totalACMinutes++
		}
	}

	dailyKWh := (math.Abs(acParams.CoolingPower) / 1000.0) * (float64(totalACMinutes) / 60.0)
	monthlyKWh := dailyKWh * float64(daysInMonth)
	totalMonthlyKWh := existingUsage + monthlyKWh

	rate := GetResidentialRate()

	energyCharge := 0.0
	remainingKWh := totalMonthlyKWh

	for i := range rate.Blocks {
		var blockKWh float64
		if i == 0 {
			blockKWh = math.Min(remainingKWh, rate.Blocks[i])
		} else if i < len(rate.Blocks)-1 {
			blockKWh = math.Min(remainingKWh, rate.Blocks[i]-rate.Blocks[i-1])
		} else {
			blockKWh = remainingKWh
		}

		if blockKWh <= 0 {
			break
		}

		energyCharge += blockKWh * rate.BlockRates[i]
		remainingKWh -= blockKWh
	}

	acProportion := monthlyKWh / totalMonthlyKWh
	acEnergyCharge := energyCharge * acProportion
	ftCharge := monthlyKWh * rate.FtRate
	serviceCharge := rate.ServiceFee * acProportion

	// tax
	subtotal := acEnergyCharge + ftCharge + serviceCharge
	vat := subtotal * (rate.VatPercent / 100.0)
	totalCost := subtotal + vat

	return totalCost, acRunningProfile
}

func EstimateACOperatingCost(acParams *ACParams, acRunningProfile []bool, existingUsage float64) (float64, float64, float64, []bool) {
	if acParams == nil || !acParams.Enabled || len(acRunningProfile) == 0 {
		return 0.0, 0.0, 0.0, nil
	}

	currentTime := time.Now()
	currentYear, currentMonth, _ := currentTime.Date()
	daysInMonth := time.Date(currentYear, currentMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()

	monthlyCost, acRunningProfile := CalculateACElectricityCost(acParams, daysInMonth, existingUsage, acRunningProfile)

	dailyCost := monthlyCost / float64(daysInMonth)

	operatingHours := float64(acParams.OffTime-acParams.OnTime) / 60.0
	hourlyCost := 0.0
	if operatingHours > 0 {
		hourlyCost = dailyCost / operatingHours
	}

	return hourlyCost, dailyCost, monthlyCost, acRunningProfile
}

func CalculateACCostForSimulation(acParams *ACParams, acRunningProfile []bool, existingUsage float64) (float64, []bool) {
	if acParams == nil || !acParams.Enabled || len(acRunningProfile) == 0 {
		return 0.0, nil
	}

	monthlyCost, _ := CalculateACElectricityCost(acParams, 30, existingUsage, acRunningProfile)

	return monthlyCost, acRunningProfile
}

func CalculateMonthlyCost(p ACParams, existingUsage float64, acProfile []bool) float64 {
	_, _, monthlyCost, _ := EstimateACOperatingCost(&p, acProfile, existingUsage)

	return monthlyCost
}
