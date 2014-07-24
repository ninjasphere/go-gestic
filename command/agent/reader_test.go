package agent

// 1a081a911f00f180007300000000000000000000000000000000

// 1a0812911f01af8d007300000000020000000000aa53ac7c0000

import (
	"testing"

	"log"

	"github.com/joshlf13/gopack"
	. "launchpad.net/gocheck"
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
		0x1a, 0x08, 0x1a, 0x91,
		0x1f, 0x00, 0xf1, 0x80,
		0x00, 0x73, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	rs.gestureVal = []byte{
		0x1a, 0x08, 0x12, 0x91,
		0x1f, 0x01, 0xaf, 0x8d,
		0x00, 0x73, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0xaa, 0x53, 0xac, 0x7c, 0x00, 0x00}

}

func (rs *ReaderSuite) TestReadGesture(c *C) {
	header := &Header{}

	gopack.Unpack(rs.blankValue[:4], header)

	log.Printf("T %+v", header)

}
