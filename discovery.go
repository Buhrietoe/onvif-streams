package main

import (
	"encoding/xml"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

func discoverONVIFDevices(timeout time.Duration) ([]string, error) {
	probe := `<?xml version="1.0" encoding="UTF-8"?>
<e:Envelope xmlns:e="http://www.w3.org/2003/05/soap-envelope"
 xmlns:w="http://schemas.xmlsoap.org/ws/2004/08/addressing"
 xmlns:d="http://schemas.xmlsoap.org/ws/2005/04/discovery">
 <e:Header>
  <w:MessageID>uuid:` + generateUUID() + `</w:MessageID>
  <w:To>urn:schemas-xmlsoap-org:ws:2005:04:discovery</w:To>
  <w:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</w:Action>
 </e:Header>
 <e:Body>
  <d:Probe>
   <d:Types>dn:NetworkVideoTransmitter</d:Types>
  </d:Probe>
 </e:Body>
</e:Envelope>`

	addr := net.UDPAddr{
		IP:   net.ParseIP("239.255.255.250"),
		Port: 3702,
	}
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(timeout))

	_, err = conn.WriteToUDP([]byte(probe), &addr)
	if err != nil {
		return nil, err
	}

	var devices []string
	buf := make([]byte, 8192)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				break
			}
			return nil, err
		}
		var resp DiscoveryResponse
		if xml.Unmarshal(buf[:n], &resp) == nil {
			for _, match := range resp.ProbeMatches.ProbeMatch {
				for _, xaddr := range strings.Fields(match.XAddrs) {
					u, err := url.Parse(xaddr)
					if err == nil && u.Host != "" && !contains(devices, u.Host) {
						devices = append(devices, u.Host)
					}
				}
			}
		}
	}
	return devices, nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func generateUUID() string {
	// Simple UUID generator for WS-Discovery (not RFC4122 compliant, but sufficient for this use)
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
