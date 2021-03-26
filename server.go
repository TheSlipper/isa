package main

import (
	"fmt"
	"html/template"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/TheSlipper/isa/evolalg"
)

type ExecResult struct {
	XReal     string
	XInt      int
	XBin      []byte
	XIntConv  int
	XRealConv string
	Grade     float64
}

// root pobiera plik strony root.html z dysku i prezentuje go przeglądarce.
func root(w http.ResponseWriter, r *http.Request) {
	// Get the GET params
	generate := true
	nStr, aStr, bStr, dStr := getGETParam("N", w, r), getGETParam("a", w, r), getGETParam("b", w, r), getGETParam("d", w, r)
	if nStr == "" || aStr == "" || bStr == "" || dStr == "" {
		generate = false
	}

	// Generate the values if all the necessary data was given
	var results []ExecResult
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

		// function and solver for this assignment
		// F(x)= x MOD1 *(COS(20*π *x)–SIN(x))
		// d = 0,001 -> 3
		gradingFunc := func(x float64) float64 {
			return math.Mod(x, 1*(math.Cos(20*math.Pi*x)-math.Sin(x)))
		}

		gas, err := evolalg.NewGeneticAlgorithmSolver(a, b, d)
		if err != nil {
			throwErr(w, r, err, http.StatusInternalServerError)
			return
		}

		// Generate random xreals in the range of the function
		// res := make([]ExecResults, N)
		for i := 0; i < N; i++ {
			// template for floating point representation
			temp := strings.Replace(fmt.Sprint("%."+strconv.Itoa(int(d))+"f"), " ", "", -1)

			xreal := a + rand.Float64()*(b-a)
			xreal = math.Round(xreal*math.Pow10(int(d))) / math.Pow10(int(d))
			xint := gas.XRealToXInt(xreal)
			xbin := gas.XIntToXBin(uint32(xint))
			xintConv := gas.XBinToXInt(xbin)
			xrealConv := gas.XIntToXReal(xintConv)
			xrealConv = math.Round(xrealConv*math.Pow10(int(d))) / math.Pow10(int(d))
			grade := gradingFunc(xreal)

			res := ExecResult{
				XReal:     fmt.Sprintf(temp, xreal),
				XInt:      xint,
				XBin:      xbin,
				XIntConv:  xintConv,
				XRealConv: fmt.Sprintf(temp, xreal),
				Grade:     grade,
			}
			results = append(results, res)
		}

	}

	// generate template and process it
	t, err := template.ParseFiles("root.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "InternalServerError")
		fmt.Println(err.Error())
		return
	}
	err = t.Execute(w, results)
	if err != nil {
		fmt.Println(err.Error())
	}
}
