package evolalg

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestGenAlgorithmConstructor(t *testing.T) {
	gas, err := NewGeneticAlgorithmSolver(-2, 3, 3, func(x float64) float64 {
		return math.Mod(x, 1*(math.Cos(20*math.Pi*x)-math.Sin(x)))
	})
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	if gas.popSize != 5001 {
		t.Log(fmt.Sprintf("incorrect population initialized - gas.pop = %d\texpected = 5001", gas.popSize))
		t.Fail()
	} else if gas.l != 13 {
		t.Log(fmt.Sprintf("incorrect l or population array size\nl=%d\tlen(gas.popArr)=%d", gas.l, len(gas.popArr)))
		t.Fail()
	}

	gas, err = NewGeneticAlgorithmSolver(-1.322, 3.219, 2, func(x float64) float64 {
		return math.Mod(x, 1*(math.Cos(20*math.Pi*x)-math.Sin(x)))
	})
	if err == nil {
		t.Log("provided precision was lower than sufficient yet it passed with no error")
		t.Fail()
	}
}

func TestConversions(t *testing.T) {
	// Create the generic algorithm solver
	gas, err := NewGeneticAlgorithmSolver(-2, 3, 3, func(x float64) float64 {
		return math.Mod(x, 1*(math.Cos(20*math.Pi*x)-math.Sin(x)))
	})
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	// bin -> int
	arr := []byte{0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 1}
	resInt := gas.XBinToXInt(arr)
	if resInt != 69 {
		t.Log("binary to integer conversion failed")
		t.Log(fmt.Sprintf("resInt=%d", resInt))
		t.Fail()
	}

	// int -> bin
	resBin := gas.XIntToXBin(uint32(resInt))
	if !bytes.Equal(arr, resBin) {
		t.Log("integer to binary conversion failed")
		t.Log(fmt.Sprintf("resBin=%v", resBin))
		t.Fail()
	}

	// int -> real
	resFloat := gas.XIntToXReal(1255)
	if resFloat != -1.2339152728604565 {
		t.Log("integer to real conversion failed")
		t.Log(fmt.Sprintf("resFloat=%f", resFloat))
		t.Fail()
	}

	// real -> int
	resInt = gas.XRealToXInt(-1.234)
	if resInt != 1255 {
		t.Log("real to integer conversion failed")
		t.Log(fmt.Sprintf("resInt=%d", resInt))
		t.Fail()
	}

	// bin -> real
	resFloat = gas.XIntToXReal(int(gas.XBinToXInt([]byte{0, 0, 1, 0, 0, 1, 1, 1, 0, 0, 1, 1, 1})))
	if resFloat != -1.2339152728604565 {
		t.Log("binary to real conversion failed")
		t.Log(fmt.Sprintf("resFloat=%f", resFloat))
		t.Fail()
	}

	// real -> bin
	resBin = gas.XIntToXBin(uint32(gas.XRealToXInt(-1.234)))
	if !bytes.Equal([]byte{0, 0, 1, 0, 0, 1, 1, 1, 0, 0, 1, 1, 1}, resBin) {
		t.Log("real to binary conversion failed")
		t.Log(fmt.Sprintf("resBin=%v", resBin))
		t.Fail()
	}
}

func TestCrossoverAndMutation(t *testing.T) {
	N := 10
	a, b := -4.0, 12.0
	d := 3
	gas, err := NewGeneticAlgorithmSolver(a, b, byte(d), func(x float64) float64 {
		return math.Mod(x, 1*(math.Cos(20*math.Pi*x)-math.Sin(x)))
	})
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	rands := make([]float64, N)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < N; i++ {
		rands[i] = a + rand.Float64()*(b-a)
		rands[i] = math.Round(rands[i]*math.Pow10(int(d))) / math.Pow10(int(d)) // TODO Sprawdzić czy to potrzebne
	}
	err = gas.Selection(rands...)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	t.Log("Before crossover")
	for i := 0; i < len(gas.popArr); i++ {
		t.Log(fmt.Sprintf("%v", gas.popArr[i]))
	}
	_, _, _, err = gas.Crossover(0.75)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	fmt.Println()
	t.Log("After crossover")
	for i := 0; i < len(gas.popArr); i++ {
		t.Log(fmt.Sprintf("%v", gas.popArr[i]))
	}

	mut, err := gas.Mutate(0.005)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	fmt.Println()
	t.Log("After mutations")
	for i := 0; i < len(gas.popArr); i++ {
		t.Log(fmt.Sprintf("%v", gas.popArr[i]))
	}
	t.Log("Mutation points")
	for i := 0; i < len(gas.popArr); i++ {
		t.Log(fmt.Sprintf("%v", mut[i]))
	}
}
