package agent

import (
	"log"
	"os"
	"syscall"

	"github.com/joshlf13/gopack"
)

const (
	DSPIfoFlag         uint16 = 1
	GestureInfoFlag    uint16 = 2
	TouchInfoFlag      uint16 = 4
	AirWheelInfoFlag   uint16 = 8
	CoordinateInfoFlag uint16 = 16
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

var ClickLocations = []{

}

func (gi *GestureInfo) GetGestureName() string {
	return Gestures[int(gi.Gesture)]
}

func (r *Reader) Start() {
	log.Printf("Opening %s", GesticDevicePath)

	fd, err := syscall.Open(GesticDevicePath, os.O_RDWR, 0666)

	rfds := &syscall.FdSet{}
	timeout := &syscall.Timeval{1, 1}

	if err != nil {
		log.Fatalf("Can't open %s - %s", GesticDevicePath, err)
	}

	log.Printf("Reading %s", GesticDevicePath)

	for {

		FD_ZERO(rfds)
		FD_SET(rfds, fd)

		_, err := syscall.Select(fd+1, rfds, nil, nil, timeout)

		if err != nil {
			log.Fatalf("Can't read %s - %s", GesticDevicePath, err)
		}
		//  One of the fds changed
		if FD_ISSET(rfds, int(fd)) {

			buf := make([]byte, 255)
			n, err := syscall.Read(fd, buf)


			BuildGestureEvent(buf)
		}

	}

}

func BuildGestureEvent(buf []byte) Gesture {

	header := &Header{}

	gopack.Unpack(buf[:4], header)

	//log.Printf("header %+v", header)

	dataHeader := &DateHeader{}

	gopack.Unpack(buf[4:8], dataHeader)

	//log.Printf("dataHeader %+v", dataHeader)

	// var for offset
	offset := 8

	// grab the DSPIfo
	if dataHeader.DataMask&DSPIfoFlag == DSPIfoFlag {
		offset += 2
	}

	// grab the GestureInfo
	if dataHeader.DataMask&GestureInfoFlag == GestureInfoFlag {

		gestureInfo := &GestureInfo{}

		gopack.Unpack(buf[offset:offset+4], gestureInfo)

		//log.Printf("gesture %d", gestureInfo.Gesture&0xff)

		offset += 4

	}

	// grab the TouchInfo
	if dataHeader.DataMask&TouchInfoFlag == TouchInfoFlag {

		touchInfo := &TouchInfo{}

		gopack.Unpack(buf[offset:offset+4], touchInfo)

		//log.Printf("touchInfo %v", touchInfo)

		offset += 4
	}

	// grab the AirWheelInfo
	if dataHeader.DataMask&AirWheelInfoFlag == AirWheelInfoFlag {

		airWheelInfo := &AirWheelInfo{}

		gopack.Unpack(buf[offset:offset+2], airWheelInfo)

		log.Printf("airWheelInfo %v", airWheelInfo)

		offset += 2
	}

	// grab the CoordinateInfo
	if dataHeader.DataMask&CoordinateInfoFlag == CoordinateInfoFlag {

		coordinateInfo := &CoordinateInfo{}

		gopack.Unpack(buf[offset:offset+6], coordinateInfo)

		//log.Printf("coordinateInfo %v", coordinateInfo)

		offset += 6
	}

	if err != nil {
		log.Fatalf("Can't read %s - %s", GesticDevicePath, err)
	}

	if n > 0 {
		log.Printf("read %x", buf[:n])
	}
}

func FD_SET(p *syscall.FdSet, i int) {
	p.Bits[i/64] |= 1 << uint(i) % 64
}

func FD_ISSET(p *syscall.FdSet, i int) bool {
	return (p.Bits[i/64] & (1 << uint(i) % 64)) != 0
}

func FD_ZERO(p *syscall.FdSet) {
	for i := range p.Bits {
		p.Bits[i] = 0
	}
}
