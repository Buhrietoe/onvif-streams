package main

import (
	"encoding/xml"
	"fmt"

	goonvif "github.com/use-go/onvif"
	"github.com/use-go/onvif/media"
	"github.com/use-go/onvif/xsd/onvif"
	"github.com/use-go/onvif/device"
)

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
