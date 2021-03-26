package evolalg

import (
	"errors"
	"math"
)

// GeneticAlgorithmSolver is a struct that contains all of the data related to a generic algorithm instance and shares
// a set of functions for solving this certain genetic algorithm.
type GeneticAlgorithmSolver struct {
	a       float64 // lower bound of the set (inclusive)
	b       float64 // higher bound of the set (inclusive)
	d       byte    // accuracy (e.g. accuracy 3 means 10^3 for each decimal therefore for 3 it would generate a 3001 population)
	l       int     // minimal bit size for representation of all of the population
	popSize uint    // actual population size (higher bound of <0, pop> set)
	popArr  []byte  // population array in little endian
}

// NewGeneticAlgorithm creates a new instance of a genetic algorithm solver.
//
// Arguments:
//
// a - lower bound of the set (inclusive)
//
// b - higher bound of the set (inclusive)
//
// d - accuracy (e.g. accuracy 3 means 10^3 for each decimal therefore for 3 it would generate a 3001 population)
func NewGeneticAlgorithmSolver(a float64, b float64, d byte) (ga GeneticAlgorithmSolver, err error) {
	if a > b {
		err = errors.New("provided lower bound greater than higher bound")
		return
	} else if d < 1 {
		err = errors.New("provided precision is equal to or lower than zero")
		return
	}

	ga.a = a
	ga.b = b
	ga.d = d

	// get the population size
	var lb, hb float64
	if ga.a < 0 {
		lb = 0
		hb = ga.b + math.Abs(ga.a)
	} else {
		lb = ga.a
		hb = ga.b
	}
	popSize := ((hb - lb) * (math.Pow(10, float64(ga.d)))) + 1
	if popSize-math.Floor(popSize) != 0 { // TODO Co w tym przypadku ? na razie daje error
		err = errors.New("provided precision was not big enough")
	}
	ga.popSize = uint(popSize)

	// calculate size of the population array and create it
	ga.l = int(math.Ceil(math.Log2((ga.b-ga.a)*(1/math.Pow(10, -float64(ga.d))) + 1)))
	ga.popArr = make([]byte, ga.l)

	return
}

// XBinToXInt converts x in binary form to x in integer form.
func (gas GeneticAlgorithmSolver) XBinToXInt(arr []byte) int {
	res := 0
	j := 0
	for i := len(arr) - 1; i >= 0; i-- {
		if arr[i] == 1 {
			res += int(math.Pow(2, float64(j)))
		}
		j++
	}
	return res
}

// XIntToXBin converts x in integer form to x in little endian binary form.
// func (gas GeneticAlgorithmSolver) XIntToXBin(val uint32, size int) []byte {
func (gas GeneticAlgorithmSolver) XIntToXBin(val uint32) []byte {
	arr := []byte{}

	for val != 0 {
		arr = append(arr, byte(val%2))
		val = val / 2
	}

	if len(arr) < gas.l {
		addLen := gas.l - len(arr)
		zeroArr := make([]byte, addLen)
		arr = append(arr, zeroArr...)
	}

	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}

	return arr
}

// XIntToXReal converts x in integer form to x in floating point form.
func (gas GeneticAlgorithmSolver) XIntToXReal(xint int) float64 {
	return gas.a + ((gas.b - gas.a) * float64(xint) / (math.Pow(2, float64(gas.l)) - 1))
}

// XRealToXInt converts x in floating point form to x in integer form.
func (gas GeneticAlgorithmSolver) XRealToXInt(xreal float64) int {
	return int(math.Ceil((xreal - gas.a) * (math.Pow(2, float64(gas.l)) - 1) / (gas.b - gas.a)))
}

// L returns the minimal bit size for representation of all of the population.
func (gas GeneticAlgorithmSolver) L() int {
	return gas.l
}
