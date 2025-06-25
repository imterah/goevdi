package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"time"

	"git.terah.dev/imterah/goevdi/libevdi"
)

func main() {
	log.Print("opening EVDI device")
	dev, err := libevdi.Open(nil)

	if err != nil {
		log.Fatal(err)
	}

	eventHandler := &libevdi.EvdiEventContext{
		UpdateReadyHandler: func(bufferToBeUpdated int) {
			log.Printf("recieved update ready")
		},
	}

	rect := &libevdi.EvdiDisplayRect{
		X1: 0,
		Y1: 0,
		X2: 1920,
		Y2: 1080,
	}

	log.Print("reading EDID file")
	edid, err := os.ReadFile("test.bin")

	if err != nil {
		log.Fatal(err)
	}

	log.Print("attempting to create tempdir")
	os.Mkdir("photos", 0755)

	log.Print("connecting to EVDI device")
	dev.Connect(edid, 1920, 1080, 120)

	log.Print("registering event handler and creating buffer")
	dev.RegisterEventHandler(eventHandler)
	buffer := dev.CreateBuffer(1920, 1080, 4, rect)

	img := &image.RGBA{
		Stride: 1920 * 4,
		Rect:   image.Rect(0, 0, 1920, 1080),
	}

	for frame := range 100 {
		time.Sleep(1 * time.Second)
		pending := !dev.RequestUpdate(buffer)

		if pending {
			if err := dev.BlockUntilOnReady(); err != nil {
				log.Fatal(err)
			}

			dev.HandleEvents(eventHandler)
		}

		dev.GrabPixels(rect)

		img.Pix = buffer.Buffer
		f, err := os.OpenFile(fmt.Sprintf("photos/frame-%03d.png", frame), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)

		if err != nil {
			log.Fatal(err)
		}

		defer f.Close()

		if err := png.Encode(f, img); err != nil {
			log.Fatal(err)
		}

		log.Printf("wrote frame %-3d", frame)
	}
}
