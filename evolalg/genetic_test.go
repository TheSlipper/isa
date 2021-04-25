package evolalg

import (
	"fmt"
	"io/ioutil"
	"math"
	"strings"
	"testing"
	"time"
)

// func TestGenAlgorithmConstructor(t *testing.T) {
// 	gas, err := NewGeneticAlgorithmSolver(-2, 3, 3, func(x float64) float64 {
// 		return math.Mod(x, 1*(math.Cos(20*math.Pi*x)-math.Sin(x)))
// 	})
// 	if err != nil {
// 		t.Log(err.Error())
// 		t.Fail()
// 	}
// 	if gas.popSize != 5001 {
// 		t.Log(fmt.Sprintf("incorrect population initialized - gas.pop = %d\texpected = 5001", gas.popSize))
// 		t.Fail()
// 	} else if gas.l != 13 {
// 		t.Log(fmt.Sprintf("incorrect l or population array size\nl=%d\tlen(gas.popArr)=%d", gas.l, len(gas.popArr)))
// 		t.Fail()
// 	}

// 	gas, err = NewGeneticAlgorithmSolver(-1.322, 3.219, 2, func(x float64) float64 {
// 		return math.Mod(x, 1*(math.Cos(20*math.Pi*x)-math.Sin(x)))
// 	})
// 	if err == nil {
// 		t.Log("provided precision was lower than sufficient yet it passed with no error")
// 		t.Fail()
// 	}
// }

// func TestConversions(t *testing.T) {
// 	// Create the generic algorithm solver
// 	gas, err := NewGeneticAlgorithmSolver(-2, 3, 3, func(x float64) float64 {
// 		return math.Mod(x, 1*(math.Cos(20*math.Pi*x)-math.Sin(x)))
// 	})
// 	if err != nil {
// 		t.Log(err.Error())
// 		t.Fail()
// 	}

// 	// bin -> int
// 	arr := []byte{0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 1}
// 	resInt := gas.XBinToXInt(arr)
// 	if resInt != 69 {
// 		t.Log("binary to integer conversion failed")
// 		t.Log(fmt.Sprintf("resInt=%d", resInt))
// 		t.Fail()
// 	}

// 	// int -> bin
// 	resBin := gas.XIntToXBin(uint32(resInt))
// 	if !bytes.Equal(arr, resBin) {
// 		t.Log("integer to binary conversion failed")
// 		t.Log(fmt.Sprintf("resBin=%v", resBin))
// 		t.Fail()
// 	}

// 	// int -> real
// 	resFloat := gas.XIntToXReal(1255)
// 	if resFloat != -1.2339152728604565 {
// 		t.Log("integer to real conversion failed")
// 		t.Log(fmt.Sprintf("resFloat=%f", resFloat))
// 		t.Fail()
// 	}

// 	// real -> int
// 	resInt = gas.XRealToXInt(-1.234)
// 	if resInt != 1255 {
// 		t.Log("real to integer conversion failed")
// 		t.Log(fmt.Sprintf("resInt=%d", resInt))
// 		t.Fail()
// 	}

// 	// bin -> real
// 	resFloat = gas.XIntToXReal(int(gas.XBinToXInt([]byte{0, 0, 1, 0, 0, 1, 1, 1, 0, 0, 1, 1, 1})))
// 	if resFloat != -1.2339152728604565 {
// 		t.Log("binary to real conversion failed")
// 		t.Log(fmt.Sprintf("resFloat=%f", resFloat))
// 		t.Fail()
// 	}

// 	// real -> bin
// 	resBin = gas.XIntToXBin(uint32(gas.XRealToXInt(-1.234)))
// 	if !bytes.Equal([]byte{0, 0, 1, 0, 0, 1, 1, 1, 0, 0, 1, 1, 1}, resBin) {
// 		t.Log("real to binary conversion failed")
// 		t.Log(fmt.Sprintf("resBin=%v", resBin))
// 		t.Fail()
// 	}
// }

// func TestCrossoverAndMutation(t *testing.T) {
// 	N := 10
// 	a, b := -4.0, 12.0
// 	d := 3
// 	gas, err := NewGeneticAlgorithmSolver(a, b, byte(d), func(x float64) float64 {
// 		return math.Mod(x, 1*(math.Cos(20*math.Pi*x)-math.Sin(x)))
// 	})
// 	if err != nil {
// 		t.Log(err.Error())
// 		t.Fail()
// 	}

// 	rands := make([]float64, N)
// 	rand.Seed(time.Now().UnixNano())
// 	for i := 0; i < N; i++ {
// 		rands[i] = a + rand.Float64()*(b-a)
// 		rands[i] = math.Round(rands[i]*math.Pow10(int(d))) / math.Pow10(int(d)) // TODO Sprawdzić czy to potrzebne
// 	}
// 	err = gas.Selection(rands...)
// 	if err != nil {
// 		t.Log(err.Error())
// 		t.Fail()
// 		return
// 	}

// 	t.Log("Before crossover")
// 	for i := 0; i < len(gas.popArr); i++ {
// 		t.Log(fmt.Sprintf("%v", gas.popArr[i]))
// 	}
// 	_, _, _, err = gas.Crossover(0.75)
// 	if err != nil {
// 		t.Log(err.Error())
// 		t.Fail()
// 		return
// 	}
// 	fmt.Println()
// 	t.Log("After crossover")
// 	for i := 0; i < len(gas.popArr); i++ {
// 		t.Log(fmt.Sprintf("%v", gas.popArr[i]))
// 	}

// 	mut, err := gas.Mutate(0.005)
// 	if err != nil {
// 		t.Log(err.Error())
// 		t.Fail()
// 		return
// 	}
// 	fmt.Println()
// 	t.Log("After mutations")
// 	for i := 0; i < len(gas.popArr); i++ {
// 		t.Log(fmt.Sprintf("%v", gas.popArr[i]))
// 	}
// 	t.Log("Mutation points")
// 	for i := 0; i < len(gas.popArr); i++ {
// 		t.Log(fmt.Sprintf("%v", mut[i]))
// 	}
// }

func run(a, b, cp, mp float64, d byte, N, epochs int, bench *testing.B) (fmin, favg, fmax float64) {
	// Generate the values if all the necessary data was given
	var hist []EpochData

	// solver for this assignment with a grading function described by this formula:
	// F(x)= x MOD1 *(COS(20*π *x)–SIN(x))
	// d = 0,001 -> 3
	gas, err := NewGeneticAlgorithmSolver(a, b, d, func(x float64) float64 {
		return math.Mod(x, 1*(math.Cos(20*math.Pi*x)-math.Sin(x)))
	})
	if err != nil {
		bench.Log(err)
		bench.Fail()
		return
	}

	hist, err = gas.Solve(N, epochs, cp, mp)
	if err != nil {
		bench.Log(err)
		bench.Fail()
		return
	}

	// Create the fmax, favg, fmin graph
	// fmax, favg, fmin := make([]float64, len(hist)), make([]float64, len(hist)), make([]float64, len(hist))
	// epochsArr := make([]float64, len(hist))
	// i := 0
	// for ; i < len(hist); i++ {
	// 	fmax[i] = hist[i].FMax
	// 	favg[i] = hist[i].FAVG
	// 	fmin[i] = hist[i].FMin
	// 	epochsArr[i] = float64(i)
	// }
	fmin = hist[len(hist)-1].FMin
	favg = hist[len(hist)-1].FAVG
	fmax = hist[len(hist)-1].FMax
	return

	// sc <- fmt.Sprintf("%d, %d, %f, %f, %f, %f, %f", N, epochs, cp, mp, fmax, favg, fmin)
}

func runLoop(a, b, cp, mp float64, d byte, N, epochs, iters int, bench *testing.B, sc chan string) {
	fmax, favg, fmin := 0.0, 0.0, 0.0
	for i := 0; i < iters; i++ {
		fminI, favgI, fmaxI := run(a, b, cp, mp, d, N, epochs, bench)
		fmin += fminI
		favg += favgI
		fmax += fmaxI
	}
	fmax, favg, fmin = fmax/float64(iters), favg/float64(iters), fmin/float64(iters)

	sc <- fmt.Sprintf("%d, %d, %f, %f, %f, %f, %f\n", N, epochs, cp, mp, fmax, favg, fmin)
}

func BenchmarkLabs(bench *testing.B) {
	a, b := float64(-4), float64(12)
	d := byte(3)

	// N set
	var Ns []int
	for i := 30; i <= 80; i += 5 {
		Ns = append(Ns, i)
	}
	NsLen := len(Ns)

	// cp set
	var cps []float64
	for i := 0.50; i <= 0.9; i += 0.05 {
		cps = append(cps, i)
	}
	cpsLen := len(cps)

	// mp set
	var mps []float64
	mps = append(mps, 0.0001)
	for i := 0.0005; i <= 0.01; i += 0.0005 {
		mps = append(mps, i)
	}
	mpsLen := len(mps)

	// epoch set
	var epochSet []int
	for i := 50; i <= 150; i += 10 {
		epochSet = append(epochSet, i)
	}
	epochsLen := len(epochSet)

	bench.StartTimer()
	workerRoutineCnt := 16
	strChan := make(chan string, workerRoutineCnt)
	// var wg sync.WaitGroup
	var sb strings.Builder
	sb.WriteString("N, epochs, cp, mp, fmax, favg, fmin\n")
	launched := 0
	for i := 0; i < mpsLen; i++ {
		for j := 0; j < epochsLen; j++ {
			for k := 0; k < NsLen; k++ {
				for l := 0; l < cpsLen; l++ {
					for len(strChan) > 0 || launched == workerRoutineCnt {
						select {
						case foo, ok := <-strChan:
							if ok {
								sb.WriteString(foo)
								launched--
							} else {
								panic("string channel closed")
							}
						default:
							time.Sleep(time.Microsecond * 20)
						}
					}

					// go run(a, b, cps[l], mps[i], d, Ns[k], epochSet[j], bench, strChan)
					go runLoop(a, b, cps[l], mps[i], d, Ns[k], epochSet[j], 2, bench, strChan)
					launched++
				}
			}
		}
	}
	for len(strChan) > 0 {
		select {
		case foo, ok := <-strChan:
			if ok {
				sb.WriteString(foo)
			} else {
				panic("string channel closed")
			}
		default:
			time.Sleep(time.Microsecond * 20)
		}
	}
	bench.StopTimer()

	ioutil.WriteFile("../static/stats.csv", []byte(sb.String()), 0666)
}
