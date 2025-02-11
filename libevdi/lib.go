package lib

// #include "evdi_lib.h"
// #include "go_ffi.h"
// #cgo CFLAGS: -w
import "C"
import (
	"fmt"
	"os"
	"sync"
	"unsafe"

	"log"
)

var (
	// Buffer allocation
	currentBufNumber   = 0
	bufferRegisterLock = sync.Mutex{}

	// EVDI Event data
	cEventToGoEventMapping = map[unsafe.Pointer]*EvdiEventContext{}
	activeLogger           = &EvdiLogger{
		Log: func(msg string) {
			log.Printf("evdi: %s", msg)
		},
	}
)

type EvdiLogger struct {
	Log func(message string)
}

type EvdiMode struct {
	Width        int
	Height       int
	RefreshRate  int
	BitsPerPixel int

	PixelFormat uint
}

type EvdiCursorSet struct {
	HotX        int32
	HotY        int32
	Width       uint32
	Height      uint32
	Enabled     uint8
	Buffer      []byte
	PixelFormat uint32
	Stride      uint32
}

type DDCCIData struct {
	Address uint16
	Flags   uint16
	Buffer  []byte
}

type EvdiEventContext struct {
	cEventContext *C.struct_evdi_event_context

	DPMSHandler        func(dpmsMode int)
	ModeChangeHandler  func(mode *EvdiMode)
	UpdateReadyHandler func(bufferToBeUpdated int)
	CRTCStateHandler   func(state int)
	CursorSetHandler   func(cursor *EvdiCursorSet)
	CursorMoveHandler  func(x, y int32)
	DDCCIDataHandler   func(ddcciData *DDCCIData)
}

type EvdiBuffer struct {
	ID     int
	Buffer []byte
	Width  int
	Height int
	Stride int

	internalEvdiBuffer *C.struct_evdi_buffer
}

type EvdiDisplayRect struct {
	X1 int
	Y1 int
	X2 int
	Y2 int

	hasCInit     bool
	cDisplayRect *C.struct_evdi_rect
}

//export goDPMSHandler
func goDPMSHandler(event C.int, userData unsafe.Pointer) {
	goData, ok := cEventToGoEventMapping[userData]

	if !ok {
		panic("could not find Go event from C event map for EvdiEventContext")
	}

	goData.DPMSHandler(int(event))
}

//export goModeChangedHandler
func goModeChangedHandler(mode C.struct_evdi_mode, userData unsafe.Pointer) {
	goData, ok := cEventToGoEventMapping[userData]

	if !ok {
		panic("could not find Go event from C event map for EvdiEventContext")
	}

	goData.ModeChangeHandler(&EvdiMode{
		Width:        int(mode.width),
		Height:       int(mode.height),
		RefreshRate:  int(mode.refresh_rate),
		BitsPerPixel: int(mode.bits_per_pixel),
		PixelFormat:  uint(mode.pixel_format),
	})
}

//export goUpdateReadyHandler
func goUpdateReadyHandler(event C.int, userData unsafe.Pointer) {
	goData, ok := cEventToGoEventMapping[userData]

	if !ok {
		panic("could not find Go event from C event map for EvdiEventContext")
	}

	goData.UpdateReadyHandler(int(event))
}

//export goCRTCStateHandler
func goCRTCStateHandler(event C.int, userData unsafe.Pointer) {
	goData, ok := cEventToGoEventMapping[userData]

	if !ok {
		panic("could not find Go event from C event map for EvdiEventContext")
	}

	goData.UpdateReadyHandler(int(event))
}

//export goCursorSetHandler
func goCursorSetHandler(cursor C.struct_evdi_cursor_set, userData unsafe.Pointer) {
	goData, ok := cEventToGoEventMapping[userData]

	if !ok {
		panic("could not find Go event from C event map for EvdiEventContext")
	}

	pointerBuffer := unsafe.Slice((*byte)(unsafe.Pointer(cursor.buffer)), cursor.buffer_length)

	// Clone it for good measure
	safeBuffer := make([]byte, cursor.buffer_length)
	copy(safeBuffer, pointerBuffer)

	goData.CursorSetHandler(&EvdiCursorSet{
		HotX:        int32(cursor.hot_x),
		HotY:        int32(cursor.hot_y),
		Width:       uint32(cursor.width),
		Height:      uint32(cursor.height),
		Enabled:     uint8(cursor.enabled),
		Buffer:      safeBuffer,
		PixelFormat: uint32(cursor.pixel_format),
		Stride:      uint32(cursor.stride),
	})
}

//export goCursorMoveHandler
func goCursorMoveHandler(cursor C.struct_evdi_cursor_move, userData unsafe.Pointer) {
	goData, ok := cEventToGoEventMapping[userData]

	if !ok {
		panic("could not find Go event from C event map for EvdiEventContext")
	}

	goData.CursorMoveHandler(int32(cursor.x), int32(cursor.y))
}

//export goDDCCIDataHandler
func goDDCCIDataHandler(data C.struct_evdi_ddcci_data, userData unsafe.Pointer) {
	goData, ok := cEventToGoEventMapping[userData]

	if !ok {
		panic("could not find Go event from C event map for EvdiEventContext")
	}

	pointerBuffer := unsafe.Slice((*byte)(unsafe.Pointer(data.buffer)), data.buffer_length)

	// Clone it for good measure
	safeBuffer := make([]byte, data.buffer_length)
	copy(safeBuffer, pointerBuffer)

	goData.DDCCIDataHandler(&DDCCIData{
		Address: uint16(data.address),
		Flags:   uint16(data.flags),
		Buffer:  safeBuffer,
	})
}

//export goLoggerHandler
func goLoggerHandler(message *C.char) {
	activeLogger.Log(C.GoString(message))
	C.free(unsafe.Pointer(message))
}

type EvdiNode struct {
	handle C.evdi_handle
}

func (node *EvdiNode) Disconnect() {
	C.evdi_disconnect(node.handle)
}

func (node *EvdiNode) Close() {
	C.evdi_disconnect(node.handle)
}

func (node *EvdiNode) RegisterEventHandler(handler *EvdiEventContext) {
	if handler.cEventContext == nil {
		handler.cEventContext = &C.struct_evdi_event_context{
			dpms_handler:         (*[0]byte)(C.dpmsHandler),
			mode_changed_handler: (*[0]byte)(C.modeChangedHandler),
			update_ready_handler: (*[0]byte)(C.updateReadyHandler),
			crtc_state_handler:   (*[0]byte)(C.crtcStateHandler),
			cursor_set_handler:   (*[0]byte)(C.cursorSetHandler),
			cursor_move_handler:  (*[0]byte)(C.cursorMoveHandler),
			ddcci_data_handler:   (*[0]byte)(C.ddcciDataHandler),
		}

		handler.cEventContext.user_data = unsafe.Pointer(handler.cEventContext)
	}

	cEventToGoEventMapping[unsafe.Pointer(handler.cEventContext)] = handler
}

func (node *EvdiNode) UnregisterEventHandler(handler *EvdiEventContext) error {
	if _, ok := cEventToGoEventMapping[unsafe.Pointer(handler.cEventContext)]; !ok {
		return fmt.Errorf("could not find event map")
	}

	delete(cEventToGoEventMapping, unsafe.Pointer(handler.cEventContext))

	if handler.cEventContext == nil {
		return fmt.Errorf("cEventContext pointer is somehow nil! Please report this bug at https://git.terah.dev/imterah/goevdi")
	}

	C.free(unsafe.Pointer(handler.cEventContext))

	return nil
}

func (node *EvdiNode) HandleEvents(handler *EvdiEventContext) error {
	if _, ok := cEventToGoEventMapping[unsafe.Pointer(handler.cEventContext)]; ok {
		return fmt.Errorf("could not find event map")
	}

	C.evdi_handle_events(node.handle, handler.cEventContext)

	return nil
}

func (node *EvdiNode) Connect(EDID []byte, pixelWidthLimit, pixelHeightLimit, FPSLimit uint) {
	rawCEDID := C.CString(string(EDID))
	cEDID := (*C.uchar)(unsafe.Pointer(rawCEDID))

	defer C.free(unsafe.Pointer(cEDID))

	pixelAreaLimit := pixelWidthLimit * pixelHeightLimit

	C.evdi_connect2(node.handle, cEDID, C.uint(uint(len(EDID))), C.uint(pixelAreaLimit), C.uint(pixelAreaLimit*FPSLimit))
}

func (node *EvdiNode) EnableCursorEvents(enable bool) {
	C.evdi_enable_cursor_events(node.handle, C.bool(enable))
}

func (node *EvdiNode) GetOnReadyFile() *os.File {
	fdC := C.evdi_get_event_ready(node.handle)
	fd := int(fdC)

	file := os.NewFile(uintptr(fd), "evdi-fd")

	return file
}

func (node *EvdiNode) CreateBuffer(width, height, stride int, rect *EvdiDisplayRect) *EvdiBuffer {
	bufferRegisterLock.Lock()
	defer bufferRegisterLock.Unlock()

	cBuffer := C.malloc(C.size_t(width * height * stride))
	normalBuffer := unsafe.Slice((*byte)(cBuffer), width*height*stride)

	evdiRect := C.struct_evdi_rect{
		x1: C.int(rect.X1),
		x2: C.int(rect.X2),
		y1: C.int(rect.Y1),
		y2: C.int(rect.Y2),
	}

	rect.hasCInit = true
	rect.cDisplayRect = &evdiRect

	evdiBuffer := C.struct_evdi_buffer{
		id:     C.int(currentBufNumber),
		buffer: cBuffer,
		width:  C.int(width),
		height: C.int(height),
		stride: C.int(stride),

		rects:      &evdiRect,
		rect_count: C.int(0),
	}

	buf := &EvdiBuffer{
		ID:     currentBufNumber,
		Buffer: normalBuffer,
		Width:  width,
		Height: height,
		Stride: stride,

		internalEvdiBuffer: &evdiBuffer,
	}

	C.evdi_register_buffer(node.handle, evdiBuffer)
	currentBufNumber++

	return buf
}

func (node *EvdiNode) RemoveBuffer(buffer *EvdiBuffer) {
	C.evdi_unregister_buffer(node.handle, C.int(buffer.ID))

	buffer.Buffer = nil
	C.free(buffer.internalEvdiBuffer.buffer)
}

func (node *EvdiNode) GrabPixels(rect *EvdiDisplayRect) int {
	rectNumCIntPointer := C.malloc(C.sizeof_int)
	defer C.free(rectNumCIntPointer)

	rectNumCInt := (*C.int)(rectNumCIntPointer)

	if !rect.hasCInit || rect.cDisplayRect == nil || int(rect.cDisplayRect.x1) != rect.X1 || int(rect.cDisplayRect.x2) != rect.X2 || int(rect.cDisplayRect.y1) != rect.Y1 || int(rect.cDisplayRect.y2) != rect.Y2 {
		evdiRect := C.struct_evdi_rect{
			x1: C.int(rect.X1),
			x2: C.int(rect.X2),
			y1: C.int(rect.Y1),
			y2: C.int(rect.Y2),
		}

		rect.hasCInit = true
		rect.cDisplayRect = &evdiRect
	}

	C.evdi_grab_pixels(node.handle, rect.cDisplayRect, rectNumCInt)

	rectNum := int(*rectNumCInt)
	return rectNum
}

func (node *EvdiNode) RequestUpdate(buffer *EvdiBuffer) {
	C.evdi_request_update(node.handle, C.int(buffer.ID))
}

// Creates a new EVDI node. Loosely based on `evdi_open_attached_to_fixed()`.
func Open(parentDevice *string) (*EvdiNode, error) {
	var parentCString *C.char
	length := 0

	if parentDevice != nil {
		parentCString = C.CString(*parentDevice)
		length = len(*parentDevice)

		defer C.free(unsafe.Pointer(parentCString))
	}

	handle := C.evdi_open_attached_to_fixed(parentCString, C.size_t(uint(length)))

	if handle == nil {
		return nil, fmt.Errorf("failed to initialize EVDI node")
	}

	node := &EvdiNode{
		handle: handle,
	}

	return node, nil
}

// Checks if Xorg is running. Based on the C function `Xorg_running()`.
func IsXorgRunning() bool {
	xorgRunningC := C.Xorg_running()

	return bool(xorgRunningC)
}

// Gets the underlying library version. Based on the C function `evdi_get_lib_version()`.
//
// 1st int: major version
//
// 2nd int: minor version
//
// 3rd int: patch level
func GetLibraryVersion() (int, int, int) {
	version := C.struct_evdi_lib_version{}
	C.evdi_get_lib_version(&version)

	return int(version.version_major), int(version.version_minor), int(version.version_patchlevel)
}

func SetupLogger(logger *EvdiLogger) {
	activeLogger = logger
}

func init() {
	C.loggerInit()
}
