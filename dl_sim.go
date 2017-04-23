package main

import (
	"fmt"
	//"sync"
	"os"
	//"io"
	"bufio"
	//"strconv"
	"math/rand"
	"time"
	"github.com/wcharczuk/go-chart"
)

const DL_LEN = 512
const DL_QTY = 1
const AV_TRAIN_PER_BIN = 128
const AV_TEST_PER_BIN = 1024
const TRAIN_QTY = DL_LEN * AV_TRAIN_PER_BIN
const TEST_QTY = DL_LEN * AV_TEST_PER_BIN
const AV_BIN_SZ_PS = 15
const BIN_SZ_STD_PS = 7
const MAX_TIME_PS = 7000

func draw_graph(len int, x, y []float64, filename, xlabel, ylabel string) {
	
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
	}
	file, err := os.Create(filename)
	fmt.Println(err)
	writer := bufio.NewWriter(file)
	defer file.Close()
	graph.Render(chart.PNG, writer)
	
}	

func main() {
	//Parameters
	rand.Seed(time.Now().UnixNano())
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
		draw_graph(DL_LEN, x, y, fmt.Sprintf("tau_%03v.png", i), "Bin_Number", "Width (ps)")
		
		x = make([]float64, DL_LEN+1)
		y = acc_taus[i]
		for j,_ := range x {
			x[j] = float64 (j)
			//y[j] = float64(acc_taus[i][j])
		}
		draw_graph(DL_LEN+1, x, y, fmt.Sprintf("tau_acc_%03v.png", i), "Bin Number", "Relative Offset (ps)")

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
		draw_graph(DL_LEN, x, y, fmt.Sprintf("est_tau_%03v.png", i), "Bin Number", "Estimated Width (ps)")
		y = make([]float64, DL_LEN)
		for j,_ := range x {
			y[j] = est_taus[i][j] - taus[i][j]
		}
		draw_graph(DL_LEN, x, y, fmt.Sprintf("calib_err_%03v.png", i), "Bin Number", "Calibration Error (ps)")

		x = make([]float64, DL_LEN+1)
		y = est_acc_taus[i]
		for j,_ := range x {
			x[j] = float64(j)
		}
		draw_graph(DL_LEN+1, x, y, fmt.Sprintf("est_acc_tau_%03v.png", i), "Bin Number", "Estimated Relative Offset (ps)")
		y = make([]float64, DL_LEN+1)
		for j,_ := range x {
			y[j] = est_acc_taus[i][j] - acc_taus[i][j]
		}
		draw_graph(DL_LEN+1, x, y, fmt.Sprintf("sum_err_%03v.png", i), "Bin Number", "INL (ps)")
	}
}
