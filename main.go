package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	goonvif "github.com/use-go/onvif"
	"github.com/use-go/onvif/device"
	"github.com/use-go/onvif/media"
	"github.com/use-go/onvif/xsd/onvif"
)

type Envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    Body     `xml:"Body"`
}

type Body struct {
	GetProfilesResponse          *GetProfilesResponse          `xml:"GetProfilesResponse"`
	GetStreamUriResponse         *GetStreamUriResponse         `xml:"GetStreamUriResponse"`
	GetDeviceInformationResponse *GetDeviceInformationResponse `xml:"GetDeviceInformationResponse"`
}

type GetProfilesResponse struct {
	Profiles []Profile `xml:"Profiles"`
}

type Profile struct {
	Name                      string                     `xml:"Name"`
	Token                     string                     `xml:"token,attr"`
	VideoEncoderConfiguration *VideoEncoderConfiguration `xml:"VideoEncoderConfiguration"`
}

type VideoEncoderConfiguration struct {
	Resolution     Resolution `xml:"Resolution"`
	FrameRateLimit int        `xml:"RateControl>FrameRateLimit"`
}

type Resolution struct {
	Width  int `xml:"Width"`
	Height int `xml:"Height"`
}

type GetStreamUriResponse struct {
	MediaUri MediaUri `xml:"MediaUri"`
}

type MediaUri struct {
	Uri string `xml:"Uri"`
}

type GetDeviceInformationResponse struct {
	Manufacturer    string `xml:"Manufacturer"`
	Model           string `xml:"Model"`
	FirmwareVersion string `xml:"FirmwareVersion"`
	SerialNumber    string `xml:"SerialNumber"`
	HardwareId      string `xml:"HardwareId"`
}

type ProbeMatch struct {
	XAddrs string `xml:"XAddrs"`
	Types  string `xml:"Types"`
}

type ProbeMatches struct {
	ProbeMatch []ProbeMatch `xml:"ProbeMatch"`
}

type DiscoveryResponse struct {
	ProbeMatches ProbeMatches `xml:"Body>ProbeMatches"`
}

type DiscoveredDevice struct {
	Address         string              `json:"address,omitempty"`
	Manufacturer    string              `json:"manufacturer,omitempty"`
	Model           string              `json:"model,omitempty"`
	FirmwareVersion string              `json:"firmware_version,omitempty"`
	SerialNumber    string              `json:"serial_number,omitempty"`
	HardwareId      string              `json:"hardware_id,omitempty"`
	Profiles        []DiscoveredProfile `json:"profile,omitemptys"`
}

type DiscoveredProfile struct {
	Name       string `json:"name,omitempty"`
	Token      string `json:"token,omitempty"`
	Resolution string `json:"resolution,omitempty"`
	FrameRate  int    `json:"frame_rate,omitempty"`
	StreamURI  string `json:"stream_uri,omitempty"`
}

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

func processDevice(addr, user, pass string) *DiscoveredDevice {
	dev, err := goonvif.NewDevice(goonvif.DeviceParams{
		Xaddr:    addr,
		Username: user,
		Password: pass,
	})
	if err != nil {
		return nil
	}

	deviceInfoResp, err := dev.CallMethod(device.GetDeviceInformation{})
	if err != nil {
		return nil
	}
	defer deviceInfoResp.Body.Close()
	var envInfo Envelope
	if err := xml.NewDecoder(deviceInfoResp.Body).Decode(&envInfo); err != nil || envInfo.Body.GetDeviceInformationResponse == nil {
		return nil
	}
	info := envInfo.Body.GetDeviceInformationResponse
	result := &DiscoveredDevice{
		Address:         addr,
		Manufacturer:    info.Manufacturer,
		Model:           info.Model,
		FirmwareVersion: info.FirmwareVersion,
		SerialNumber:    info.SerialNumber,
		HardwareId:      info.HardwareId,
	}

	profilesResp, err := dev.CallMethod(media.GetProfiles{})
	if err != nil {
		return result
	}
	defer profilesResp.Body.Close()
	var envProfiles Envelope
	if err := xml.NewDecoder(profilesResp.Body).Decode(&envProfiles); err != nil || envProfiles.Body.GetProfilesResponse == nil {
		return result
	}

	for _, profile := range envProfiles.Body.GetProfilesResponse.Profiles {
		p := DiscoveredProfile{
			Name:  profile.Name,
			Token: profile.Token,
		}
		if profile.VideoEncoderConfiguration != nil {
			p.Resolution = fmt.Sprintf("%dx%d", profile.VideoEncoderConfiguration.Resolution.Width, profile.VideoEncoderConfiguration.Resolution.Height)
			p.FrameRate = profile.VideoEncoderConfiguration.FrameRateLimit
		}
		getStreamUri := media.GetStreamUri{
			ProfileToken: onvif.ReferenceToken(profile.Token),
			StreamSetup: onvif.StreamSetup{
				Stream:    "RTP-Unicast",
				Transport: onvif.Transport{Protocol: "RTSP"},
			},
		}
		streamUriResp, err := dev.CallMethod(getStreamUri)
		if err == nil {
			var envStream Envelope
			if xml.NewDecoder(streamUriResp.Body).Decode(&envStream) == nil && envStream.Body.GetStreamUriResponse != nil {
				p.StreamURI = envStream.Body.GetStreamUriResponse.MediaUri.Uri
			}
			streamUriResp.Body.Close()
		}
		result.Profiles = append(result.Profiles, p)
	}
	return result
}

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
