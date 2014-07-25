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
		0x00, 0x00, 0x00, 0x00, // Unused
		0x00, 0x00, // Air wheel 
		0x00, 0x00, // x
		0x00, 0x00, // y
		0x00, 0x00 // z
	}

	rs.gestureVal = []byte{
		0x1a, 0x08, 0x12, 0x91, // header
		0x1f, 0x01, 0xaf, 0x8d, // data header
		0x00, 0x73, // DSPInfo
		0x00, 0x00, 0x00, 0x00, // GestureInfo
		0x02, 0x00, 0x00, 0x00, // Unused
		0x00, 0x00, // Air wheel
		0xaa, 0x53, // x
		0xac, 0x7c, // y
		0x00, 0x00 // z
	}

}

func (rs *ReaderSuite) TestReadGesture(c *C) {
	header := &Header{}

	gopack.Unpack(rs.blankValue[:4], header)

	log.Printf("T %+v", header)

	c.Assert(header.Id, Equals, uint8(0x91))

	dataHeader := &DateHeader{}

	gopack.Unpack(rs.blankValue[5:9], dataHeader)
	log.Printf("T %+v", dataHeader)

	c.Assert(dataHeader.DataMask, Equals, uint16(0xf100))

	c.Assert(dataHeader.DataMask&1, Equals, uint16(0))
	c.Assert(dataHeader.DataMask&2, Equals, uint16(0))
	c.Assert(dataHeader.DataMask&4, Equals, uint16(0))
	c.Assert(dataHeader.DataMask&8, Equals, uint16(0))

}

func (rs *ReaderSuite) TestReadGestureTwo(c *C) {
	header := &Header{}

	gopack.Unpack(rs.gestureVal[:4], header)

	log.Printf("T %+v", header)

	c.Assert(header.Id, Equals, uint8(0x91))

	dataHeader := &DateHeader{}

	gopack.Unpack(rs.gestureVal[5:9], dataHeader)
	log.Printf("T %+v", dataHeader)

	c.Assert(dataHeader.DataMask, Equals, uint16(0xf101))

	gopack.Unpack()

	c.Assert(dataHeader.DataMask&1, Equals, uint16(1))
	c.Assert(dataHeader.DataMask&2, Equals, uint16(0))
	c.Assert(dataHeader.DataMask&4, Equals, uint16(0))
	c.Assert(dataHeader.DataMask&8, Equals, uint16(0))
	c.Assert(dataHeader.DataMask&16, Equals, uint16(0))
}
