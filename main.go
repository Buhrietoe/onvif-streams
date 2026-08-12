package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type Config struct {
	Addr      string
	User      string
	Pass      string
	Timeout   time.Duration
	Retries   int
	Output    string
	Format    string
	NamesOnly bool
}

func envDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseConfig() Config {
	cfg := Config{
		Timeout: 2 * time.Second,
		Retries: 0,
		Format:  "json",
		User:    envDefault("ONVIF_USER", "admin"),
		Pass:    envDefault("ONVIF_PASS", "admin"),
	}

	flag.StringVar(&cfg.Addr, "a", "", "Camera address (e.g. 192.168.1.100:8080)")
	flag.StringVar(&cfg.User, "u", cfg.User, "Camera username (default from ONVIF_USER or 'admin')")
	flag.StringVar(&cfg.Pass, "p", cfg.Pass, "Camera password (default from ONVIF_PASS or 'admin')")
	flag.DurationVar(&cfg.Timeout, "t", cfg.Timeout, "UDP discovery timeout (e.g. 2s, 5s)")
	flag.IntVar(&cfg.Retries, "r", cfg.Retries, "Number of retries for SOAP calls")
	flag.StringVar(&cfg.Output, "o", "", "Output file (default: stdout)")
	flag.StringVar(&cfg.Format, "f", cfg.Format, "Output format: json or table")
	flag.BoolVar(&cfg.NamesOnly, "n", false, "Show only device names (implies -f table)")

	flag.Parse()

	if cfg.NamesOnly {
		cfg.Format = "table"
	}

	if cfg.Format != "json" && cfg.Format != "table" {
		fmt.Fprintf(os.Stderr, "Error: invalid format '%s'. Use 'json' or 'table'\n", cfg.Format)
		os.Exit(1)
	}

	return cfg
}

func main() {
	cfg := parseConfig()

	var devices []*DiscoveredDevice
	if cfg.Addr != "" {
		if dev := processDevice(cfg.Addr, cfg.User, cfg.Pass, cfg.Retries); dev != nil {
			devices = append(devices, dev)
		}
	} else {
		found, err := discoverONVIFDevices(cfg.Timeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: discovery failed: %v\n", err)
		}
		for _, xaddr := range found {
			if dev := processDevice(xaddr, cfg.User, cfg.Pass, cfg.Retries); dev != nil {
				devices = append(devices, dev)
			}
		}
	}

	var out *os.File
	if cfg.Output != "" {
		var err error
		out, err = os.Create(cfg.Output)
		if err != nil {
			log.Fatalf("Failed to create output file: %v", err)
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}

	if cfg.Format == "table" || cfg.NamesOnly {
		printTable(out, devices, cfg.NamesOnly)
	} else {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		enc.Encode(devices)
	}
}

func printTable(out io.Writer, devices []*DiscoveredDevice, namesOnly bool) {
	if len(devices) == 0 {
		fmt.Fprintln(out, "No devices found.")
		return
	}

	if namesOnly {
		for _, d := range devices {
			name := d.Manufacturer
			if d.Model != "" {
				name = fmt.Sprintf("%s %s", d.Manufacturer, d.Model)
			}
			fmt.Fprintln(out, name)
		}
		return
	}

	for _, d := range devices {
		fmt.Fprintf(out, "%s\n", d.Address)
		fmt.Fprintf(out, "  Manufacturer:   %s\n", d.Manufacturer)
		fmt.Fprintf(out, "  Model:          %s\n", d.Model)
		fmt.Fprintf(out, "  Firmware:       %s\n", d.FirmwareVersion)
		fmt.Fprintf(out, "  Serial:         %s\n", d.SerialNumber)
		fmt.Fprintf(out, "  Hardware ID:    %s\n", d.HardwareId)
		if len(d.Profiles) > 0 {
			fmt.Fprintln(out, "  Profiles:")
			for i, p := range d.Profiles {
				fmt.Fprintf(out, "    %d. %s\n", i+1, p.Name)
				if p.Resolution != "" {
					fmt.Fprintf(out, "       Resolution: %s\n", p.Resolution)
				}
				if p.FrameRate > 0 {
					fmt.Fprintf(out, "       Frame Rate: %d fps\n", p.FrameRate)
				}
				if p.StreamURI != "" {
					fmt.Fprintf(out, "       URI:        %s\n", p.StreamURI)
				}
			}
		}
		fmt.Fprintln(out)
	}
}
