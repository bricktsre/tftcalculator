package main

import (
	"net/http"
	"strconv"

	"gonum.org/v1/gonum/mat"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
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
	_, cols := t.Dims()
	start := mat.NewDense(cols, 1, nil)
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

func generateLineItems() []opts.LineData {
	iterations := 101
	prob := make([]float64, iterations+1)
	t := generateTransitionMatrix(7, 3, 3, 6, 20)

	cdf := make([]opts.LineData, 0)
	for i := 0; i < iterations; i++ {
		prob[i] = simulateRolls(t, i)
		cdf = append(cdf, opts.LineData{Value: prob[i]})
	}

	pdf := make([]opts.LineData, 0)
	for i := 0; i < iterations-1; i++ {
		probDiff := prob[i+1] - prob[i]
		pdf = append(pdf, opts.LineData{Value: probDiff})
	}

	return pdf
}

func httpserver(w http.ResponseWriter, _ *http.Request) {
	numStrings := make([]string, 101)

	// Populate the slice with the string representations of numbers 0 to 100
	for i := 0; i <= 100; i++ {
		numStrings[i] = strconv.Itoa(i)
	}

	// create a new line instance
	line := charts.NewLine()
	// set some global options like Title/Legend/ToolTip or anything else
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Line example in Westeros theme",
			Subtitle: "Line chart rendered by the http server this time",
		}))

	// Put data into instance
	line.SetXAxis(numStrings).
		AddSeries("CDF", generateLineItems()).
		SetSeriesOptions(
			charts.WithLineChartOpts(
				opts.LineChart{
					Step: opts.Bool(true),
				}),
			charts.WithLabelOpts(
				opts.Label{
					Show: opts.Bool(false),
				}),
			charts.WithAreaStyleOpts(
				opts.AreaStyle{
					Opacity: 0.2,
				}),
		)
	line.Render(w)
}

// func main() {
// 	http.HandleFunc("/", httpserver)
// 	http.ListenAndServe(":8081", nil)
// }
