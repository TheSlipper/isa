package main

import (
	"encoding/json"
	"html/template"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/TheSlipper/isa/evolalg"
	chart "github.com/wcharczuk/go-chart/v2"
)

// root pobiera plik strony root.html z dysku i prezentuje go przeglądarce.
func root(w http.ResponseWriter, r *http.Request) {
	// Get the GET params
	generate := true
	nStr, aStr, bStr, dStr := getGETParam("N", w, r), getGETParam("a", w, r), getGETParam("b", w, r), getGETParam("d", w, r)
	cpStr, mpStr, epochsStr := getGETParam("Pk", w, r), getGETParam("Pm", w, r), getGETParam("epoki", w, r)
	if nStr == "" || aStr == "" || bStr == "" || dStr == "" || cpStr == "" || mpStr == "" || epochsStr == "" {
		generate = false
	}
	jsonFormat := getGETParam("json", w, r)

	// Generate the values if all the necessary data was given
	var hist []evolalg.EpochData
	if generate {
		// Convert to values
		N, err := strconv.Atoi(nStr)
		if err != nil {
			throwErr(w, r, err, http.StatusInternalServerError)
			return
		}
		dInt, err := strconv.Atoi(dStr)
		if err != nil {
			throwErr(w, r, err, http.StatusInternalServerError)
			return
		}
		d := byte(dInt)
		a, err := strconv.ParseFloat(aStr, 64)
		if err != nil {
			throwErr(w, r, err, http.StatusInternalServerError)
			return
		}
		b, err := strconv.ParseFloat(bStr, 64)
		if err != nil {
			throwErr(w, r, err, http.StatusInternalServerError)
			return
		}
		cp, err := strconv.ParseFloat(cpStr, 64)
		if err != nil {
			throwErr(w, r, err, http.StatusInternalServerError)
			return
		}
		mp, err := strconv.ParseFloat(mpStr, 64)
		if err != nil {
			throwErr(w, r, err, http.StatusInternalServerError)
			return
		}
		epochs, err := strconv.Atoi(epochsStr)
		if err != nil {
			throwErr(w, r, err, http.StatusInternalServerError)
			return
		}

		// solver for this assignment with a grading function described by this formula:
		// F(x)= x MOD1 *(COS(20*π *x)–SIN(x))
		// d = 0,001 -> 3
		gas, err := evolalg.NewGeneticAlgorithmSolver(a, b, d, func(x float64) float64 {
			return math.Mod(x, 1*(math.Cos(20*math.Pi*x)-math.Sin(x)))
		})
		if err != nil {
			throwErr(w, r, err, http.StatusInternalServerError)
			return
		}

		hist, err = gas.Solve(N, epochs, cp, mp)
		if err != nil {
			throwErr(w, r, err, http.StatusInternalServerError)
			return
		}

		// Create the fmax, favg, fmin graph
		fmax, favg, fmin := make([]float64, len(hist)), make([]float64, len(hist)), make([]float64, len(hist))
		epochsArr := make([]float64, len(hist))
		ticker := len(hist) / 20
		epochTicks := []chart.Tick{}
		i := 0
		for ; i < len(hist); i++ {
			fmax[i] = hist[i].FMax
			favg[i] = hist[i].FAVG
			fmin[i] = hist[i].FMin
			epochsArr[i] = float64(i)

			if len(hist) < 20 {
				epochTicks = append(epochTicks, chart.Tick{Value: float64(i), Label: strconv.Itoa(i)})
			} else if i%ticker == 0 {
				epochTicks = append(epochTicks, chart.Tick{Value: float64(i), Label: strconv.Itoa(i)})
			}
		}
		if epochTicks[len(epochTicks)-1].Value != float64(i) {
			epochTicks = append(epochTicks, chart.Tick{Value: float64(epochs),
				Label: strconv.Itoa(epochs)})
		}

		graph := chart.Chart{
			XAxis: chart.XAxis{
				Name: "Epoka",
				Range: &chart.ContinuousRange{
					Min: 0.0,
					Max: float64(epochs),
				},
				Ticks: epochTicks,
			},
			YAxis: chart.YAxis{
				Name: "f(x)",
			},
			Series: []chart.Series{
				chart.ContinuousSeries{
					Name: "fmax",
					Style: chart.Style{
						StrokeColor: chart.GetDefaultColor(0).WithAlpha(64),
						StrokeWidth: 3.5,
					},
					XValues: epochsArr,
					YValues: fmax,
				},
				chart.ContinuousSeries{
					Name: "favg",
					Style: chart.Style{
						StrokeColor: chart.GetDefaultColor(1).WithAlpha(64),
						StrokeWidth: 3.5,
					},
					XValues: epochsArr,
					YValues: favg,
				},
				chart.ContinuousSeries{
					Name: "fmin",
					Style: chart.Style{
						StrokeColor: chart.GetDefaultColor(2).WithAlpha(64),
						StrokeWidth: 3.5,
					},
					XValues: epochsArr,
					YValues: fmin,
				},
			},
		}

		os.Remove("static/fmax_favg_fmin.svg")
		f, _ := os.Create("static/fmax_favg_fmin.svg")
		defer f.Close()
		graph.Render(chart.SVG, f)
	}

	// generate template and process it
	if jsonFormat == "" {
		t, err := template.ParseFiles("root.html")
		if err != nil {
			throwErr(w, r, err, http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, hist)
		if err != nil {
			throwErr(w, r, err, http.StatusInternalServerError)
			return
		}
	} else {
		byteArr, err := json.Marshal(hist)
		if err != nil {
			throwErr(w, r, err, http.StatusInternalServerError)
			return
		}
		w.Write(byteArr)
		w.WriteHeader(200)
	}
}
