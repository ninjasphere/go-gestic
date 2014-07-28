package agent

import (
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/joshlf13/gopack"
	"github.com/wolfeidau/epoller"
)

const (
	DSPIfoFlag uint16 = 1 << iota
	GestureInfoFlag
	TouchInfoFlag
	AirWheelInfoFlag
	CoordinateInfoFlag
)

// Gestic
// http://ww1.microchip.com/downloads/en/DeviceDoc/40001718B.pdf
// Page 36

// Gestic device path
const GesticDevicePath = "/dev/gestic"

// Flag which indicates if the payload contains data
const SensorDataPresentFlag = 0x91

const (
	MaxEpollEvents     = 32
	MaxMessageSize     = 255
	IdSensorDataOutput = 0x91
)

type Reader struct {
	currentGesture *gestureData
}

type gestureData struct {
	Event       EventHeader
	Data        DataHeader
	Gesture     GestureInfo
	Touch       TouchInfo
	AirWheel    AirWheelInfo
	Coordinates CoordinateInfo
}

type EventHeader struct {
	Length, Flags, Seq, Id uint8
}

type DataHeader struct {
	DataMask              uint16
	TimeStamp, SystemInfo uint8
}

type GestureInfo struct {
	Gesture uint32
}

type TouchInfo struct {
	Touch uint32
}

type AirWheelInfo struct {
	AirWheel uint8
	Crap     uint8
}

type CoordinateInfo struct {
	X uint8
	Y uint8
	Z uint8
}

type GestureCoordinates struct {
	X, Y, Z int
}

// decoded gesture
type Gesture struct {
	Gesture     *string
	Touch       *string
	AirWheel    *int
	Coordinates *GestureCoordinates
}

var Gestures = []string{
	"No gesture",
	"Garbage model",
	"Flick West to East",
	"Flick East to West",
	"Flick South to North",
	"Flick North to South",
	"Circle clockwise",
	"Circle counter-clockwise",
}

func (gi *GestureInfo) GetGestureName() string {
	return Gestures[int(gi.Gesture)]
}

func (r *Reader) Start() {
	log.Printf("Opening %s", GesticDevicePath)

	r.currentGesture = &gestureData{}

	if err := epoller.OpenAndDispatchEvents(GesticDevicePath, r.buildGestureEvent); err != nil {
		log.Fatalf("Error opening device reader %v", err)
	}
}

func (r *Reader) buildGestureEvent(buf []byte, n int) {

	gopack.Unpack(buf[:4], r.currentGesture.Event)
	gopack.Unpack(buf[4:8], r.currentGesture.Data)

	// var for offset
	offset := 8

	// grab the DSPIfo
	if r.currentGesture.Data.DataMask&DSPIfoFlag == DSPIfoFlag {
		offset += 2
	}

	// grab the GestureInfo
	if r.currentGesture.Data.DataMask&GestureInfoFlag == GestureInfoFlag {

		gopack.Unpack(buf[offset:offset+4], r.currentGesture.Gesture)
		//		n := gestureInfo.Gesture & 0xff

		offset += 4

	}

	// grab the TouchInfo
	if r.currentGesture.Data.DataMask&TouchInfoFlag == TouchInfoFlag {
		gopack.Unpack(buf[offset:offset+4], r.currentGesture.Touch)
		offset += 4
	}

	// grab the AirWheelInfo
	if r.currentGesture.Data.DataMask&AirWheelInfoFlag == AirWheelInfoFlag {
		gopack.Unpack(buf[offset:offset+2], r.currentGesture.AirWheel)
		offset += 2
	}

	// grab the CoordinateInfo
	if r.currentGesture.Data.DataMask&CoordinateInfoFlag == CoordinateInfoFlag {
		gopack.Unpack(buf[offset:offset+6], r.currentGesture.Coordinates)
		offset += 6
	}

	log.Printf("%s", spew.Sdump(r.currentGesture))

}
