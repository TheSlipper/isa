package evolalg

import (
	"errors"
	"math"
)

// GeneticAlgorithmSolver is a struct that contains all of the data related to a generic algorithm instance and shares
// a set of functions for solving this certain genetic algorithm.
type GeneticAlgorithmSolver struct {
	a       float64                 // lower bound of the set (inclusive)
	b       float64                 // higher bound of the set (inclusive)
	d       byte                    // accuracy (e.g. accuracy 3 means 10^3 for each decimal therefore for 3 it would generate a 3001 population)
	l       int                     // minimal bit size for representation of all of the population
	fmin    float64                 // lowest value of the gFunc in the <a, b> set
	popSize uint                    // actual population size (higher bound of <0, pop> set)
	popArr  []byte                  // population array in little endian
	gFunc   func(x float64) float64 // function responsible for grading the received solution

	// fitCache  map[float64]float64 // fit cache. Holds the values of calculated fits until cleared.
	// probCache map[float64]float64 // probability cache. Holds the values of calculated probabilities until cleared.
	fitSum      float64   // sum of all the cached fits. Stored in struct for optimization purposes.
	gradeCache  []float64 // grade cache. Holds the values of calculated grades.
	fitCache    []float64 // fit cache. Holds the values of calculated fits until cleared.
	probCache   []float64 // probability cache. Holds the values of calculated probabilities until cleared.
	probHBCache []float64 // Holds the values of probability's higher bound in cumulative distribution bound.
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
func NewGeneticAlgorithmSolver(a float64, b float64, d byte, gFunc func(x float64) float64) (ga GeneticAlgorithmSolver, err error) {
	if a > b {
		err = errors.New("provided lower bound greater than higher bound")
		return
	} else if d < 1 {
		err = errors.New("provided precision is equal to or lower than zero")
		return
	}

	// populate the members of the struct
	ga.a = a
	ga.b = b
	ga.d = d
	ga.gFunc = gFunc

	// calculate fmin
	ga.fmin = math.MaxFloat64
	for i := ga.a; i < ga.b; i += math.Pow(1, -float64(ga.d)) { // TODO Check if this returns -0.001 for 10^-3
		val := ga.gFunc(i)
		if ga.fmin > val {
			ga.fmin = val
		}
	}

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

// Selection starts the process of selection for the passed values.
func (gas *GeneticAlgorithmSolver) Selection(vals ...float64) error {
	// Check if the given values are valid
	for _, val := range vals {
		if val < gas.a || val > gas.b {
			return errors.New("at least one passed value is not contained in <a,b> set")
		}
	}

	// Calculate the grades and fits
	N := len(vals)
	grades := make([]float64, len(vals))
	fits := make([]float64, len(vals))
	for i := 0; i < N; i++ {
		grades[i] = gas.Grade(vals[i])
		fits[i] = gas.fit(vals[i])
	}
	gas.gradeCache = grades
	gas.fitCache = fits

	// Calculate the probability
	prob := make([]float64, len(vals))
	probHBounds := make([]float64, len(vals))
	for i := 0; i < N; i++ {
		prob[i] = gas.probability(i)
	}
	gas.probCache = prob

	// Calculate the cumulative distribution
	for i := 0; i < N; i++ {
		probHBounds[i] = gas.cdfUpperBound(i)
	}
	gas.probHBCache = probHBounds

	return nil
}

// Grade calculates the grade of the x argument at the point x.
func (gas GeneticAlgorithmSolver) Grade(x float64) float64 {
	return gas.gFunc(x)
}

// fit returns x's fit (implemented for searching MAX).
func (gas *GeneticAlgorithmSolver) fit(x float64) float64 {
	fit := gas.Grade(x) - gas.fmin + math.Pow(1, -float64(gas.d))
	gas.fitSum += fit
	return fit
}

// Probability calculates the probability of the i-th fit. Should be ran after all the Grade and Fit calls.
func (gas *GeneticAlgorithmSolver) probability(i int) float64 {
	return gas.fitCache[i] / gas.fitSum
}

// cdfUpperBound returns the upper bound of a given i in the cumulative distribution function of this algorithm.
func (gas *GeneticAlgorithmSolver) cdfUpperBound(i int) float64 {
	if i == 0 {
		return gas.probCache[i]
	} else {
		return gas.cdfUpperBound(i-1) + gas.probCache[i]
	}
}

// Cache returns all of the cached results from the previous selection.
func (gas *GeneticAlgorithmSolver) Cache() (gradeCache []float64, fitCache []float64, probCache []float64, probHBCache []float64) {
	gradeCache = gas.gradeCache
	fitCache = gas.fitCache
	probCache = gas.probCache
	probHBCache = gas.probHBCache

	return
}
