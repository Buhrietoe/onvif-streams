package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"
)

func main() {
	addr := flag.String("a", "", "Camera address (e.g. 192.168.1.100:8080)")
	user := flag.String("u", "admin", "Camera username")
	pass := flag.String("p", "admin", "Camera password")
	flag.Parse()

	var devices []*DiscoveredDevice
	if *addr != "" {
		if dev := processDevice(*addr, *user, *pass); dev != nil {
			devices = append(devices, dev)
		}
	} else {
		found, err := discoverONVIFDevices(2 * time.Second)
		if err != nil {
			log.Fatalf("Discovery failed: %v", err)
		}
		for _, xaddr := range found {
			if dev := processDevice(xaddr, *user, *pass); dev != nil {
				devices = append(devices, dev)
			}
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(devices)
}
