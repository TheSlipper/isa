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
	InitXReal       string
	InitXBin        []byte
	ParentXBin      []byte
	CutPoint        int
	Offspring       []byte
	PostCrossover   []byte
	MutationIndeces []int
	PostMutation    []byte
	FinalXReal      string
	FinalFx         float64
}

// root pobiera plik strony root.html z dysku i prezentuje go przeglądarce.
func root(w http.ResponseWriter, r *http.Request) {
	// Get the GET params
	generate := true
	nStr, aStr, bStr, dStr := getGETParam("N", w, r), getGETParam("a", w, r), getGETParam("b", w, r), getGETParam("d", w, r)
	cpStr, mpStr := getGETParam("Pk", w, r), getGETParam("Pm", w, r)
	if nStr == "" || aStr == "" || bStr == "" || dStr == "" || cpStr == "" || mpStr == "" {
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

		// Crossover and save the results
		parents, offsprings, cutpoints, err := gas.Crossover(cp)
		if err != nil {
			throwErr(w, r, err, http.StatusInternalServerError)
			return
		}

		// Get the population after the crossover
		postCrossoverPop := gas.Population()

		// Mutate and save the results
		mut, err := gas.Mutate(mp)
		if err != nil {
			throwErr(w, r, err, http.StatusInternalServerError)
			return
		}

		// Get the population after the mutation
		postMutPop := gas.Population()

		// Populate the entries
		template := strings.Replace(fmt.Sprint("%."+strconv.Itoa(int(d))+"f"), " ", "", -1)
		for i := 0; i < N; i++ {
			results[i] = ExecResult{
				InitXReal:       fmt.Sprintf(template, rands[i]),
				InitXBin:        gas.XIntToXBin(uint32(gas.XRealToXInt(rands[i]))),
				ParentXBin:      parents[i],
				CutPoint:        cutpoints[i],
				Offspring:       offsprings[i],
				PostCrossover:   postCrossoverPop[i],
				MutationIndeces: mut[i],
				PostMutation:    postMutPop[i],
			}

			fXReal := gas.XIntToXReal(gas.XBinToXInt(postMutPop[i]))
			results[i].FinalXReal = fmt.Sprintf(template, fXReal)
			results[i].FinalFx = gas.Grade(fXReal)
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
