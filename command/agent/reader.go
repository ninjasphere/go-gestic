package agent

import (
	"log"
	"os"
	"syscall"

	"github.com/joshlf13/gopack"
)

// Gestic device path
const GESTIC_DEV = "/dev/gestic"

// Flag which indicates if the payload contains data
const ID_SENSOR_DATA_OUTPUT = 0x91

type Reader struct {
}

type Header struct {
	Length, Flags, Seq, Id uint8
}

type DateHeader struct {
	DataMask              uint16
	TimeStamp, SystemInfo uint8
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

type GestureInfo struct {
	Gesture uint32
}

func (gi *GestureInfo) GetGestureName() string {
	gest := gi.Gesture & 0xff
	return Gestures[int(gest)]
}

func (r *Reader) Start() {
	log.Printf("Opening %s", GESTIC_DEV)

	fd, err := syscall.Open(GESTIC_DEV, os.O_RDWR, 0666)

	rfds := &syscall.FdSet{}
	timeout := &syscall.Timeval{}

	if err != nil {
		log.Fatalf("Can't open %s - %s", GESTIC_DEV, err)
	}

	log.Printf("Reading %s", GESTIC_DEV)

	//	ping_at := time.Now()

	for {

		FD_ZERO(rfds)
		FD_SET(rfds, fd)

		_, err := syscall.Select(fd+1, rfds, nil, nil, timeout)

		if err != nil {
			log.Fatalf("Can't read %s - %s", GESTIC_DEV, err)
		}
		//  One of the fds changed
		if FD_ISSET(rfds, int(fd)) {

			buf := make([]byte, 255)
			n, err := syscall.Read(fd, buf)

			header := &Header{}
			dataHeader := &DateHeader{}
			//gestureInfo := &GestureInfo{}

			gopack.Unpack(buf[:4], header)
			payload := buf[4:]

			log.Printf("header %v", header)

			if header.Id == ID_SENSOR_DATA_OUTPUT {
				gopack.Unpack(payload[:4], dataHeader)

				log.Printf("dataOutputConfig 1 %d", dataHeader.DataMask)

				payload = payload[4:]

				log.Printf("dataHeader %v", dataHeader)

				// if dataHeader.DataOutputConfigMask&1 != 0 {
				// 	//contains DSPInfo, ignore
				// 	rest = rest[2:]
				// }
				//
				// if dataHeader.DataOutputConfigMask&2 != 0 {
				// 	// contains GestureInfo
				// 	//(GestureInfo,) = struct.unpack( 'I', rest[:4])
				// 	gopack.Unpack(rest[:4], gestureInfo)
				//
				// 	log.Printf("gesture %s", gestureInfo.GetGestureName())
				// 	rest = rest[2:]
				// }
			}

			// if the msgid == CONST  we have data

			// 		data = SensorDataOutputHeader( *struct.unpack( 'HBB', payload[:4] ) )

			//	unsigned short and two bytes
			// namedtuple('SensorDataOutput', ['DataOutputConfigMask', 'TimeStamp', 'SystemInfo'])

			if err != nil {
				log.Fatalf("Can't read %s - %s", GESTIC_DEV, err)
			}

			if n > 0 {
				log.Printf("read %x", buf[:n])
			}
		}

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
