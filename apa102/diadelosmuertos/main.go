package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"periph.io/x/extra/devices/screen"

	"periph.io/x/periph/conn/display"

	"periph.io/x/periph/conn/physic"

	"github.com/maruel/anim1d"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/devices/apa102"
	"periph.io/x/periph/host"
)

type DrawerWriter interface {
	display.Drawer
	io.Writer
}

func mainImpl() error {
	verbose := flag.Bool("v", false, "verbose mode")
	fake := flag.Bool("terminal", false, "print the animation at the terminal")
	spiID := flag.String("spi", "", "SPI port to use")
	hz := flag.Int64("hz", 0, "SPI port speed")
	numPixels := flag.Int("n", 44, "number of pixels on the strip")
	intensity := flag.Int("l", 127, "light intensity [1-255]")
	temperature := flag.Int("t", 5000, "light temperature in °Kelvin [3500-7500]")
	fps := flag.Int("fps", 30, "frames per second")
	fileName := flag.String("f", "", "file to load the animation from")
	raw := flag.String("r", "", "inline serialized animation")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)
	if flag.NArg() != 0 {
		return errors.New("unexpected argument, try -help")
	}
	if *intensity < 1 || *intensity > 255 {
		return errors.New("intensity must be between 1 and 255")
	}
	if *temperature < 0 || *temperature > 65535 {
		return errors.New("temperature must be between 0 and 65535")
	}
	if *numPixels < 1 || *numPixels > 10000 {
		return errors.New("number of pixels must be between 1 and 10000")
	}
	if *fps < 1 || *fps > 200 {
		return errors.New("fps must be between 1 and 200")
	}
	var pat anim1d.SPattern
	if *fileName != "" {
		if *raw != "" {
			return errors.New("can't use both -f and -r")
		}
		c, err := ioutil.ReadFile(*fileName)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(c, &pat); err != nil {
			return err
		}
	} else if *raw != "" {
		// cache := &anim1d.ThumbnailsCache{NumberLEDs: *numPixels, ThumbnailHz: *fps, ThumbnailSeconds: 2}
		// gif, err := cache.GIF([]byte(*raw))
		// if err != nil {
		// 	panic(err)
		// }
		// of, err := os.Create("gif.gif")
		// if err != nil {
		// 	panic(err)
		// }
		// of.Write(gif)
		// of.Close()
		if err := json.Unmarshal([]byte(*raw), &pat); err != nil {
			return err
		}
	} else {
		return errors.New("use one of -f or -r; try -r '\"0101ff\"'")
	}

	var display DrawerWriter
	if *fake {
		display = screen.New(int(*numPixels))
	} else {
		if _, err := host.Init(); err != nil {
			return err
		}
		s, err := spireg.Open(*spiID)
		if err != nil {
			if *spiID == "" {
				return fmt.Errorf("use -terminal if you don't have LEDs; error opening SPI: %v", err)
			}
			return err
		}
		defer s.Close()
		if *hz != 0 {
			if err := s.LimitSpeed(physic.Frequency(*hz)); err != nil {
				return err
			}
		}
		opts := &apa102.Opts{NumPixels: *numPixels, Intensity: uint8(*intensity), Temperature: uint16(*temperature)}
		display, err = apa102.New(s, opts)
		if err != nil {
			return err
		}
		defer display.Halt()
	}

	return runLoop(display, pat.Pattern, *fps)
}

func runLoop(display DrawerWriter, p anim1d.Pattern, fps int) error {
	delta := time.Second / time.Duration(fps)
	numLights := display.Bounds().Dx()
	buf := make([]byte, numLights*3)
	f := make(anim1d.Frame, numLights)
	t := time.NewTicker(delta)
	start := time.Now()
	for {
		// Wraps after 49.71 days.
		p.Render(f, uint32(time.Since(start)/time.Millisecond))
		f.ToRGB(buf)
		if _, err := display.Write(buf); err != nil {
			return err
		}
		<-t.C
	}
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "anim1d: %s.\n", err)
		os.Exit(1)
	}
}
