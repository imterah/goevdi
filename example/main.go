// Portions of this code are from the UnrealXR project:
// https://git.terah.dev/UnrealXR/unrealxr
package main

import (
	"image/color"
	"log"
	"math"
	"os"
	"time"
	"unsafe"

	"git.terah.dev/imterah/goevdi/libevdi"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func findMaxVerticalSize(fovyDeg float32, distance float32) float32 {
	fovyRad := float64(fovyDeg * math.Pi / 180.0)
	return 2 * distance * float32(math.Tan(fovyRad/2))
}

func findOptimalHorizontalRes(verticalDisplayRes float32, horizontalDisplayRes float32, verticalSize float32) float32 {
	aspectRatio := horizontalDisplayRes / verticalDisplayRes
	horizontalSize := verticalSize * aspectRatio

	return horizontalSize
}

func main() {
	log.Print("opening EVDI device")
	dev, err := libevdi.Open(nil)

	if err != nil {
		log.Fatal(err)
	}

	eventHandler := &libevdi.EvdiEventContext{}

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

	timeoutDuration := 0 * time.Millisecond

	// HACK: sometimes the buffer doesn't get initialized properly if we don't wait a bit...
	time.Sleep(250 * time.Millisecond)

	rl.InitWindow(1920, 1080, "GoEVDI Bindings Example")
	defer rl.CloseWindow()

	rl.SetTargetFPS(120)

	// Do a 3D scene
	// Unnecessary, but I implemented this already, so we're going for it
	fovY := float32(45.0)
	verticalSize := findMaxVerticalSize(fovY, 5.0)

	camera := rl.NewCamera3D(
		rl.Vector3{
			X: 0.0,
			Y: verticalSize / 2,
			Z: 5.0,
		},
		rl.Vector3{
			X: 0.0,
			Y: verticalSize / 2,
			Z: 0.0,
		},
		rl.Vector3{
			X: 0.0,
			Y: 1.0,
			Z: 0.0,
		},
		fovY,
		rl.CameraPerspective,
	)

	horizontalSize := findOptimalHorizontalRes(1080, 1920, verticalSize)
	coreMesh := rl.GenMeshPlane(horizontalSize, verticalSize, 1, 1)

	image := rl.NewImage(buffer.Buffer, 1920, 1080, 1, rl.UncompressedR8g8b8a8)

	texture := rl.LoadTextureFromImage(image)
	model := rl.LoadModelFromMesh(coreMesh)

	rl.SetMaterialTexture(model.Materials, rl.MapAlbedo, texture)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		rl.BeginMode3D(camera)

		ready, err := dev.WaitUntilEventsAreReadyToHandle(timeoutDuration)

		if err != nil {
			log.Fatalf("WaitUntilEventsAreReadyToHandle: %v", err)
		}

		if ready {
			if err := dev.HandleEvents(eventHandler); err != nil {
				log.Fatalf("HandleEvents: %v", err)
			}

			dev.GrabPixels(rect)
			pixels := unsafe.Slice(
				(*color.RGBA)(unsafe.Pointer(&buffer.Buffer[0])),
				len(buffer.Buffer)/4,
			)

			rl.UpdateTexture(texture, pixels)
			dev.RequestUpdate(buffer)
		}

		rl.DrawModelEx(
			model,
			rl.Vector3{
				X: 0,
				Y: verticalSize / 2,
				Z: 0,
			},
			// rotate around X to make it vertical
			rl.Vector3{
				X: 1,
				Y: 0,
				Z: 0,
			},
			90,
			rl.Vector3{
				X: 1,
				Y: 1,
				Z: 1,
			},
			rl.White,
		)

		rl.EndMode3D()
		rl.EndDrawing()
	}

	log.Printf("Goodbye!")
	rl.CloseWindow()
}
