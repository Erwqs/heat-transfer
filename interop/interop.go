package interop

import "math"

func MovingWindowInterpolateTemperature(hourlyTemps []float64) [840]float64 {
	minutelyTemps := [840]float64{}

	expandedTemps := make([]float64, len(hourlyTemps)+4)

	if len(hourlyTemps) >= 2 {
		slope := hourlyTemps[1] - hourlyTemps[0]
		expandedTemps[0] = hourlyTemps[0] - 2*slope
		expandedTemps[1] = hourlyTemps[0] - slope
	} else {
		expandedTemps[0] = hourlyTemps[0]
		expandedTemps[1] = hourlyTemps[0]
	}

	copy(expandedTemps[2:2+len(hourlyTemps)], hourlyTemps)

	if len(hourlyTemps) >= 2 {
		slope := hourlyTemps[len(hourlyTemps)-1] - hourlyTemps[len(hourlyTemps)-2]
		expandedTemps[len(hourlyTemps)+2] = hourlyTemps[len(hourlyTemps)-1] + slope
		expandedTemps[len(hourlyTemps)+3] = hourlyTemps[len(hourlyTemps)-1] + 2*slope
	} else {
		expandedTemps[len(hourlyTemps)+2] = hourlyTemps[len(hourlyTemps)-1]
		expandedTemps[len(hourlyTemps)+3] = hourlyTemps[len(hourlyTemps)-1]
	}

	// cubic spline
	for minute := 0; minute < 840; minute++ {
		hourIndex := minute / 60
		minuteInHour := minute % 60

		// first and last hr
		if hourIndex >= len(hourlyTemps)-1 {
			minutelyTemps[minute] = hourlyTemps[len(hourlyTemps)-1]
			continue
		}

		// cubic interop
		t := float64(minuteInHour) / 60.0

		// control points
		idx := hourIndex + 2
		p0 := expandedTemps[idx-1]
		p1 := expandedTemps[idx]
		p2 := expandedTemps[idx+1]
		p3 := expandedTemps[idx+2]

		// apply
		t2 := t * t
		t3 := t2 * t

		// hermite basis fn
		h00 := 2*t3 - 3*t2 + 1
		h10 := t3 - 2*t2 + t
		h01 := -2*t3 + 3*t2
		h11 := t3 - t2

		// tangent, catmull rom
		m0 := 0.5 * (p2 - p0)
		m1 := 0.5 * (p3 - p1)

		// interpolate
		interpolatedTemp := h00*p1 + h10*m0 + h01*p2 + h11*m1

		// more smooth
		if minuteInHour < 5 || minuteInHour > 55 {
			if hourIndex > 0 && hourIndex < len(hourlyTemps)-2 {
				windowAvg := (expandedTemps[idx-1] + expandedTemps[idx] +
					expandedTemps[idx+1] + expandedTemps[idx+2]) / 4.0

				// blend weight
				blendFactor := 0.0
				if minuteInHour < 5 {
					blendFactor = 1.0 - float64(minuteInHour)/5.0
				} else {
					blendFactor = (float64(minuteInHour) - 55.0) / 5.0
				}

				// blend with window avg
				interpolatedTemp = interpolatedTemp*(1-blendFactor*0.3) + windowAvg*(blendFactor*0.3)
			}
		}

		minutelyTemps[minute] = interpolatedTemp
	}

	// last pass
	smoothedTemps := [840]float64{}
	copy(smoothedTemps[:], minutelyTemps[:])

	windowSize := 30
	for i := windowSize; i < 840-windowSize; i++ {
		sum := 0.0
		weightSum := 0.0

		for j := -windowSize; j <= windowSize; j++ {
			weight := math.Exp(-float64(j*j) / (2.0 * float64(windowSize/3) * float64(windowSize/3)))
			sum += minutelyTemps[i+j] * weight
			weightSum += weight
		}

		smoothedTemps[i] = sum / weightSum
	}

	return smoothedTemps
}
