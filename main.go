package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"

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

func main() {
	addr := flag.String("a", "", "Camera ONVIF API address/port (example: 192.168.1.100:8080)")
	user := flag.String("u", "admin", "Camera username")
	pass := flag.String("p", "admin", "Camera password")
	flag.Parse()

	if *addr == "" || *user == "" || *pass == "" {
		flag.Usage()
		os.Exit(1)
	}

	dev, err := goonvif.NewDevice(goonvif.DeviceParams{
		Xaddr:    *addr,
		Username: *user,
		Password: *pass,
	})
	if err != nil {
		panic(err)
	}

	// Get device info
	deviceInfoResp, err := dev.CallMethod(device.GetDeviceInformation{})
	if err != nil {
		log.Printf("Failed to get device information: %v", err)
	} else {
		defer deviceInfoResp.Body.Close()
		var envInfo Envelope
		if err := xml.NewDecoder(deviceInfoResp.Body).Decode(&envInfo); err == nil && envInfo.Body.GetDeviceInformationResponse != nil {
			info := envInfo.Body.GetDeviceInformationResponse
			fmt.Printf("Device Information:\nManufacturer: %s\nModel: %s\nFirmware: %s\nSerial: %s\nHardwareID: %s\n\n",
				info.Manufacturer, info.Model, info.FirmwareVersion, info.SerialNumber, info.HardwareId)
		}
	}

	// Get profiles
	profilesResp, err := dev.CallMethod(media.GetProfiles{})
	if err != nil {
		log.Fatalf("Failed to get profiles: %v", err)
	}
	defer profilesResp.Body.Close()

	var envProfiles Envelope
	if err := xml.NewDecoder(profilesResp.Body).Decode(&envProfiles); err != nil || envProfiles.Body.GetProfilesResponse == nil {
		log.Fatalf("Failed to parse profiles response: %v", err)
	}

	for _, profile := range envProfiles.Body.GetProfilesResponse.Profiles {
		fmt.Printf("Profile Name: %s\n", profile.Name)
		if profile.VideoEncoderConfiguration != nil {
			fmt.Printf("Resolution: %dx%d\nFrame Rate: %d\n",
				profile.VideoEncoderConfiguration.Resolution.Width,
				profile.VideoEncoderConfiguration.Resolution.Height,
				profile.VideoEncoderConfiguration.FrameRateLimit)
		}

		getStreamUri := media.GetStreamUri{
			ProfileToken: onvif.ReferenceToken(profile.Token),
			StreamSetup: onvif.StreamSetup{
				Stream:    "RTP-Unicast",
				Transport: onvif.Transport{Protocol: "RTSP"},
			},
		}
		streamUriResp, err := dev.CallMethod(getStreamUri)
		if err != nil {
			log.Printf("Failed to get stream URI for profile %s: %v", profile.Name, err)
			continue
		}

		var envStream Envelope
		if err := xml.NewDecoder(streamUriResp.Body).Decode(&envStream); err != nil || envStream.Body.GetStreamUriResponse == nil {
			log.Printf("Failed to parse stream URI response for profile %s: %v", profile.Name, err)
			streamUriResp.Body.Close()
			continue
		}
		fmt.Printf("Stream URI: %s\n\n", envStream.Body.GetStreamUriResponse.MediaUri.Uri)
		streamUriResp.Body.Close()
	}
}
