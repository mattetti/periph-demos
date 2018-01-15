// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/experimental/devices/ht16k33"
	"periph.io/x/periph/host"
)

func mainImpl() error {
	if _, err := host.Init(); err != nil {
		return fmt.Errorf("couldn't init the host - %s", err)
	}

	i2cAddr := flag.Uint("ia", 0x70, "I²C bus address to use")
	nbrDevices := flag.Int("devices", 1, "The number of devices to connect to")
	flag.Parse()

	i2cBus, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer i2cBus.Close()
	if p, ok := i2cBus.(i2c.Pins); ok {
		printPin("SCL", p.SCL())
		printPin("SDA", p.SDA())
	} else {
		log.Println("i2cBus.(i2c.Pins) failed")
	}

	matrixAddresses := []uint{0x70, 0x72}

	dev, err := newDevice(i2cBus, *i2cAddr)
	if err != nil {
		log.Fatal(err)
	}
	dev.ClearAll()
	log.Printf("Requesting to connect to %d devices\n", *nbrDevices)
	devs := []*ht16k33.Dev{dev}
	if *nbrDevices > 1 {
		for i := 1; i < *nbrDevices; i++ {
			addr := matrixAddresses[i]
			mdev, err := newDevice(i2cBus, addr)
			if err != nil {
				log.Fatalf("Failed to connect to device at %#x - %v\n", addr, err)
			}
			log.Printf("Other device at address: %#x\n", addr)
			devs = append(devs, mdev)
		}
	}

	for _, pad := range devs {
		for i := 0; i < 16; i++ {
			pad.SetLED(i, true)
			if err := pad.WriteDisplay(); err != nil {
				panic(err)
			}
			time.Sleep(50 * time.Millisecond)
		}
		time.Sleep(500 * time.Millisecond)
	}
	dev.ClearAll()

	for {
		if err := dev.ReadKeys(); err != nil {
			log.Fatal(err)
		}
		changed := false
		for i := 0; i < 16; i++ {
			if dev.IsPressed(i) {
				turnRow(dev, i, true)
				changed = true
			} else if dev.WasJustReleased(i) {
				turnRow(dev, i, false)
				changed = true
			}
		}
		if changed {
			dev.WriteDisplay()
		}
		time.Sleep(30 * time.Millisecond)
	}

	return nil
}

// turnRow turns the entire row of a specific button on or off.
// WriteDisplay must be called to push the change to the device.
func turnRow(dev *ht16k33.Dev, idx int, isOn bool) {
	if idx < 4 {
		dev.SetLED(0, isOn)
		dev.SetLED(1, isOn)
		dev.SetLED(2, isOn)
		dev.SetLED(3, isOn)
		return
	}
	if idx < 8 {
		dev.SetLED(4, isOn)
		dev.SetLED(5, isOn)
		dev.SetLED(6, isOn)
		dev.SetLED(7, isOn)
		return
	}
	if idx < 12 {
		dev.SetLED(8, isOn)
		dev.SetLED(9, isOn)
		dev.SetLED(10, isOn)
		dev.SetLED(11, isOn)
		return
	}
	dev.SetLED(12, isOn)
	dev.SetLED(13, isOn)
	dev.SetLED(14, isOn)
	dev.SetLED(15, isOn)
}

func newDevice(i2cBus i2c.Bus, i2cAddr uint) (*ht16k33.Dev, error) {
	opts := ht16k33.DefaultOpts()
	opts.Debug = true
	opts.I2CAddr = uint16(i2cAddr)

	return ht16k33.NewI2C(i2cBus, opts)
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "ht16k33: %s.\n", err)
		os.Exit(1)
	}
}

func printPin(fn string, p pin.Pin) {
	name, pos := pinreg.Position(p)
	if name != "" {
		log.Printf("  %-4s: %-10s found on header %s, #%d\n", fn, p, name, pos)
	} else {
		log.Printf("  %-4s: %-10s\n", fn, p)
	}
}
