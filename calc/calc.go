package calc

import (
	"errors"
	"heat-transfer/constants"
)

// ac params struct
type ACParams struct {
	Enabled     bool
	OnTime      int
	OffTime     int
	SetTemp     float64
	CoolingPower float64
}

func CalculateTemperatureProfile(width, height, depth, insideTemp float64, outsideTemps [840]float64, heatTransferCoeff float64, acParams *ACParams) ([]float64, []float64, []bool) {

	volume := width * height * depth
	totalArea := 2 * height * (width + depth)

	Tref := insideTemp + 273.15
	rhoRef := 101325 / (287 * Tref)
	mass := rhoRef * volume

	cp := 1005.0

	totalMinutes := len(outsideTemps)
	Tfinal := float64(totalMinutes) * 60.0

	var acLowerBand, acUpperBand float64
	var acCompressorOn bool = false
	
	useAC := acParams != nil && acParams.Enabled
	if useAC {
		acLowerBand = acParams.SetTemp - 1.5
		acUpperBand = acParams.SetTemp + 1.5
		acCompressorOn = insideTemp > acParams.SetTemp
	}

	// ODE: dT/dt = (h * totalArea * (T_outside(t) - T) + acPower) / (mass * cp)
	f := func(t, Tcurr float64) float64 {
		// base value
		minuteOfSim := int(t / 60)
		idx := minuteOfSim
		if idx >= len(outsideTemps) {
			idx = len(outsideTemps) - 1
		}
		Toutside := outsideTemps[idx]
		heatFlow := heatTransferCoeff * totalArea * (Toutside - Tcurr)
		
		if useAC {
			isWithinOperatingTime := minuteOfSim >= acParams.OnTime && minuteOfSim < acParams.OffTime
			
			if isWithinOperatingTime {
				if acCompressorOn && Tcurr <= acLowerBand {
					acCompressorOn = false
				} else if !acCompressorOn && Tcurr >= acUpperBand {
					acCompressorOn = true
				}
				
				if acCompressorOn {
					heatFlow += acParams.CoolingPower
				}
			}
		}
		
		return heatFlow / (mass * cp)
	}

	// rk4 method
	dt := 10.0 // secs
	steps := int(Tfinal / dt)
	Tcurrent := insideTemp

	timeMinutes := make([]float64, 0, totalMinutes)
	insideProfile := make([]float64, 0, totalMinutes)
	acRunningProfile := make([]bool, 0, totalMinutes)
	nextRecordTime := 0.0

	for i := range steps {
		t := float64(i) * dt

		if t >= nextRecordTime {
			timeMinutes = append(timeMinutes, nextRecordTime/60.0)
			insideProfile = append(insideProfile, Tcurrent)
			acRunningProfile = append(acRunningProfile, acCompressorOn)
			nextRecordTime += 60.0
		}

		k1 := f(t, Tcurrent)
		k2 := f(t+dt/2, Tcurrent+dt*k1/2)
		k3 := f(t+dt/2, Tcurrent+dt*k2/2)
		k4 := f(t+dt, Tcurrent+dt*k3)
		Tcurrent = Tcurrent + (dt/6)*(k1+2*k2+2*k3+k4)
	}

	return timeMinutes, insideProfile, acRunningProfile
}

func CalculateMaterialCost(x, y, z, t, costPerM3 float64) float64 {
	s1 := x * y * t
	s2 := y * z * t

	return 2*(costPerM3*s1) + 2*(costPerM3*s2)
}

func CalculateCoeffByThickness(material string, thickness float64) (float64, error) {
	k, exists := constants.ThermalConductivity[material]
	if !exists {
		return 0, errors.New("material not found")
	}

	if thickness <= 0 {
		return 0, errors.New("thickness must be greater than zero")
	}

	// r val
	rValue := thickness / k

	// u value
	uValue := 1 / rValue

	return uValue, nil
}