# onvif-streams

Discovers ONVIF-compatible devices on your local network and displays their available media stream URIs.

If the `-a` flag is omitted, the tool will perform a network auto-discovery to find all compatible devices.

The output is a JSON array of all discovered devices and their stream URIs.

## Usage

```
Usage of onvif-streams:
  -a string
    	Camera address (e.g. 192.168.1.100:8080)
  -p string
    	Camera password (default "admin")
  -u string
    	Camera username (default "admin")
```

## Install

    go install -v github.com/Buhrietoe/onvif-streams@latest

## Example Output

```json
[
  {
    "address": "192.168.1.100:8080",
    "manufacturer": "Example Manufacturer",
    "model": "IP-CAM-01",
    "firmware_version": "1.2.3",
    "serial_number": "SN123456789",
    "hardware_id": "HW-ABC-123",
    "profiles": [
      {
        "name": "Profile_1",
        "token": "profile_token_1",
        "resolution": "1920x1080",
        "frame_rate": 30,
        "stream_uri": "rtsp://192.168.1.100/stream1"
      }
    ]
  },
  {
    "address": "192.168.1.101:8899",
    "manufacturer": "Another Manufacturer",
    "model": "PTZ-CAM-PRO",
    "firmware_version": "2.0.1",
    "serial_number": "SN987654321",
    "hardware_id": "HW-XYZ-456",
    "profiles": [
      {
        "name": "mainStream",
        "token": "main_stream_token",
        "resolution": "1280x720",
        "frame_rate": 25,
        "stream_uri": "rtsp://192.168.1.101:554/stream1"
      },
      {
        "name": "subStream",
        "token": "sub_stream_token",
        "resolution": "640x480",
        "frame_rate": 15,
        "stream_uri": "rtsp://192.168.1.101:554/stream2"
      }
    ]
  }
]
```


