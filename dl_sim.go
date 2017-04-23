package main

import (
	"fmt"
	//"sync"
	"os"
	//"io"
	"bufio"
	"strconv"
	"math/rand"
	"time"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

const DL_LEN = 512
var DL_QTY int = 8
const AV_TRAIN_PER_BIN = 128
const AV_TEST_PER_BIN = 1024
const TRAIN_QTY = DL_LEN * AV_TRAIN_PER_BIN
const TEST_QTY = DL_LEN * AV_TEST_PER_BIN
const AV_BIN_SZ_PS = 15
const BIN_SZ_STD_PS = 7
const MAX_TIME_PS = 7000

func draw_graph(x, y []float64, filename, xlabel, ylabel string) {
	graph := chart.Chart{
		XAxis: chart.XAxis {
			Name: xlabel,
			NameStyle: chart.StyleShow(),
			Style: chart.StyleShow(),
		},
		YAxis: chart.YAxis {
			Name: ylabel,
			NameStyle: chart.StyleShow(),
			Style: chart.StyleShow(),
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style {
					Show: true,
					StrokeColor: chart.GetDefaultColor(0).WithAlpha(64),
					FillColor: chart.GetDefaultColor(0).WithAlpha(64),
				},
				XValues: x,
				YValues: y,
			},
		},
		Height: 1000,
		Width: 2000,
	}
	file, err := os.Create(filename)
	fmt.Println(err)
	writer := bufio.NewWriter(file)
	graph.Render(chart.PNG, writer)
	file.Close()
}	

func draw_multi_scatter(x, y [][]float64, cols []drawing.Color, filename, xlabel, ylabel string){
	series := make([]chart.Series, len(x))
	for i,_ := range series {
		series[i] = chart.ContinuousSeries{
			Style: chart.Style {
				Show: true,
				StrokeWidth: chart.Disabled,
				DotWidth: 0.01,
				DotColor: cols[i],
			},
			XValues: x[i],
			YValues: y[i],
		}
	}
	graph := chart.Chart{
		XAxis: chart.XAxis {
			Name: xlabel,
			NameStyle: chart.StyleShow(),
			Style: chart.StyleShow(),
		},
		YAxis: chart.YAxis {
			Name: ylabel,
			NameStyle: chart.StyleShow(),
			Style: chart.StyleShow(),
		},
		Series: series,
		Height: 1000,
		Width: 2000,
	}
	file, err := os.Create(filename)
	fmt.Println(err)
	writer := bufio.NewWriter(file)
	graph.Render(chart.PNG, writer)
	file.Close()
}

func main() {
	//Parameters
	rand.Seed(time.Now().UnixNano())
	DL_QTY,_ = strconv.Atoi(os.Args[1])
	fmt.Printf("DL Length: %v\n", DL_LEN)
	fmt.Printf("Number of Delay Lines: %v\n", DL_QTY)
	fmt.Printf("Average Training Hits per Bin: %v\n", AV_TRAIN_PER_BIN)
	fmt.Printf("Average Test Hits per Bin: %v\n", AV_TEST_PER_BIN)
	fmt.Printf("Total Training Hits: %v\n", TRAIN_QTY)
	fmt.Printf("Total Test Hits: %v\n", TEST_QTY)
	fmt.Printf("Average Bin Size in Picoseconds: %v\n", AV_BIN_SZ_PS)
	fmt.Printf("Bin Size Standard Deviation in Picoseconds: %v\n", BIN_SZ_STD_PS)
	fmt.Printf("Maximum Start-to-Stop Time in Picoseconds: %v\n", MAX_TIME_PS)
	//Set-up delay lines
	taus := make([][]float64, DL_QTY)
	acc_taus := make([][]float64, DL_QTY)
	est_taus := make([][]float64, DL_QTY)
	est_acc_taus := make([][]float64, DL_QTY)
	for i := 0; i < DL_QTY; i++ {
		taus[i] = make([]float64, DL_LEN)
		acc_taus[i] = make([]float64, DL_LEN+1)
		est_taus[i] = make([]float64, DL_LEN)
		est_acc_taus[i] = make([]float64, DL_LEN+1)
		acc_taus[i][0] = 0;
		for j := 0; j < DL_LEN; j++ {
			taus[i][j] = AV_BIN_SZ_PS + (BIN_SZ_STD_PS * rand.NormFloat64())
			acc_taus[i][j+1] = acc_taus[i][j] + taus[i][j]
		}
		
		x := make([]float64, DL_LEN)
		y := taus[i]
		for j,_ := range x {
			x[j] = float64 (j)
			//y[j] = float64(taus[i][j])
		}
		draw_graph(x, y, fmt.Sprintf("tau_%03v.png", i), "Bin_Number", "Width (ps)")
		
		x = make([]float64, DL_LEN+1)
		y = acc_taus[i]
		for j,_ := range x {
			x[j] = float64 (j)
			//y[j] = float64(acc_taus[i][j])
		}
		draw_graph(x, y, fmt.Sprintf("tau_acc_%03v.png", i), "Bin Number", "Relative Offset (ps)")
		
	}
	
	//Training Process
	for i := 0; i < DL_QTY; i++ {
		accs := make([]int, DL_LEN)
		for j := 0; j < DL_LEN; j++ {
			accs[j] = 0
		}
		for j := 0; j < TRAIN_QTY; j++ {
			time_elapsed := MAX_TIME_PS * rand.Float64()
			for k := DL_LEN-1; acc_taus[i][k+1] > time_elapsed; k-- {
				accs[k]++;
			}
		} 
		for j := 1; j < DL_LEN; j++ {
			if accs[j-1] >= TRAIN_QTY {
				accs[j] = accs[j-1] + ((AV_BIN_SZ_PS * TRAIN_QTY)) / MAX_TIME_PS
			}
		}
		est_acc_taus[i][0] = 0
		for j := 0; j < DL_LEN; j++ {
			est_acc_taus[i][j+1] = float64(accs[j] * MAX_TIME_PS) / float64(TRAIN_QTY)
			est_taus[i][j] = est_acc_taus[i][j+1] - est_acc_taus[i][j]
		}
		
		x := make([]float64, DL_LEN)
		y := est_taus[i]
		for j,_ := range x {
			x[j] = float64(j)
		}
		draw_graph(x, y, fmt.Sprintf("est_tau_%03v.png", i), "Bin Number", "Estimated Width (ps)")
		y = make([]float64, DL_LEN)
		for j,_ := range x {
			y[j] = est_taus[i][j] - taus[i][j]
		}
		draw_graph(x, y, fmt.Sprintf("calib_err_%03v.png", i), "Bin Number", "Calibration Error (ps)")
		
		x = make([]float64, DL_LEN+1)
		y = est_acc_taus[i]
		for j,_ := range x {
			x[j] = float64(j)
		}
		draw_graph(x, y, fmt.Sprintf("est_acc_tau_%03v.png", i), "Bin Number", "Estimated Relative Offset (ps)")
		y = make([]float64, DL_LEN+1)
		for j,_ := range x {
			y[j] = est_acc_taus[i][j] - acc_taus[i][j]
		}
		draw_graph(x, y, fmt.Sprintf("sum_err_%03v.png", i), "Bin Number", "INL (ps)")
	}
	
	//Testing
	x := make([]float64, TEST_QTY)
	xs := make([][]float64, 3)
	y_av := make([]float64, TEST_QTY)
	y_wav := make([]float64, TEST_QTY)
	y_mm := make([]float64, TEST_QTY)
	ys := make([][]float64, 3)
	cols := make([]drawing.Color, 3)
	for i := 0; i < TEST_QTY; i++ {
		time_elapsed := MAX_TIME_PS * rand.Float64()
		positions := make([]int, DL_QTY)
		for j := 0; j < DL_QTY; j++ {
			for k := 0; k < DL_LEN; k++ {
				if acc_taus[j][k+1] > time_elapsed {
					positions[j] = k
					break
				}
			}
		}
		var av, min, max, mm, wt, wav float64 = 0, 0, MAX_TIME_PS, 0, 0, 0
		for j := 0; j < DL_QTY; j++ {
			av += (est_acc_taus[j][positions[j]] + est_acc_taus[j][positions[j]+1]) / 2
			if(min < est_acc_taus[j][positions[j]]) {
				min = est_acc_taus[j][positions[j]]
			}
			if(max > est_acc_taus[j][positions[j]+1]) {
				max = est_acc_taus[j][positions[j]+1]
			}
			wt += 1/est_taus[j][positions[j]]
			wav += ((est_acc_taus[j][positions[j]] + est_acc_taus[j][positions[j]+1]) / 2) / est_taus[j][positions[j]]
		}
		av /= float64(DL_QTY)
		wav /= wt
		mm = (min + max) / 2
		av_err := av - time_elapsed
		wav_err := wav - time_elapsed
		mm_err := mm - time_elapsed
		x[i] = time_elapsed
		y_av[i] = av_err
		y_wav[i] = wav_err
		y_mm[i] = mm_err
	}
	for i,_ := range xs {
		xs[i] = x
	}
	ys[0] = y_av
	ys[1] = y_wav
	ys[2] = y_mm
	cols[0] = chart.GetDefaultColor(0)
	cols[1] = chart.GetDefaultColor(1)
	cols[2] = chart.GetDefaultColor(2)
	draw_multi_scatter(xs, ys, cols, "train_errs.png", "Time Elapsed (ps)", "Error (ps)")
}

