package gestic

import (
	"log"
	"math"

	"github.com/bitly/go-simplejson"
	"github.com/joshlf13/gopack"
	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-ninja/logger"
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
	IdSensorDataOutput = 0x91
)

type Reader struct {
	conn           *ninja.NinjaConnection
	log            *logger.Logger
	currentGesture *gestureData
}

func NewReader(conn *ninja.NinjaConnection, log *logger.Logger) *Reader {
	return &Reader{conn: conn, log: log}
}

type gestureData struct {
	Event       *EventHeader
	DataHeader  *DataHeader
	Gesture     *GestureInfo
	Touch       *TouchInfo
	AirWheel    *AirWheelInfo
	Coordinates *CoordinateInfo
}

func NewGestureData() *gestureData {
	return &gestureData{
		Event:       &EventHeader{},
		DataHeader:  &DataHeader{},
		Gesture:     &GestureInfo{},
		Touch:       &TouchInfo{},
		AirWheel:    &AirWheelInfo{},
		Coordinates: &CoordinateInfo{},
	}
}

type EventHeader struct {
	Length, Flags, Seq, Id uint8
}

type DataHeader struct {
	DataMask              uint16
	TimeStamp, SystemInfo uint8
}

type GestureInfo struct {
	GestureVal uint32
}

func (gi *GestureInfo) Name() string {
	if int(gi.GestureVal) < len(Gestures) {
		return Gestures[gi.GestureVal]
	} else {
		return Gestures[0]
	}
}

type TouchInfo struct {
	TouchVal uint32
}

func (ti *TouchInfo) Name() string {
	if ti.TouchVal > 0 {
		i := math.Log(float64(ti.TouchVal)) / math.Log(2)
		return TouchList[int(i)]
	}
	return "None"
}

type AirWheelInfo struct {
	AirWheelVal uint8
	Crap        uint8
}

type CoordinateInfo struct {
	X uint16
	Y uint16
	Z uint16
}

var Gestures = []string{
	"None",
	"Garbage",
	"WestToEast",
	"EastToWest",
	"SouthToNorth",
	"NorthToSouth",
	"CircleClockwise",
	"CircleCounterClockwise",
}

var TouchList = []string{
	"TouchSouth",
	"TouchWest",
	"TouchNorth",
	"TouchEast",
	"TouchCenter",
	"TapSouth",
	"TapWest",
	"TapNorth",
	"TapEast",
	"TapCenter",
	"DoubleTapSouth",
	"DoubleTapWest",
	"DoubleTapNorth",
	"DoubleTapEast",
	"DoubleTapCenter",
}

func (r *Reader) Start() {
	r.log.Infof("Opening %s", GesticDevicePath)

	r.currentGesture = NewGestureData()

	if err := epoller.OpenAndDispatchEvents(GesticDevicePath, r.buildGestureEvent); err != nil {
		log.Fatalf("Error opening device reader %v", err)
	}
}

func (r *Reader) buildGestureEvent(buf []byte, n int) {

	g := r.currentGesture

	gopack.Unpack(buf[:4], g.Event)
	gopack.Unpack(buf[4:8], g.DataHeader)

	// var for offset
	offset := 8

	// grab the DSPIfo
	if g.DataHeader.DataMask&DSPIfoFlag == DSPIfoFlag {
		offset += 2
	}

	// grab the GestureInfo
	if g.DataHeader.DataMask&GestureInfoFlag == GestureInfoFlag {

		gopack.Unpack(buf[offset:offset+4], g.Gesture)
		g.Gesture.GestureVal = g.Gesture.GestureVal & uint32(0xff)
		offset += 4
	}

	// grab the TouchInfo
	if g.DataHeader.DataMask&TouchInfoFlag == TouchInfoFlag {
		gopack.Unpack(buf[offset:offset+4], g.Touch)
		offset += 4
	}

	// grab the AirWheelInfo
	if g.DataHeader.DataMask&AirWheelInfoFlag == AirWheelInfoFlag {
		gopack.Unpack(buf[offset:offset+2], g.AirWheel)
		offset += 2
	}

	// grab the CoordinateInfo
	if g.DataHeader.DataMask&CoordinateInfoFlag == CoordinateInfoFlag {
		gopack.Unpack(buf[offset:offset+6], g.Coordinates)
		offset += 6
	}

	r.log.Debugf("Gesture: %s, Airwheel: %d, Touch: %s", g.Gesture.Name(), g.AirWheel.AirWheelVal, g.Touch.Name())

	r.publishCurrentGesture()
}

func (r *Reader) publishCurrentGesture() {

	g := r.currentGesture

	if g.Gesture.GestureVal > 0 {
		jsonmsg, _ := simplejson.NewJson([]byte(`{}`))
		jsonmsg.Set("gesture", g.Gesture.Name())
		r.conn.PublishRPCMessage("$client/gesture/gesture", jsonmsg)
	}

	if g.Touch.TouchVal > 0 {
		jsonmsg, _ := simplejson.NewJson([]byte(`{}`))
		jsonmsg.Set("touch", g.Touch.Name())
		r.conn.PublishRPCMessage("$client/gesture/touch", jsonmsg)
	}

	if g.AirWheel.AirWheelVal > 0 {
		jsonmsg, _ := simplejson.NewJson([]byte(`{}`))
		jsonmsg.Set("airwheel", g.AirWheel.AirWheelVal)
		r.conn.PublishRPCMessage("$client/gesture/airwheel", jsonmsg)
	}

	if g.Coordinates.X != 0 || g.Coordinates.Y != 0 || g.Coordinates.Z != 0 {
		jsonmsg, _ := simplejson.NewJson([]byte(`{}`))
		jsonmsg.Set("x", g.Coordinates.X)
		jsonmsg.Set("y", g.Coordinates.Y)
		jsonmsg.Set("z", g.Coordinates.Z)
		r.conn.PublishRPCMessage("$client/gesture/position", jsonmsg)
	}
}
