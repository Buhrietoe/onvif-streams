package main

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegrationDiscoveryResponse(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "discovery_response.xml"))
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var resp DiscoveryResponse
	if err := xml.Unmarshal(data, &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(resp.ProbeMatches.ProbeMatch) != 2 {
		t.Fatalf("expected 2 probe matches, got %d", len(resp.ProbeMatches.ProbeMatch))
	}

	if !strings.Contains(resp.ProbeMatches.ProbeMatch[0].XAddrs, "192.168.1.100:8080") {
		t.Errorf("unexpected XAddrs for camera: %s", resp.ProbeMatches.ProbeMatch[0].XAddrs)
	}

	if !strings.Contains(resp.ProbeMatches.ProbeMatch[1].XAddrs, "192.168.1.200:80") {
		t.Errorf("unexpected XAddrs for NVR: %s", resp.ProbeMatches.ProbeMatch[1].XAddrs)
	}
}

func TestIntegrationDeviceInfoResponse(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "device_info_response.xml"))
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var env Envelope
	if err := xml.Unmarshal(data, &env); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if env.Body.GetDeviceInformationResponse == nil {
		t.Fatal("GetDeviceInformationResponse is nil")
	}

	info := env.Body.GetDeviceInformationResponse
	if info.Manufacturer != "Hikvision" {
		t.Errorf("expected Manufacturer Hikvision, got %s", info.Manufacturer)
	}
	if info.Model != "DS-2CD2043G0-I" {
		t.Errorf("expected Model DS-2CD2043G0-I, got %s", info.Model)
	}
}

func TestIntegrationProfilesResponse(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "profiles_response.xml"))
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var env Envelope
	if err := xml.Unmarshal(data, &env); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if env.Body.GetProfilesResponse == nil {
		t.Fatal("GetProfilesResponse is nil")
	}

	profiles := env.Body.GetProfilesResponse.Profiles
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}

	mainStream := profiles[0]
	if mainStream.Name != "MainStream" {
		t.Errorf("expected MainStream, got %s", mainStream.Name)
	}
	if mainStream.VideoEncoderConfiguration.Resolution.Width != 2560 {
		t.Errorf("expected Width 2560, got %d", mainStream.VideoEncoderConfiguration.Resolution.Width)
	}

	subStream := profiles[1]
	if subStream.Name != "SubStream" {
		t.Errorf("expected SubStream, got %s", subStream.Name)
	}
	if subStream.VideoEncoderConfiguration.Resolution.Width != 704 {
		t.Errorf("expected Width 704, got %d", subStream.VideoEncoderConfiguration.Resolution.Width)
	}
}

func TestIntegrationStreamUriResponse(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "stream_uri_response.xml"))
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var env Envelope
	if err := xml.Unmarshal(data, &env); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if env.Body.GetStreamUriResponse == nil {
		t.Fatal("GetStreamUriResponse is nil")
	}

	uri := env.Body.GetStreamUriResponse.MediaUri.Uri
	expected := "rtsp://192.168.1.100:554/Streaming/Channels/101"
	if uri != expected {
		t.Errorf("expected URI %s, got %s", expected, uri)
	}
}
