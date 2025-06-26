package main

import (
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

	updateReady := false

	eventHandler := &libevdi.EvdiEventContext{
		UpdateReadyHandler: func(bufferToBeUpdated int) {
			updateReady = true
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

	timeoutDuration := 1 * time.Millisecond
	shouldRequestUpdate := true

	// HACK: sometimes the buffer doesn't get initialized properly if we don't wait a bit...
	time.Sleep(250 * time.Millisecond)

	for frame := range 100 {
		if shouldRequestUpdate {
			dev.RequestUpdate(buffer)
			shouldRequestUpdate = false
		}

		isReady, err := dev.WaitUntilEventsAreReadyToHandle(timeoutDuration)

		if err != nil {
			log.Fatal(err)
		}

		if isReady {
			dev.HandleEvents(eventHandler)
		}

		if updateReady {
			shouldRequestUpdate = true
			updateReady = false
		}

		log.Print("events are ready, continuing...")
		dev.GrabPixels(rect)

		log.Printf("wrote frame %-3d", frame)
	}
}
