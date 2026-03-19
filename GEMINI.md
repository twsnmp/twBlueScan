# twBlueScan

Bluetooth Device Sensor for TWSNMP FC.

## Project Overview

`twBlueScan` is a Go-based Bluetooth device sensor designed for Linux environments. Its primary purpose is to scan for nearby Bluetooth devices and send collected information to TWSNMP FC (or any other syslog receiver) via syslog. 

### Key Features

- **Bluetooth Scanning**: Detects Bluetooth devices, including their MAC address, address type (public/random), name, RSSI, and vendor.
- **Sensor Support**: Specialized decoding for OMRON environment sensors and various SwitchBot devices (Temperature/Humidity, CO2, Plug Mini, Motion Sensor).
- **Resource Monitoring**: Periodically sends system resource metrics (CPU, Memory, Load, Network) of the host machine.
- **Syslog Integration**: Sends data to multiple syslog destinations.
- **Vendor Mapping**: Includes logic to map MAC addresses and company codes to their respective manufacturers.

### Core Technologies

- **Language**: Go (1.16+)
- **Libraries**:
  - `bluewalker`: Low-level Bluetooth scanning.
  - `gopsutil`: System resource monitoring.
  - `google/uuid`: UUID handling.
- **Platform**: Linux (requires `bluez` and root/sudo for raw socket access).

---

## Building and Running

### Build Commands

Building is managed via the `Makefile`.

- **Build all**: `make` or `make all` (generates binaries for Linux amd64, arm, and arm64 in `dist/`).
- **Clean**: `make clean`
- **Release (ZIP)**: `make zip` (packages binaries into ZIP files).
- **Test**: `make test`

### Running the Sensor

The sensor typically requires root privileges to access Bluetooth adapters.

```bash
sudo ./twBlueScan -adapter hci0 -syslog 192.168.1.1
```

**Common Flags:**
- `-adapter <string>`: Bluetooth adapter to use (default: "hci0").
- `-syslog <string>`: Comma-separated list of syslog destinations (e.g., `192.168.1.1,192.168.1.2:5514`).
- `-interval <int>`: Syslog reporting interval in seconds (default: 600).
- `-active`: Enable active scan mode (required for some sensors like SwitchBot).
- `-debug`: Enable debug logging.

---

## Architecture and File Structure

- `main.go`: Entry point. Handles flags, environment variables, OS signals, and orchestrates the scanning and syslog goroutines.
- `blueScan.go`: Core scanning logic. Manages the device map, decodes Bluetooth advertisement data, and handles sensor-specific (OMRON/SwitchBot) parsing.
- `syslog.go`: Manages UDP syslog connections and the message queue.
- `monitor.go`: Collects and reports system resource usage (CPU, RAM, Net).
- `vendor.go`: Contains large maps for mapping company codes and MAC prefixes to vendor names.
- `Makefile`: Defines build targets for cross-platform Linux support.

---

## Development Conventions

- **Goroutines**: Uses goroutines for concurrent scanning and syslog delivery, coordinated via a `context.Context`.
- **Concurrency**: Protects shared state (like device lists) using `sync.Map`.
- **Formatting**: Adheres to standard Go formatting.
- **Cross-Compilation**: The `Makefile` explicitly handles `GOOS=linux` for various architectures (`amd64`, `arm`, `arm64`).
- **Versioning**: Version and commit hashes are injected at build time using `-ldflags`.
