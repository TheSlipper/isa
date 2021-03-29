package main

import (
	"fmt"
	"html/template"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/TheSlipper/isa/evolalg"
)

type ExecResult struct {
	XReal  string
	Fx     float64
	Gx     float64
	Px     float64
	Qx     float64
	R      float64
	XReal2 int
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

		// Run a selection
		results = make([]ExecResult, N)
		rands := make([]float64, N)
		rand.Seed(time.Now().UnixNano())
		for i := 0; i < N; i++ {
			rands[i] = a + rand.Float64()*(b-a)
			rands[i] = math.Round(rands[i]*math.Pow10(int(d))) / math.Pow10(int(d)) // TODO Sprawdzić czy to potrzebne
		}
		err = gas.Selection(rands...)
		if err != nil {
			throwErr(w, r, err, http.StatusInternalServerError)
			return
		}

		// Get the results calculated so far and store them in the array
		template := strings.Replace(fmt.Sprint("%."+strconv.Itoa(int(d))+"f"), " ", "", -1)
		grades, fits, probs, probsHB := gas.Cache()
		for i := 0; i < N; i++ {
			results[i] = ExecResult{
				XReal: fmt.Sprintf(template, rands[i]),
				Fx:    grades[i],
				Gx:    fits[i],
				Px:    probs[i],
				Qx:    probsHB[i],
			}
		}

		// Generate new random numbers and show which subset it fits in
		rand.Seed(time.Now().Add(5 * time.Second).UnixNano())
		for i := 0; i < N; i++ {
			r := rand.Float64()
			results[i].R = r
			for j := 0; j < N; j++ {
				if r <= probsHB[j] {
					results[i].XReal2 = j
					break
				}
			}
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
