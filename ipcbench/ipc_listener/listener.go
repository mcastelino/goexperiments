package main

import (
	"flag"
	"fmt"

	"github.com/mcastelino/goexperiments/ipcbench"
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

	flag.Parse()

	listener := &ipcbench.UnixNotify{}

	if err := listener.CreateListener(*uri, *network, *notifications); err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	listener.Wait()
	listener.Close()
	return
}
