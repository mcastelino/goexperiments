package main

/*
#define _GNU_SOURCE
#include <sched.h>
#include <pthread.h>
#include <signal.h>
#include <unistd.h>
#include <stdio.h>
#include <sys/types.h>
#include <sys/syscall.h>

int
checkLatency(int param)
{
	return (param+1);
}

*/
import "C"

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/wcharczuk/go-chart"
)

func checkGoLatency(param int) int {
	return param + 1
}

func benchmark(benchmarkGo bool, cycles int) (minLatency, maxLatency, avgLatency time.Duration, latencies []time.Duration, err error) {

	var max, avg time.Duration
	var now, then time.Time
	min := time.Hour

	for i := 0; i < cycles; i++ {
		var ret int

		// Send it out and wait for the reply
		// We are measuring round trip time
		// as we want to use the same reference time
		if benchmarkGo {
			then = time.Now()
			ret = checkGoLatency(i)
			now = time.Now()
		} else {
			then = time.Now()
			ret = int(C.checkLatency(C.int(i)))
			now = time.Now()
		}

		if int(ret) != i+1 {
			return min, max, avg, latencies, fmt.Errorf("ERROR: cgo call verification failure", i, ret)
		}

		latency := now.Sub(then)
		avg = avg + latency

		latencies = append(latencies, latency)

		if latency > max {
			max = latency
		}
		if latency < min {
			min = latency
		}
	}

	avgLat := time.Duration((avg.Nanoseconds() / int64(cycles)))
	return min, max, avgLat, latencies, nil
}

func main() {
	var cycles = flag.Int("cycles", 100, "Number of cgo call cycles used to benchmark")
	var rounds = flag.Int("rounds", 3, "Number of rounds used to benchmark")
	var chartFile = flag.String("chart", "chart.PNG", "file to graph the latency")

	flag.Parse()

	var XRValues, YRValues [][]float64
	var min, max, avg time.Duration
	for i := 0; i < *rounds; i++ {
		var latencies []time.Duration
		var err error
		min, max, avg, latencies, err = benchmark(false, *cycles)
		if err != nil {
			fmt.Println(err)
			return
		}

		var XValues, YValues []float64
		for i, v := range latencies {
			XValues = append(XValues, float64(i))
			YValues = append(YValues, float64(v))
		}
		XRValues = append(XRValues, XValues)
		YRValues = append(YRValues, YValues)
	}

	minGo, maxGo, avgGo, goLatencies, err := benchmark(true, *cycles)
	if err != nil {
		fmt.Println(err)
		return
	}

	var XGoValues, YGoValues []float64
	for i, v := range goLatencies {
		XGoValues = append(XGoValues, float64(i))
		YGoValues = append(YGoValues, float64(v))
	}

	graph := chart.Chart{
		Title:      fmt.Sprintf("cgo latency: cycles %d", *cycles),
		TitleStyle: chart.StyleShow(),
		XAxis: chart.XAxis{
			Name:      "Iteration",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
		},
		YAxis: chart.YAxis{
			Name:      "Latency (µs)",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Name:    "cgo",
				Style:   chart.StyleShow(),
				XValues: XRValues[0],
				YValues: YRValues[0],
			},
			chart.ContinuousSeries{
				Name:    "gogo",
				Style:   chart.StyleShow(),
				XValues: XGoValues,
				YValues: YGoValues,
			},
		},
	}

	for i := 1; i < *rounds; i++ {
		s := chart.ContinuousSeries{
			Name:    "cgo" + string(i),
			Style:   chart.StyleShow(),
			XValues: XRValues[i],
			YValues: YRValues[i],
		}
		graph.Series = append(graph.Series, s)
	}

	fmt.Printf("cgo latency: Cycles[%v]\n", *cycles)
	fmt.Println("Min Latency :", min)
	fmt.Println("Max Latency :", max)
	fmt.Println("Avg Latency :", avg)
	fmt.Printf("go latency: Cycles[%v]\n", *cycles)
	fmt.Println("Min Latency :", minGo)
	fmt.Println("Max Latency :", maxGo)
	fmt.Println("Avg Latency :", avgGo)
	//fmt.Println("Latencies   ", latencies)

	f, err := os.Create(*chartFile)
	defer f.Close()
	if err != nil {
		fmt.Println("Error creating the graph file")
	}

	err = graph.Render(chart.PNG, f)

	if err != nil {
		fmt.Println("Error rendering the graph file")
	}
	return
}
