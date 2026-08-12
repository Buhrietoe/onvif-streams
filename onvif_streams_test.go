package main

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"strings"
	"testing"
)

func TestDiscoveryResponseUnmarshal(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<e:Envelope xmlns:e="http://www.w3.org/2003/05/soap-envelope">
 <e:Body>
  <d:ProbeMatches xmlns:d="http://schemas.xmlsoap.org/ws/2005/04/discovery">
   <d:ProbeMatch>
    <wsa:Address xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing">uuid:123</wsa:Address>
    <d:Types>dn:NetworkVideoTransmitter</d:Types>
    <d:XAddrs>http://192.168.1.100:8080/onvif/device_service</d:XAddrs>
   </d:ProbeMatch>
  </d:ProbeMatches>
 </e:Body>
</e:Envelope>`

	var resp DiscoveryResponse
	if err := xml.Unmarshal([]byte(xmlData), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(resp.ProbeMatches.ProbeMatch) != 1 {
		t.Fatalf("expected 1 probe match, got %d", len(resp.ProbeMatches.ProbeMatch))
	}

	if !strings.Contains(resp.ProbeMatches.ProbeMatch[0].XAddrs, "192.168.1.100:8080") {
		t.Errorf("unexpected XAddrs: %s", resp.ProbeMatches.ProbeMatch[0].XAddrs)
	}
}

func TestGetDeviceInformationResponseUnmarshal(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<e:Envelope xmlns:e="http://www.w3.org/2003/05/soap-envelope">
 <e:Body>
  <GetDeviceInformationResponse xmlns="http://www.onvif.org/ver10/device/wsdl">
   <Manufacturer>TestCam</Manufacturer>
   <Model>TC-100</Model>
   <FirmwareVersion>1.2.3</FirmwareVersion>
   <SerialNumber>SN123456</SerialNumber>
   <HardwareId>HW-001</HardwareId>
  </GetDeviceInformationResponse>
 </e:Body>
</e:Envelope>`

	var env Envelope
	if err := xml.Unmarshal([]byte(xmlData), &env); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if env.Body.GetDeviceInformationResponse == nil {
		t.Fatal("GetDeviceInformationResponse is nil")
	}

	info := env.Body.GetDeviceInformationResponse
	if info.Manufacturer != "TestCam" {
		t.Errorf("expected Manufacturer TestCam, got %s", info.Manufacturer)
	}
	if info.Model != "TC-100" {
		t.Errorf("expected Model TC-100, got %s", info.Model)
	}
	if info.FirmwareVersion != "1.2.3" {
		t.Errorf("expected FirmwareVersion 1.2.3, got %s", info.FirmwareVersion)
	}
	if info.SerialNumber != "SN123456" {
		t.Errorf("expected SerialNumber SN123456, got %s", info.SerialNumber)
	}
	if info.HardwareId != "HW-001" {
		t.Errorf("expected HardwareId HW-001, got %s", info.HardwareId)
	}
}

func TestGetProfilesResponseUnmarshal(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<e:Envelope xmlns:e="http://www.w3.org/2003/05/soap-envelope">
 <e:Body>
  <GetProfilesResponse xmlns="http://www.onvif.org/ver10/media/wsdl">
   <Profiles token="profile1">
    <Name>Main Stream</Name>
    <VideoEncoderConfiguration>
     <Resolution>
      <Width>1920</Width>
      <Height>1080</Height>
     </Resolution>
     <RateControl>
      <FrameRateLimit>30</FrameRateLimit>
     </RateControl>
    </VideoEncoderConfiguration>
   </Profiles>
  </GetProfilesResponse>
 </e:Body>
</e:Envelope>`

	var env Envelope
	if err := xml.Unmarshal([]byte(xmlData), &env); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if env.Body.GetProfilesResponse == nil {
		t.Fatal("GetProfilesResponse is nil")
	}

	profiles := env.Body.GetProfilesResponse.Profiles
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}

	p := profiles[0]
	if p.Name != "Main Stream" {
		t.Errorf("expected Name 'Main Stream', got %s", p.Name)
	}
	if p.Token != "profile1" {
		t.Errorf("expected Token 'profile1', got %s", p.Token)
	}
	if p.VideoEncoderConfiguration == nil {
		t.Fatal("VideoEncoderConfiguration is nil")
	}
	if p.VideoEncoderConfiguration.Resolution.Width != 1920 {
		t.Errorf("expected Width 1920, got %d", p.VideoEncoderConfiguration.Resolution.Width)
	}
	if p.VideoEncoderConfiguration.FrameRateLimit != 30 {
		t.Errorf("expected FrameRateLimit 30, got %d", p.VideoEncoderConfiguration.FrameRateLimit)
	}
}

func TestGetStreamUriResponseUnmarshal(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<e:Envelope xmlns:e="http://www.w3.org/2003/05/soap-envelope">
 <e:Body>
  <GetStreamUriResponse xmlns="http://www.onvif.org/ver10/media/wsdl">
   <MediaUri>
    <Uri>rtsp://192.168.1.100/stream1</Uri>
   </MediaUri>
  </GetStreamUriResponse>
 </e:Body>
</e:Envelope>`

	var env Envelope
	if err := xml.Unmarshal([]byte(xmlData), &env); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if env.Body.GetStreamUriResponse == nil {
		t.Fatal("GetStreamUriResponse is nil")
	}

	uri := env.Body.GetStreamUriResponse.MediaUri.Uri
	if uri != "rtsp://192.168.1.100/stream1" {
		t.Errorf("expected URI rtsp://192.168.1.100/stream1, got %s", uri)
	}
}

func TestDiscoveredDeviceJSONTags(t *testing.T) {
	device := &DiscoveredDevice{
		Address:         "192.168.1.100:8080",
		Manufacturer:    "TestCam",
		Model:           "TC-100",
		FirmwareVersion: "1.2.3",
		SerialNumber:    "SN123456",
		HardwareId:      "HW-001",
		Profiles: []DiscoveredProfile{
			{
				Name:       "Main",
				Token:      "profile1",
				Resolution: "1920x1080",
				FrameRate:  30,
				StreamURI:  "rtsp://192.168.1.100/stream1",
			},
		},
	}

	data, err := json.Marshal(device)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	jsonStr := string(data)
	expectedFields := []string{
		`"address"`,
		`"firmware_version"`,
		`"serial_number"`,
		`"hardware_id"`,
		`"frame_rate"`,
		`"stream_uri"`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON missing expected field %s", field)
		}
	}
}

func TestPrintTableEmptyDevices(t *testing.T) {
	var buf strings.Builder
	printTable(&buf, nil, false)
	if buf.String() != "No devices found.\n" {
		t.Errorf("unexpected output for empty devices: %q", buf.String())
	}
}

func TestPrintTableNamesOnly(t *testing.T) {
	devices := []*DiscoveredDevice{
		{Manufacturer: "TestCam", Model: "TC-100"},
		{Manufacturer: "OtherCam", Model: ""},
	}

	var buf strings.Builder
	printTable(&buf, devices, true)
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	if lines[0] != "TestCam TC-100" {
		t.Errorf("expected 'TestCam TC-100', got %q", lines[0])
	}

	if lines[1] != "OtherCam" {
		t.Errorf("expected 'OtherCam', got %q", lines[1])
	}
}

func TestEnvDefault(t *testing.T) {
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	if got := envDefault("TEST_VAR", "fallback"); got != "test_value" {
		t.Errorf("expected 'test_value', got %q", got)
	}

	if got := envDefault("NONEXISTENT_VAR", "fallback"); got != "fallback" {
		t.Errorf("expected 'fallback', got %q", got)
	}
}
