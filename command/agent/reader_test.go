package agent

// 1a081a911f00f180007300000000000000000000000000000000

// 1a0812911f01af8d007300000000020000000000aa53ac7c0000

import (
	"testing"

	"log"

	"github.com/joshlf13/gopack"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type ReaderSuite struct {
	blankValue []byte
	gestureVal []byte
	reader     *Reader
}

var _ = Suite(&ReaderSuite{})

func (rs *ReaderSuite) SetUpTest(c *C) {
	rs.blankValue = []byte{
		0x1a, 0x08, 0x1a, 0x91, // header
		0x1f, 0x00, 0xf1, 0x80, // data header
		0x00, 0x73, // DSPInfo
		0x00, 0x00, 0x00, 0x00, // GestureInfo
		0x00, 0x00, 0x00, 0x00, // TouchInfo
		0x00, 0x00, // Air wheel
		0x00, 0x00, // x
		0x00, 0x00, // y
		0x00, 0x00, // z
	}

	rs.gestureVal = []byte{
		0x1a, 0x08, 0x12, 0x91, // header
		0x1f, 0x01, 0xaf, 0x8d, // data header
		0x00, 0x73, // DSPInfo
		0x00, 0x00, 0x00, 0x00, // GestureInfo
		0x02, 0x00, 0x00, 0x00, // TouchInfo
		0x00, 0x00, // Air wheel
		0xaa, 0x53, // x
		0xac, 0x7c, // y
		0x00, 0x00, // z
	}

}

func (rs *ReaderSuite) TestReadGesture(c *C) {
	header := &Header{}

	gopack.Unpack(rs.blankValue[:4], header)

	log.Printf("T %+v", header)

	c.Assert(header.Id, Equals, uint8(0x91))

	dataHeader := &DateHeader{}

	gopack.Unpack(rs.blankValue[4:8], dataHeader)
	log.Printf("T %+v", dataHeader)

	c.Assert(dataHeader.DataMask, Equals, uint16(0x1f))

	c.Assert(dataHeader.DataMask&1, Equals, uint16(1))
	c.Assert(dataHeader.DataMask&2, Equals, uint16(2))
	c.Assert(dataHeader.DataMask&4, Equals, uint16(4))
	c.Assert(dataHeader.DataMask&8, Equals, uint16(8))
	c.Assert(dataHeader.DataMask&16, Equals, uint16(16))

	log.Printf("len %d", len(rs.blankValue))

}

func (rs *ReaderSuite) TestReadGestureTwo(c *C) {
	header := &Header{}

	gopack.Unpack(rs.gestureVal[:4], header)

	log.Printf("header %+v", header)

	c.Assert(header.Id, Equals, uint8(0x91))

	dataHeader := &DateHeader{}

	gopack.Unpack(rs.gestureVal[4:8], dataHeader)
	log.Printf("dataHeader %+v", dataHeader)

	c.Assert(dataHeader.DataMask, Equals, uint16(0x11f))

	c.Assert(dataHeader.DataMask&1, Equals, uint16(1))

	// var for offset
	offset := 8

	// grab the DSPIfo
	if dataHeader.DataMask&BIT_ONE == BIT_ONE {

		dspinfo := &DSPInfo{}

		gopack.Unpack(rs.gestureVal[offset:offset+2], dspinfo)

		log.Printf("dspinfo %+v", dspinfo)

		offset += 2
	}

	// grab the GestureInfo
	if dataHeader.DataMask&BIT_TWO == BIT_TWO {

		gestureInfo := &GestureInfo{}

		gopack.Unpack(rs.gestureVal[offset:offset+4], gestureInfo)

		log.Printf("gesture %d", gestureInfo.Gesture&0xff)

		offset += 4

	}

	// SKIP 4 bytes

	// grab the TouchInfo
	if dataHeader.DataMask&BIT_FOUR == BIT_FOUR {

		touchInfo := &TouchInfo{}

		gopack.Unpack(rs.gestureVal[offset:offset+4], touchInfo)

		log.Printf("touchInfo %v", touchInfo)

		offset += 4
	}

	// grab the AirWheelInfo
	if dataHeader.DataMask&BIT_EIGHT == BIT_EIGHT {

		airWheelInfo := &AirWheelInfo{}

		gopack.Unpack(rs.gestureVal[offset:offset+2], airWheelInfo)

		log.Printf("airWheelInfo %v", airWheelInfo)

		offset += 2
	}

	// grab the CoordinateInfo
	if dataHeader.DataMask&BIT_SIXTEEN == BIT_SIXTEEN {

		coordinateInfo := &CoordinateInfo{}

		gopack.Unpack(rs.gestureVal[offset:offset+6], coordinateInfo)

		log.Printf("coordinateInfo %v", coordinateInfo)

		offset += 6
	}

	c.Assert(dataHeader.DataMask&2, Equals, uint16(2))
	c.Assert(dataHeader.DataMask&4, Equals, uint16(4))
	c.Assert(dataHeader.DataMask&8, Equals, uint16(8))
	c.Assert(dataHeader.DataMask&16, Equals, uint16(16))
}
