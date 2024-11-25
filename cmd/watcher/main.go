package main

import (
	"log"
	"time"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

// TimeCharacteristicUUID is the UUID for the time characteristic on PineTime.
// Replace with the actual UUID from your PineTime device.
const TimeCharacteristicUUID = "00002a2b-0000-1000-8000-00805f9b34fb"

func main() {
	// Enable Bluetooth on the adapter.
	err := adapter.Enable()
	if err != nil {
		log.Fatalf("Failed to enable Bluetooth adapter: %v", err)
	}

	log.Println("Scan for InfiniTime...")
	address, err := scanForAddress("InfiniTime")
	if err != nil {
		log.Fatalf("Error scanning for device: %v", err)
	}
	log.Printf("Found address %s\n", address.String())

	dev, err := adapter.Connect(*address, bluetooth.ConnectionParams{})
	if err != nil {
		log.Fatalf("Error connecting to device: %v", err)
	}
	defer dev.Disconnect()

	services, err := dev.DiscoverServices(nil)
	if err != nil {
		log.Fatalf("Error discovering services: %v", err)
	}

	var timeChar *bluetooth.DeviceCharacteristic

	for _, service := range services {
		chars, err := service.DiscoverCharacteristics(nil)
		if err != nil {
			log.Fatalf("service.DiscoverCharacteristics: %v", err)
		}
		for _, c := range chars {
			if c.UUID().String() == TimeCharacteristicUUID {
				timeChar = &c
				break
			}
		}

		if timeChar != nil {
			break
		}
	}

	if timeChar == nil {
		log.Fatal("time characteristic not found")
	}

	currentTime := buildCurrentTime()
	log.Printf("Setting time: %x\n", currentTime)

	_, err = timeChar.WriteWithoutResponse(currentTime)
	if err != nil {
		log.Fatalf("timeChar.WriteWithoutResponse: %v", err)
	}
}

func scanForAddress(name string) (*bluetooth.Address, error) {
	var addr *bluetooth.Address

	err := adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		if device.LocalName() == name {
			addr = &device.Address
			adapter.StopScan()
		}
	})
	if err != nil {
		return nil, err
	}
	return addr, nil
}

// buildCurrentTime builds the current time as a byte slice in CTS format:
// YYYY MM DD HH MM SS DOW
func buildCurrentTime() []byte {
	now := time.Now()
	year := uint16(now.Year())
	month := uint8(now.Month())
	day := uint8(now.Day())
	hour := uint8(now.Hour())
	minute := uint8(now.Minute())
	second := uint8(now.Second())
	dayOfWeek := uint8((int(now.Weekday()) + 6) % 7) // Sunday=0, Monday=1, ..., Saturday=6 -> 1=Monday, ..., 7=Sunday
	if dayOfWeek == 0 {
		dayOfWeek = 7
	}

	// Build the time byte array.
	timeBytes := []byte{
		byte(year & 0xFF), byte((year >> 8) & 0xFF), // Year (little-endian)
		month,
		day,
		hour,
		minute,
		second,
		dayOfWeek,
	}
	return timeBytes
}
