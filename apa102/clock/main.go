// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// apa102 writes to a strip of APA102 LED.
package main

import (
	"errors"
	"flag"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/devices/apa102"
	"periph.io/x/periph/host"
)

func access(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func findFile(name string) string {
	if access(name) {
		return name
	}
	for _, p := range strings.Split(os.Getenv("GOPATH"), ":") {
		if len(p) != 0 {
			if p2 := filepath.Join(p, "src/periph.io/x/periph/cmd/apa102", name); access(p2) {
				return p2
			}
		}
	}
	return ""
}

func mainImpl() error {
	verbose := flag.Bool("v", false, "verbose mode")
	spiID := flag.String("spi", "", "SPI port to use")

	numLights := flag.Int("n", 150, "number of lights on the strip")
	intensity := flag.Int("l", 127, "light intensity [1-255]")
	temperature := flag.Int("t", 5000, "light temperature in °Kelvin [3500-7500]")
	hz := flag.Int("hz", 0, "SPI port speed")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)
	if flag.NArg() != 0 {
		return errors.New("unexpected argument, try -help")
	}
	if *intensity > 255 {
		return errors.New("max intensity is 255")
	}
	if *temperature > 65535 {
		return errors.New("max temperature is 65535")
	}
	if _, err := host.Init(); err != nil {
		return err
	}

	// Open the display device.
	s, err := spireg.Open(*spiID)
	if err != nil {
		return err
	}
	defer s.Close()
	if *hz != 0 {
		if err := s.LimitSpeed(int64(*hz)); err != nil {
			return err
		}
	}
	if p, ok := s.(spi.Pins); ok {
		// TODO(maruel): Print where the pins are located.
		log.Printf("Using pins CLK: %s  MOSI: %s  MISO: %s", p.CLK(), p.MOSI(), p.MISO())
	}
	display, err := apa102.New(s, *numLights, uint8(*intensity), uint16(*temperature))
	if err != nil {
		return err
	}

	buf := []byte{}
	ticker := time.NewTicker(time.Second)
	for _ = range ticker.C {
		// TODO: optimize the buffer allocation
		buf = randomBufArray(*numLights)
		_, err = display.Write(buf)
		if err != nil {
			log.Printf("Failed to write color: %q, %s\n", buf, err)
		}
	}
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "apa102: %s.\n", err)
		os.Exit(1)
	}
}

func randomBufArray(numLights int) []byte {
	rgb := rand.Intn(16777216)
	r := byte(rgb >> 16)
	g := byte(rgb >> 8)
	b := byte(rgb)
	buf := make([]byte, numLights*3)
	for i := 0; i < len(buf); i += 3 {
		buf[i] = r
		buf[i+1] = g
		buf[i+2] = b
	}
	return buf
}
