package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mcastelino/goexperiments/ipcbench"
	"github.com/wcharczuk/go-chart"
)

func main() {
	var uri = flag.String("uri", "/tmp/ipc_test_uri", "URI of the IPC mechanism")
	var network = flag.String("network", "unix", `networks are
	     "tcp", "tcp4" (IPv4-only),
	     "tcp6" (IPv6-only), "udp", "udp4" (IPv4-only),
	     "udp6" (IPv6-only), "ip", "ip4" (IPv4-only),
	     "ip6" (IPv6-only),
	     "unix", "unixgram" and "unixpacket"`)
	var notifications = flag.Int("notifications", 100, "Number of notifications used to benchmark")
	var chartFile = flag.String("chart", "chart.PNG", "chart file to render latencies")

	flag.Parse()

	notifier := &ipcbench.UnixNotify{}

	if err := notifier.CreateNotifier(*uri, *network, *notifications); err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	defer notifier.Close()
	min, max, avg, latencies, err := notifier.Notify()

	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	var XValues, YValues []float64

	for i, v := range latencies {
		XValues = append(XValues, float64(i))
		YValues = append(YValues, float64(v.Nanoseconds()/1000))
	}

	graph := chart.Chart{
		Title:      *network,
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
				XValues: XValues,
				YValues: YValues,
			},
		},
	}

	fmt.Printf("Network: [%s] Notifications[%v]\n", *network, *notifications)
	fmt.Println("Min Latency :", min)
	fmt.Println("Max Latency :", max)
	fmt.Println("Avg Latency :", avg)
	//fmt.Println("Latencies   ", latencies)

	f, err := os.Create(*chartFile)
	if err != nil {
		fmt.Println("Error creating the graph file")
	}

	err = graph.Render(chart.PNG, f)

	if err != nil {
		fmt.Println("Error rendering the graph file")
	}
	return
}
