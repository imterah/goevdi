package libdisplayconfig

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

// !! TRANSPARENCY - READ THIS !!
// This code was originally written by AI, with modifications and fixes done by me (Tera).
// This is just to create a basic EDID file for each display, so imho, this is fine.
//
// This is NOT up to spec, but upon testing, neither is Apple's Pro Display XDR.
// The ways that are out of spec *shouldn't* effect us (chromacity coordinates), but this is
// absolutely something worth investigating.
// !! TRANSPARENCY - READ THIS !!

// A display mode, passed into GenerateEDID().
type Mode struct {
	Width   int // active horizontal pixels
	Height  int // active vertical lines
	Refresh int // refresh rate in Hz
}

// GenerateEDID accepts a slice of Mode values and returns a 128-byte EDID block.
//
// modes:
//
// The first mode is treated as the preferred timing. Up to 4 modes are inserted into
// the four available detailed timing descriptor slots. Any extra modes are ignored.
// The basic EDID fields (manufacturer, physical size, etc.) are fixed.
func GenerateEDID(modes []Mode) ([]byte, error) {
	if len(modes) == 0 {
		return nil, errors.New("at least one mode must be provided")
	}

	// Create a 128-byte EDID block.
	edid := make([]byte, 128)

	// --- Header (bytes 0-7): Fixed header "00 FF FF FF FF FF FF 00" ---
	edid[0] = 0x00
	for i := 1; i <= 6; i++ {
		edid[i] = 0xFF
	}
	edid[7] = 0x00

	// --- Manufacturer & Product Identification (bytes 8-11) ---

	// Use the Linux foundation as a manufacturer
	manuID := uint16((12 << 10) | (14 << 5) | 24)
	edid[8] = byte(manuID >> 8)
	edid[9] = byte(manuID & 0xFF)

	// Product code for "Linux FHD"
	binary.LittleEndian.PutUint16(edid[10:12], 490)

	// --- Display Product Serial Number (bytes 12-15) ---
	// Use the ASCII string "EVDI" (for the virtual display software on Linux)
	serial := []byte{'E', 'V', 'D', 'I'}
	copy(edid[12:16], serial)

	// --- Manufacture Week and Year (bytes 16-17) ---
	edid[16] = 5
	edid[17] = 35

	// --- EDID Version and Revision (bytes 18-19) ---
	edid[18] = 1
	edid[19] = 4

	// --- Basic Display Parameters (bytes 20-24) ---
	// Byte 20: Digital input definition.
	// Bit 7 = 1 (digital)
	// Bits 6-4: set to 001 for 6 bits per primary (a valid value for our virtual display)
	// Bits 3-0: set to 0001 for a valid digital interface (DVI)
	edid[20] = 0x80 | (0x1 << 4) | 0x01

	// Bytes 21-22: Maximum image size in centimeters.
	//
	// Emulating a 24 inch (presumably 16:9) monitor here.
	// X is the inches to cm constant, which is 2.54. 9/16 is the ratio to get a likely height.
	height := ((9 / 16) * 24) * 2.54
	width := 24 * 2.54

	edid[21] = byte(height)
	edid[22] = byte(width)

	// Byte 23: Display gamma = (gamma*100)-100; for gamma 2.2, that's 220-100 = 120.
	edid[23] = 220 - 100

	// Byte 24: Feature support; set to a common value, e.g. 0x0A.
	edid[24] = 0x0A

	// --- Chromaticity Coordinates (bytes 25-34) ---
	// Use fixed dummy values.
	chroma := []byte{0x78, 0xEA, 0x3D, 0xA2, 0x57, 0x4A, 0x9C, 0x25, 0x12, 0x50}
	copy(edid[25:35], chroma)

	// --- Established Timings (bytes 35-37) ---
	for i := 35; i < 38; i++ {
		edid[i] = 0x00
	}

	// --- Standard Timings (bytes 38-53) ---
	// Fill with 0x0101 (unused).
	for i := 38; i < 54; i++ {
		edid[i] = 0x01
	}

	// --- Detailed Timing Descriptors (4 slots: each 18 bytes) ---
	currOffsetPosition := 54

	// Calculate the total size after we put all the modes + dummy descriptor in there.
	if 54+((len(modes)+1)*18) > len(edid) {
		return nil, fmt.Errorf("too much modes")
	}

	for _, mode := range modes {
		dtd, err := buildDTD(mode)

		if err != nil {
			return nil, err
		}

		copy(edid[currOffsetPosition:currOffsetPosition+18], dtd)
		currOffsetPosition += 18
	}

	dummy := buildDummyDescriptor()
	copy(edid[currOffsetPosition:currOffsetPosition+18], dummy)

	// --- Number of EDID Extension Blocks (byte 126) ---
	edid[126] = 0x00

	// --- Checksum (byte 127): the sum of all 128 bytes must be 0 mod 256 ---
	sum := 0

	for i := 0; i < 127; i++ {
		sum += int(edid[i])
	}

	edid[127] = byte((256 - (sum % 256)) % 256)

	return edid, nil
}

// buildDTD builds an 18-byte Detailed Timing Descriptor for the given Mode.
func buildDTD(m Mode) ([]byte, error) {
	// Create an 18-byte slice.
	dtd := make([]byte, 18)

	// Active timings.
	hActive := m.Width
	vActive := m.Height

	// Heuristic: horizontal blanking is 15% of active width, at least 8 pixels, rounded to even.
	hBlank := roundEven(float64(m.Width) * 0.15)

	if hBlank < 8 {
		hBlank = 8
	}

	// Calculate horizontal sync offset and width.
	hSyncOffset := roundEven(float64(hBlank) / 4.0)
	hSyncWidth := roundEven(float64(hBlank) / 8.0)

	// Ensure back porch (hBlank - (hSyncOffset + hSyncWidth)) is positive.
	if hBlank <= (hSyncOffset + hSyncWidth) {
		hBlank = hSyncOffset + hSyncWidth + 2
		hSyncOffset = roundEven(float64(hBlank) / 4.0)
		hSyncWidth = roundEven(float64(hBlank) / 8.0)
	}

	totalH := hActive + hBlank

	// Vertical blanking: use 5% of active height, at least 2 lines.
	vBlank := roundEven(float64(m.Height) * 0.05)

	if vBlank < 2 {
		vBlank = 2
	}

	totalV := vActive + vBlank

	// Pixel clock in Hz = totalH * totalV * refresh.
	pixelClockHz := float64(totalH * totalV * m.Refresh)
	pixelClock := uint16(math.Round(pixelClockHz / 10000.0)) // in 10 kHz units

	// Bytes 0-1: Pixel clock, little-endian.
	binary.LittleEndian.PutUint16(dtd[0:2], pixelClock)

	// Bytes 2-4: Horizontal active and blanking.
	dtd[2] = byte(hActive & 0xFF)
	dtd[3] = byte(hBlank & 0xFF)
	dtd[4] = byte(((hActive>>8)&0x0F)<<4 | ((hBlank >> 8) & 0x0F))

	// Bytes 5-7: Vertical active and blanking.
	dtd[5] = byte(vActive & 0xFF)
	dtd[6] = byte(vBlank & 0xFF)
	dtd[7] = byte(((vActive>>8)&0x0F)<<4 | ((vBlank >> 8) & 0x0F))

	// Bytes 8-11: Sync timings.
	// Horizontal sync offset and width.
	dtd[8] = byte(hSyncOffset & 0xFF)
	dtd[9] = byte(hSyncWidth & 0xFF)

	// Byte 10: Upper 2 bits of hSyncOffset and hSyncWidth.
	hsyncOffsetUpper := (hSyncOffset >> 8) & 0x03
	hsyncWidthUpper := (hSyncWidth >> 8) & 0x03

	// For vertical sync offset and width we use fixed values: offset = 3, width = 5.
	dtd[10] = byte((hsyncOffsetUpper << 6) | (hsyncWidthUpper << 4)) // vertical bits assumed 0 in upper nibble

	// Byte 11: Lower nibble: vertical sync offset and width.
	vSyncOffset := 3
	vSyncWidth := 5
	dtd[11] = byte(((vSyncOffset & 0x0F) << 4) | (vSyncWidth & 0x0F))

	// Bytes 12-14: Image size in millimeters.
	// Use fixed physical size consistent with the basic block (400 mm x 300 mm).
	imgWidthMM := 400
	imgHeightMM := 300
	dtd[12] = byte(imgWidthMM & 0xFF)
	dtd[13] = byte(imgHeightMM & 0xFF)
	dtd[14] = byte(((imgWidthMM>>8)&0x0F)<<4 | ((imgHeightMM >> 8) & 0x0F))

	// Bytes 15-16: Border pixels (set to 0).
	dtd[15] = 0x00
	dtd[16] = 0x00

	// Byte 17: Flags. Set to 0 for non-interlaced.
	dtd[17] = 0x00

	return dtd, nil
}

// buildDummyDescriptor returns an 18-byte dummy descriptor.
// Here we fill it with a descriptor tag of 0xFF followed by a dummy string.
func buildDummyDescriptor() []byte {
	dummy := make([]byte, 18)

	// First 3 bytes zero, then tag 0xFF.
	dummy[0] = 0x00
	dummy[1] = 0x00
	dummy[2] = 0x00
	dummy[3] = 0xFF

	// Fill the rest with a dummy ASCII string padded with spaces.
	str := []byte("VirtDisplay.")
	copy(dummy[5:], str)

	for i := 5 + len(str); i < 18; i++ {
		dummy[i] = '#'
	}
	return dummy
}

// roundEven rounds x to the nearest even integer.
func roundEven(x float64) int {
	n := int(math.Round(x))
	if n%2 != 0 {
		return n - 1
	}
	return n
}
