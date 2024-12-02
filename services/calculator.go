package services

import (
	"context"
	"errors"
	"fmt"
	"math"

	"golang.org/x/exp/slog"
	"gonum.org/v1/gonum/mat"
)

var uniques = []int{14, 13, 13, 12, 8}
var copies = []int{30, 25, 18, 10, 9}
var weights = [][]float64{{100, 0, 0, 0, 0}, {100, 0, 0, 0, 0}, {75, 25, 0, 0, 0},
	{55, 30, 15, 0, 0}, {45, 33, 20, 2, 0}, {30, 40, 25, 5, 0},
	{19, 30, 40, 10, 1}, {18, 25, 32, 22, 3}, {15, 20, 25, 30, 10},
	{5, 10, 20, 40, 25}}

func generateTransitionMatrix(level, tier, goal, c_owned, t_owned int) *mat.Dense {
	t := mat.NewDense(goal+1, goal+1, nil)
	base_prob := weights[level-1][tier-1] / 100.0
	for i := 0; i < goal; i++ {
		prob := 0.0
		if copies[tier-1]-c_owned > i {
			prob = base_prob * float64(copies[tier-1]-c_owned-i) / float64(copies[tier-1]*uniques[tier-1]-t_owned-i)
		}

		t.Set(i, i, 1-prob)
		t.Set(i+1, i, prob)
	}
	t.Set(goal, goal, 1)

	return t
}

func iterateMarkovChain(t *mat.Dense, iterations int) mat.Dense {
	rows, _ := t.Dims()
	start := mat.NewDense(rows, 1, nil)
	start.Set(0, 0, 1)

	var power mat.Dense
	power.Pow(t, iterations)

	var result mat.Dense
	result.Mul(&power, start)

	return result
}

func simulateRolls(t *mat.Dense, rolls int) float64 {
	final := iterateMarkovChain(t, rolls*5)

	rows, _ := final.Dims()
	return final.At(rows-1, 0)
}

type Counts struct {
	ExpectedRolls int
}

func NewCalculator(log *slog.Logger) Calculator {
	return Calculator{
		Log: log,
	}
}

type Calculator struct {
	Log *slog.Logger
}

func expectedRolls(matT *mat.Dense) int {
	matQ := matT.Slice(0, 3, 0, 3).(*mat.Dense)
	fq := mat.Formatted(matQ, mat.Prefix("    "), mat.Squeeze())
	fmt.Printf("q = %v\n\n", fq)
	rows, cols := matQ.Dims()

	identity := mat.NewDense(rows, cols, nil)
	mat1 := mat.NewDense(1, cols, nil)
	for i := 0; i < rows; i++ {
		identity.Set(i, i, 1)
		mat1.Set(0, i, 1)
	}

	var sub mat.Dense
	sub.Sub(identity, matQ)

	fs := mat.Formatted(&sub, mat.Prefix("    "), mat.Squeeze())
	fmt.Printf("sub = %v\n\n", fs)

	var inverse mat.Dense
	inverse.Inverse(&sub)

	fi := mat.Formatted(&inverse, mat.Prefix("    "), mat.Squeeze())
	fmt.Printf("inverse = %v\n\n", fi)

	var result mat.Dense
	result.Mul(mat1, &inverse)

	fr := mat.Formatted(&result, mat.Prefix("    "), mat.Squeeze())
	fmt.Printf("result = %v\n\n", fr)

	return int(math.Ceil(result.At(0, 0) / 5))
}

func (cs Calculator) Calculate(ctx context.Context, sessionID string, level, tier, goal, copiesOwned, tierOwned int) (counts Counts, err error) {
	errs := make([]error, 2)

	iterations := 100
	prob := make([]float64, iterations+1)
	matT := generateTransitionMatrix(level, tier, goal, copiesOwned, tierOwned)

	for i := 0; i < iterations; i++ {
		prob[i] = simulateRolls(matT, i)
	}

	currentProb := 0.0
	for i := 0; i < iterations; i++ {
		currentProb = currentProb + prob[i]
		if currentProb >= 50.0 {
			fmt.Printf("Took %v rolls to hit 50 percent probability\n", i)
			break
		}
	}

	counts.ExpectedRolls = expectedRolls(matT)

	return counts, errors.Join(errs...)
}

func (cs Calculator) Get(ctx context.Context, sessionID string) (counts Counts, err error) {
	counts.ExpectedRolls = 0
	return
}
