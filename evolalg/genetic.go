package evolalg

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// EpochData contains data on the state of a generic algorithm's solution after an iteration of
// calculations.
type EpochData struct {
	PopulationBytes [][]byte  `json:"populationBytes"`
	PopulationF64   []float64 `json:"populationF64"`
	Fits            []float64 `json:"fits"`
	Grades          []float64 `json:"grades"`
	Elite           float64   `json:"elite"`
	EliteFit        float64   `json:"eliteFit"`
	FMin            float64   `json:"fMin"`
	FAVG            float64   `json:"fAVG"`
	FMax            float64   `json:"fMax"`
}

// GeneticAlgorithmSolver is a struct that contains all of the data related to a generic algorithm instance and shares
// a set of functions for solving this certain genetic algorithm.
type GeneticAlgorithmSolver struct {
	a        float64                 // lower bound of the set (inclusive)
	b        float64                 // higher bound of the set (inclusive)
	d        byte                    // accuracy (e.g. accuracy 3 means 10^3 for each decimal therefore for 3 it would generate a 3001 population)
	l        int                     // minimal bit size for representation of all of the population
	elite    float64                 // the value used for the current best solution.
	eliteFit float64                 // the fit of the currently best solution.
	fmin     float64                 // lowest value of the gFunc in the <a, b> set
	popSize  uint                    // actual population size (higher bound of <0, pop> set)
	popArr   [][]byte                // population array in little endian
	gFunc    func(x float64) float64 // function responsible for grading the received solution

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
	// ga.popArr = make([]byte, ga.l)

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
	val := gas.a + ((gas.b - gas.a) * float64(xint) / (math.Pow(2, float64(gas.l)) - 1))
	val = math.Round(val*math.Pow10(int(gas.d))) / math.Pow10(int(gas.d))
	return val
}

// XRealToXInt converts x in floating point form to x in integer form.
func (gas GeneticAlgorithmSolver) XRealToXInt(xreal float64) int {
	return int(math.Ceil((xreal - gas.a) * (math.Pow(2, float64(gas.l)) - 1) / (gas.b - gas.a)))
}

// L returns the minimal bit size for representation of all of the population.
func (gas GeneticAlgorithmSolver) L() int {
	return gas.l
}

// Population returns the current population.
func (gas GeneticAlgorithmSolver) Population() [][]byte {
	pop := make([][]byte, len(gas.popArr))
	for i := 0; i < len(pop); i++ {
		dst := make([]byte, len(gas.popArr[i]))
		copy(dst, gas.popArr[i])
		pop[i] = dst
	}

	return pop
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

// Probability calculates the probability of the i-th fit. Should be ran after all the Grade and Fit
// calls.
func (gas *GeneticAlgorithmSolver) probability(i int) float64 {
	return gas.fitCache[i] / gas.fitSum
}

// cdfUpperBound returns the upper bound of a given i in the cumulative distribution function of
// this algorithm.
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

// Crossover runs an operation that groups random genomes in pairs (parents) and combines their
// genetic information to generate new offsprings. The probability of crossover is defined by the
// passed parameter. Panics if that parameter is not in these bounds: 0.5 < cp <= 1.
func (gas *GeneticAlgorithmSolver) Crossover(cp float64) (parents [][]byte, offsprings [][]byte, cutpoints []int, err error) {
	// 1:58:17
	if len(gas.popArr) == 0 || len(gas.probCache) == 0 {
		return nil, nil, nil, errors.New("invalid cache state")
	} else if 0.5 > cp || cp > 1 {
		return nil, nil, nil, errors.New("provided invalid crossover probability value")
	}

	cutpoints = make([]int, len(gas.popArr))

	// Pick parents in a random manner from the current population
	// rand.Seed(time.Now().Unix())
	for i := 0; i < len(gas.popArr); i++ {
		var parent []byte
		if gas.probHBCache[i] <= cp {
			parent = gas.popArr[i]
		} else {
			parent = nil
		}
		offsprings = append(offsprings, nil)
		parents = append(parents, parent)
	}

	// Perform the operation of crossover
	var parentA, parentB []byte
	for i := 0; i < len(parents); i++ {
		if parents[i] == nil {
			cutpoints[i] = -1
			continue
		}
		parentA = parents[i]
		offsprings[i] = make([]byte, gas.l)
		j := i + 1
		for ; j < len(parents); j++ {
			if parents[j] == nil {
				continue
			}
			parentB = parents[j]
			offsprings[j] = make([]byte, gas.l)
			break
		}

		// If no parrents left then the parrent is a bachelor and will be passed further
		if parentB == nil {
			offsprings[i] = parentA
			break
		}

		// Get random cut point and crossover them in place of i and j
		cut := rand.Intn(gas.l - 1)
		slice1, slice2 := parentA[cut:], parentB[cut:]
		for k := 0; k < cut; k++ {
			offsprings[i][k] = parentA[k]
			offsprings[j][k] = parentB[k]
		}
		l := 0
		for k := cut; k < gas.l; k++ {
			offsprings[i][k] = slice2[l]
			offsprings[j][k] = slice1[l]
			l++
		}

		cutpoints[i], cutpoints[j] = cut, cut
		parentA, parentB = nil, nil
		i = j
	}

	// Update the population
	for i := 0; i < len(gas.popArr); i++ {
		if offsprings[i] == nil {
			continue
		}
		gas.popArr[i] = offsprings[i]
	}

	return
}

// Mutate runs an operation that mutates random bits in the current population based on the passed
// mutation probability. Panics if probability is not in these bounds: 0 < mp <= 0.01.
func (gas *GeneticAlgorithmSolver) Mutate(mp float64) (mutations [][]int, err error) {
	if len(gas.popArr) == 0 {
		return nil, errors.New("invalid cache state")
	} else if mp <= 0 || mp > 0.01 {
		return nil, errors.New("provided invalid mutation probability value")
	}

	mutations = make([][]int, len(gas.popArr))

	// rand.Seed(time.Now().Unix())
	for i := 0; i < len(gas.popArr); i++ {
		var localMutations []int
		for j := 0; j < len(gas.popArr[i]); j++ {
			r := rand.Float64()
			if r <= mp {
				localMutations = append(localMutations, j)
				if gas.popArr[i][j] == 0 {
					gas.popArr[i][j] = 1
				} else {
					gas.popArr[i][j] = 0
				}
			}
		}
		if localMutations != nil {
			mutations[i] = localMutations
		}
	}

	return
}

// saveStateToHistory saves the current state of a genetic algorithm solver to an epoch data struct.
func (gas GeneticAlgorithmSolver) saveStateToHistory(N int, vals []float64, ed *EpochData) (err error) {
	ed.PopulationF64 = make([]float64, N)
	ed.Fits = make([]float64, N)
	ed.Grades = make([]float64, N)
	// ed.PopulationBytes = make([][]byte, N)

	copied := copy(ed.PopulationF64, vals)
	if copied != N {
		err = fmt.Errorf("insufficient amount of elements copied - %d instead of %d", copied, N)
		return
	}
	copied = copy(ed.Fits, gas.fitCache)
	if copied != N {
		err = fmt.Errorf("insufficient amount of elements copied - %d instead of %d", copied, N)
		return
	}
	copied = copy(ed.Grades, gas.gradeCache)
	if copied != N {
		err = fmt.Errorf("insufficient amount of elements copied - %d instead of %d", copied, N)
		return
	}

	ed.PopulationBytes = gas.Population()
	ed.Elite = gas.elite
	ed.EliteFit = gas.eliteFit

	// fmin, favg fmax - values of best, worst and max values of this epoch
	ed.FMin, ed.FAVG, ed.FMax = 100000000000, 0, -100000
	i := 0
	for ; i < len(gas.gradeCache); i++ {
		if ed.FMin > gas.gradeCache[i] {
			ed.FMin = gas.gradeCache[i]
		}
		if ed.FMax < gas.gradeCache[i] {
			ed.FMax = gas.gradeCache[i]
		}
		ed.FAVG += gas.gradeCache[i]
	}
	ed.FAVG = ed.FAVG / float64(i)

	return
}

// updateElite searches for a new elite and updates the solver data.
func (gas *GeneticAlgorithmSolver) updateElite(vals []float64) {
	for i := 0; i < len(gas.fitCache); i++ {
		if gas.eliteFit < gas.fitCache[i] {
			gas.eliteFit = gas.fitCache[i]
			gas.elite = vals[i]
		}
	}
}

// Solve runs the genetic algorithm solver for N random solutions, for a given amount of epochs, for
// a given crossing probability, for a given mutation probability and returns a history of the
// algorithm's execution.
func (gas *GeneticAlgorithmSolver) Solve(N, epochs int, cp, mp float64) (hist []EpochData, err error) {
	// Create a history
	hist = make([]EpochData, epochs+1)

	// Initialize the solver
	vals := make([]float64, N)
	gas.popArr = make([][]byte, N)
	gas.fitCache = make([]float64, N)
	gas.gradeCache = make([]float64, N)
	rand.Seed(time.Now().UnixNano())
	gas.eliteFit = -10000000
	gas.elite = gas.eliteFit

	for i := 0; i < N; i++ {
		// Create a population
		vals[i] = gas.a + rand.Float64()*(gas.b-gas.a)
		vals[i] = math.Round(vals[i]*math.Pow10(int(gas.d))) / math.Pow10(int(gas.d))

		// Calculate its grades
		gas.popArr[i] = gas.XIntToXBin(uint32(gas.XRealToXInt(vals[i])))
		gas.fitCache[i] = gas.fit(vals[i])
		gas.gradeCache[i] = gas.Grade(vals[i])

		// If fit's the best then pick it as an elite
		if gas.fitCache[i] > gas.eliteFit {
			gas.elite = vals[i]
			gas.eliteFit = gas.fitCache[i]
		}
	}

	// Save the current state to the history
	err = gas.saveStateToHistory(N, vals, &hist[0])
	if err != nil {
		return
	}

	// Run the selection
	err = gas.Selection(vals...)
	if err != nil {
		return
	}

	// Run as many times as it was specified (+1 because we saved in 0 the state before the algorithm)
	for i := 1; i < epochs+1; i++ {
		// Updates elites
		gas.updateElite(vals)

		// Run crossover
		_, _, _, err = gas.Crossover(cp)
		if err != nil {
			return
		}

		// Run mutation
		_, err = gas.Mutate(mp)
		if err != nil {
			return
		}

		// Update f64 population and calculate the new fits
		for i := 0; i < N; i++ {
			vals[i] = gas.XIntToXReal(gas.XBinToXInt(gas.popArr[i]))
		}

		// Run selection before the next run (and for saving the state of the epoch after it)
		err = gas.Selection(vals...)
		if err != nil {
			return
		}

		// Check if elite is still in - if not put it in a random place (unless the random place is better)
		eliteIn := false
		for i := 0; i < N; i++ {
			if gas.eliteFit == gas.fitCache[i] {
				eliteIn = true
				break
			}
		}
		if !eliteIn {
			i := rand.Intn(N)

			if gas.eliteFit < gas.fitCache[i] {
				gas.elite = vals[i]
				gas.eliteFit = gas.fitCache[i]
			} else {
				gas.popArr[i] = gas.XIntToXBin(uint32(gas.XRealToXInt(gas.elite)))
				vals[i] = gas.elite

				err = gas.Selection(vals...)
				if err != nil {
					return
				}
			}
		}

		// Save to history
		err = gas.saveStateToHistory(N, vals, &hist[i])
		if err != nil {
			return
		}
	}

	return
}
