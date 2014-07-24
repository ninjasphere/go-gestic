package agent

import (
	"log"
	"os"
	"syscall"
)

const GESTIC_DEV = "/dev/gestic"

type Reader struct {
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
