package main

import (
	"encoding/xml"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

func discoverONVIFDevices(timeout time.Duration) ([]string, error) {
	probe := `<?xml version="1.0" encoding="UTF-8"?>
<e:Envelope xmlns:e="http://www.w3.org/2003/05/soap-envelope"
 xmlns:w="http://schemas.xmlsoap.org/ws/2004/08/addressing"
 xmlns:d="http://schemas.xmlsoap.org/ws/2005/04/discovery"
 xmlns:dn="http://www.onvif.org/ver10/network/wsdl">
 <e:Header>
  <w:MessageID>uuid:` + generateUUID() + `</w:MessageID>
  <w:To>urn:schemas-xmlsoap-org:ws:2005:04:discovery</w:To>
  <w:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</w:Action>
 </e:Header>
 <e:Body>
  <d:Probe>
   <d:Types>dn:NetworkVideoTransmitter dn:NetworkVideoRecorder dn:MediaServer</d:Types>
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
	u, err := uuid.NewV4()
	if err != nil {
		return "00000000-0000-0000-0000-000000000000"
	}
	return u.String()
}
