package main

import (
	"encoding/xml"
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
	Profiles        []DiscoveredProfile `json:"profiles,omitempty"`
}

type DiscoveredProfile struct {
	Name       string `json:"name,omitempty"`
	Token      string `json:"token,omitempty"`
	Resolution string `json:"resolution,omitempty"`
	FrameRate  int    `json:"frame_rate,omitempty"`
	StreamURI  string `json:"stream_uri,omitempty"`
}
