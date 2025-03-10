package gui

import (
	"bytes"
	"fmt"
	"heat-transfer/calc"
	"heat-transfer/chartings"
	"heat-transfer/constants"
	freader "heat-transfer/fReader"
	weatherdata "heat-transfer/weatherData"
	"math"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gonum.org/v1/plot/plotter"
)

var customMaterialInput *widget.Entry
var calculateButton *widget.Button

var currLocation, material string
var totalCost float64
var ACprofile []bool

type Result struct {
	Time    string
	InTemp  float64
	OutTemp float64

	timeWidget *widget.Label
	inTemp     *widget.Label
	outTemp    *widget.Label
}

type Results struct {
	InTemp  [840]float64
	OutTemp [840]float64
}

var thickness float64

type Params struct {
	W float64
	H float64
	L float64

	InsideTemp  *float64
	OutsideTemp [840]float64
	Location    string
	Season      string

	Coeff float64

	ACEnabled      bool
	ACOnTime       int
	ACOffTime      int
	ACSetTemp      float64
	ACCoolingPower float64
}

var results Result
var params Params

func StartGUILoop() {
	// default values so it doesnt panic on start
	params = Params{
		W:              5,
		H:              3,
		L:              4,
		InsideTemp:     nil,
		OutsideTemp:    sinusoidal840(),
		Coeff:          0.1,
		ACEnabled:      false,
		ACOnTime:       300,
		ACOffTime:      720,
		ACSetTemp:      25.0,
		ACCoolingPower: -3000.0,
	}

	temperature, err := weatherdata.GetCityTemperatureForecastNow("Khon+Kaen,+TH", freader.GetToken())
	if err != nil {
		fmt.Println(err)
	}

	resultsForDay := Results{
		InTemp:  [840]float64(temperature),
		OutTemp: [840]float64(temperature),
	}

	a := app.New()
	w := a.NewWindow("Heat Transfer Coefficient Calculator")

	w.SetOnClosed(func() {
		a.Quit()
		os.Exit(0)
	})

	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(650, 700))

	imageElem := canvas.NewImageFromReader(calculate(params.W, params.H, params.L, params.InsideTemp, [840]float64(temperature), params.Coeff), "chart.png")
	imageElem.FillMode = canvas.ImageFillOriginal

	// row 1
	label := widget.NewLabel("Material type")

	// material selector
	materialTypeSelector := widget.NewSelect([]string{"Wood", "Lightweight bricks", "Concrete", "Fibre glass", "PS foam", "PE foam", "Custom"}, selectHandler)
	materialTypeSelector.Resize(fyne.Size{Width: 150, Height: 10})
	materialTypeSelector.OnChanged = func(s string) {
		switch s {
		case "Wood":
			params.Coeff = constants.ThermalConductivity["wood"]
			material = "wood"
		case "Lightweight bricks":
			params.Coeff = constants.ThermalConductivity["brick"]
			material = "brick"
		case "Concrete":
			params.Coeff = constants.ThermalConductivity["concrete"]
			material = "concrete"
		case "Fibre glass":
			params.Coeff = constants.ThermalConductivity["fiberglass"]
			material = "fiberglass"
		case "PS foam":
			params.Coeff = constants.ThermalConductivity["ps_foam"]
			material = "ps_foam"
		case "PE foam":
			params.Coeff = constants.ThermalConductivity["pe_foam"]
			material = "pe_foam"
		case "Custom":
			v, err := strconv.ParseFloat(customMaterialInput.Text, 64)
			if err != nil {
				calculateButton.Disable()
			} else {
				calculateButton.Enable()
				params.Coeff = v
			}
		}

		customMaterialInput.SetText(fmt.Sprintf("%.2f", params.Coeff))
	}

	// custom
	customMaterialInput = widget.NewEntry()
	customMaterialInput.Disable()
	customMaterialInput.SetPlaceHolder("Coefficient")

	customMaterialInput.OnChanged = func(s string) {
		fmt.Println(s)
	}

	materialThickness := widget.NewEntry()
	materialThickness.SetPlaceHolder("Thickness (m)")
	materialThickness.OnChanged = func(s string) {
		v, err := strconv.ParseFloat(s, 64)
		thickness = v
		if err != nil {
			calculateButton.Disable()
		} else {
			selected := strings.ToLower(materialTypeSelector.Selected)
			selected = strings.ReplaceAll(selected, " ", "_")
			newCoeff, err := calc.CalculateCoeffByThickness(selected, v)
			if err != nil {
				calculateButton.Disable()
			} else {
				calculateButton.Enable()
				params.Coeff = newCoeff
				customMaterialInput.SetText(fmt.Sprintf("%.2f", newCoeff))
			}
		}
	}

	// row 2
	// room properties, Width, Height, Length

	label2 := widget.NewLabel("Room properties")

	// width
	roomWidth := widget.NewEntry()
	roomWidth.SetPlaceHolder("Width (m)")
	roomWidth.OnChanged = func(s string) {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			calculateButton.Disable()
		} else {
			calculateButton.Enable()
			params.W = v
		}
	}

	// height
	roomHeight := widget.NewEntry()
	roomHeight.SetPlaceHolder("Height (m)")
	roomHeight.OnChanged = func(s string) {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			calculateButton.Disable()
		} else {
			calculateButton.Enable()
			params.H = v
		}
	}

	// length
	roomLength := widget.NewEntry()
	roomLength.SetPlaceHolder("Length (m)")
	roomLength.OnChanged = func(s string) {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			calculateButton.Disable()
		} else {
			calculateButton.Enable()
			params.L = v
		}
	}

	// row 3

	// start conditions, inside and outside temperature,

	label3 := widget.NewLabel("Initial conditions")

	// inside temperature
	insideTemperature := widget.NewEntry()
	insideTemperature.SetPlaceHolder("In temp (°C)")
	insideTemperature.OnChanged = func(s string) {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil && s != "" {
			calculateButton.Disable()
		} else if s == "" {
			calculateButton.Enable()
			params.InsideTemp = &resultsForDay.OutTemp[0]
		} else {
			calculateButton.Enable()
			params.InsideTemp = &v
		}
	}

	// outside temperature
	outsideTemperature := widget.NewEntry()
	outsideTemperature.SetPlaceHolder("Location")
	outsideTemperature.OnChanged = func(s string) {
		params.Location = s
	}

	// row 4
	// ac settings

	acLabel := widget.NewLabel("AC Settings")

	// enable
	acEnable := widget.NewCheck("Enable AC", func(checked bool) {
		params.ACEnabled = checked
	})

	// ac temps
	acTempSetting := widget.NewEntry()
	acTempSetting.SetPlaceHolder("Set Temp (°C)")
	acTempSetting.OnChanged = func(s string) {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil && s != "" {
		} else if s != "" {
			params.ACSetTemp = v
		}
	}

	// time settings
	acOnTimeEntry := widget.NewEntry()
	acOnTimeEntry.SetPlaceHolder("On Time")
	acOnTimeEntry.OnChanged = func(s string) {
		if s != "" {
			timeParts := strings.Split(s, ":")
			if len(timeParts) == 2 {
				hour, hourErr := strconv.Atoi(timeParts[0])
				min, minErr := strconv.Atoi(timeParts[1])

				if hourErr == nil && minErr == nil && hour >= 0 && hour < 24 && min >= 0 && min < 60 {
					totalMinutes := (hour-5)*60 + min
					if totalMinutes < 0 {
						totalMinutes += 24 * 60
					}
					if totalMinutes < 840 {
						params.ACOnTime = totalMinutes
					}
				}
			}
		}
	}

	acOffTimeEntry := widget.NewEntry()
	acOffTimeEntry.SetPlaceHolder("Off Time")
	acOffTimeEntry.OnChanged = func(s string) {
		if s != "" {
			timeParts := strings.Split(s, ":")
			if len(timeParts) == 2 {
				hour, hourErr := strconv.Atoi(timeParts[0])
				min, minErr := strconv.Atoi(timeParts[1])

				if hourErr == nil && minErr == nil && hour >= 0 && hour < 24 && min >= 0 && min < 60 {
					totalMinutes := (hour-5)*60 + min
					if totalMinutes < 0 {
						totalMinutes += 24 * 60
					}
					if totalMinutes < 840 {
						params.ACOffTime = totalMinutes
					}
				}
			}
		}
	}

	// AC Cooling Power
	acPowerEntry := widget.NewEntry()
	acPowerEntry.SetPlaceHolder("Cooling Power (W)")
	acPowerEntry.OnChanged = func(s string) {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil && s != "" {
		} else if s != "" && v > 0 {
			params.ACCoolingPower = -v
		}
	}

	acTempSetting.SetText(fmt.Sprintf("%.1f", params.ACSetTemp))

	startHour := (params.ACOnTime / 60) + 5
	startMin := params.ACOnTime % 60
	acOnTimeEntry.SetText(fmt.Sprintf("%02d:%02d", startHour, startMin))

	endHour := (params.ACOffTime / 60) + 5
	endMin := params.ACOffTime % 60
	acOffTimeEntry.SetText(fmt.Sprintf("%02d:%02d", endHour, endMin))

	// * -1 to make it positive
	acPowerEntry.SetText(fmt.Sprintf("%.0f", -params.ACCoolingPower))

	costLabel := widget.NewLabel(fmt.Sprintf("%.2f THB", totalCost))

	montlyACCost := widget.NewLabel("0.00 THB")

	calculateButton = widget.NewButton("Calculate", func() {
		mat, _ := constants.GetMaterialCost(material)
		materialCost := calc.CalculateMaterialCost(params.W, params.H, params.L, thickness, mat)
		costLabel.SetText(fmt.Sprintf("%.2f THB", materialCost))
		if params.Location != currLocation {
			temperature, err = weatherdata.GetCityTemperatureForecastNow(params.Location, freader.GetToken())
			if err != nil {
				// create a pop up
				errPop := a.NewWindow("Error")
				errPop.SetContent(widget.NewLabel("Error: " + err.Error()))
				errPop.Show()
				return
			}
			currLocation = params.Location
		}

		if params.InsideTemp == nil {
			params.InsideTemp = &temperature[0]
		}

		if params.Location == "" {
			// error
			errPop := a.NewWindow("Error")
			errPop.SetContent(widget.NewLabel("Error: Location not set"))
			errPop.Show()
			return
		}

		// ac params
		var acParams *calc.ACParams
		if params.ACEnabled {
			acParams = &calc.ACParams{
				Enabled:      true,
				OnTime:       params.ACOnTime,
				OffTime:      params.ACOffTime,
				SetTemp:      params.ACSetTemp,
				CoolingPower: params.ACCoolingPower,
			}
		}

		// temp profile
		_, inTemp, acProfile := calc.CalculateTemperatureProfile(params.W, params.H, params.L, *params.InsideTemp, [840]float64(temperature), params.Coeff, acParams)
		resultsForDay.InTemp = [840]float64(inTemp)
		resultsForDay.OutTemp = [840]float64(temperature)

		// calculate cost
		monthlyCost, _ := calc.CalculateACCostForSimulation(acParams, acProfile, 0)
		montlyACCost.SetText(fmt.Sprintf("%.2f THB", monthlyCost))

		newChart := calculateWithAC(params.W, params.H, params.L, params.InsideTemp, [840]float64(temperature), params.Coeff, acParams).Bytes()
		imageElem.Resource = fyne.NewStaticResource("chart.png", newChart)
		imageElem.Refresh()
	})
	calculateButton.Disable()

	// time slider
	timeSlider := widget.NewSlider(0, 839)
	timeSlider.OnChanged = func(f float64) {
		results.Time = convertTime(f)

		results.timeWidget.SetText(results.Time)
		results.inTemp.SetText(fmt.Sprintf("%.1f °C", resultsForDay.InTemp[int(f)]))
		results.outTemp.SetText(fmt.Sprintf("%.1f °C", resultsForDay.OutTemp[int(f)]))
	}
	timeSlider.Resize(fyne.Size{Width: 525, Height: 25})

	// results
	results.timeWidget = widget.NewLabel(results.Time)
	results.inTemp = widget.NewLabel(fmt.Sprintf("%.1f °C", results.InTemp))
	results.outTemp = widget.NewLabel(fmt.Sprintf("%.1f °C", results.OutTemp))

	w.SetContent(container.NewVBox(
		// params
		container.NewGridWithColumns(
			4,

			label,
			materialTypeSelector,
			customMaterialInput,
			materialThickness,

			label2,
			roomWidth,
			roomHeight,
			roomLength,

			label3,
			insideTemperature,
			outsideTemperature,
			calculateButton,

			acLabel,
			acEnable,
			acTempSetting,
			container.NewGridWithColumns(
				2,
				acOnTimeEntry, acOffTimeEntry,
			),

			widget.NewLabel("AC Power"),
			acPowerEntry,
			widget.NewLabel("Monthly AC Cost"),
			montlyACCost,

			widget.NewLabel("Time"),
		),
		// slider
		container.NewWithoutLayout(
			timeSlider,
		),

		// results
		container.NewGridWithColumns(
			4,

			// time, in temp, out temp
			widget.NewLabel("Time"),
			widget.NewLabel("In temp"),
			widget.NewLabel("Out temp"),
			widget.NewLabel("Total cost"),

			// results
			results.timeWidget,
			results.inTemp,
			results.outTemp,
			costLabel,
		),

		// graph
		container.NewVBox(
			widget.NewLabel("Graph"),
			imageElem,
		),
	))

	OnChangeHandler()
	w.ShowAndRun()
}

func selectHandler(selected string) {
	calculateButton.Enable()
	if selected == "Custom" {
		customMaterialInput.Enable()
	} else {
		customMaterialInput.Disable()
	}
	fmt.Println("Selected:", selected)
}

func OnChangeHandler() {

}

func convertTime(step float64) string {
	// 0 to 840 is mapped to 05:00 to 19:00

	hour := int(step / 60)
	minute := int(step) % 60

	return fmt.Sprintf("%02d:%02d", hour+5, minute)
}

func calculate(w, h, d float64, insideTemp *float64, outsideTemps [840]float64, coeff float64) *bytes.Buffer {
	return calculateWithAC(w, h, d, insideTemp, outsideTemps, coeff, nil)
}

func calculateWithAC(w, h, d float64, insideTemp *float64, outsideTemps [840]float64, coeff float64, acParams *calc.ACParams) *bytes.Buffer {
	if insideTemp == nil {
		insideTemp = &outsideTemps[0]
	}

	// get the inside temperature profile.
	timeMinutes, insideProfile, acProfile := calc.CalculateTemperatureProfile(w, h, d, *insideTemp, outsideTemps, coeff, acParams)

	ACprofile = acProfile

	insidePts := make(plotter.XYs, len(timeMinutes))
	outsidePts := make(plotter.XYs, len(timeMinutes))
	for i, t := range timeMinutes {
		insidePts[i].X = t
		insidePts[i].Y = insideProfile[i]

		outsidePts[i].X = t
		outsidePts[i].Y = outsideTemps[i]

		// doesnt work
		// calculate the area under the curve for AC usage
		// if acProfile != nil {
		// }
	}

	// generate chart
	imgData, _ := chartings.Plot(insidePts, outsidePts)
	return imgData
}

// for testing
func constant840() [840]float64 {
	var arr [840]float64
	for i := range 840 {
		arr[i] = 30
	}
	return arr
}

// for testing
func sinusoidal840() [840]float64 {
	var arr [840]float64
	for i := range 840 {
		arr[i] = 30 + 5*math.Sin(4*math.Pi*float64(i)/840)
	}
	return arr
}
