import serial
import os, select, struct
from collections import namedtuple

Header = namedtuple('Header', ['length','flags','seq','msgid'])

os.system( 'echo 2f > /sys/kernel/debug/omap_mux/mii1_rxdv' )
os.system( 'echo 2f > /sys/kernel/debug/omap_mux/mii1_rxclk' )

ser = serial.Serial('/dev/ttyGestureCalibrate0', 9600, timeout=0.1)

buf = ''

fd = os.open('/dev/gestic', os.O_RDWR)

while True:
	ret = select.select([fd, ser.fileno()],[fd],[fd], 0.1)
	if fd in ret[0]:
		data = os.read( fd, 255 )
		print 'G>A', repr(data)
		if len(data) > 0:
			ser.write( '\xfe\xff' + data )
	if ser.fileno() in ret[0]:
		buf += ser.read()

	if buf.count('\xfe\xff') > 1:
		_,chunk,rest = buf.split('\xfe\xff', 2)
		header = Header( *struct.unpack( 'BBBB', chunk[:4] ) )
	 	
		buf = '\xfe\xff' + rest

		print 'G<A', repr(chunk)
		os.write( fd, chunk )


os.close(fd)

ser.close()