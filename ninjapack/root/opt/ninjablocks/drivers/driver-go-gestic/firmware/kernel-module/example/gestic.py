import os
import select, sys
import struct, json
from collections import namedtuple

Header = namedtuple('Header', ['length','flags','seq','msgid'])

SensorDataOutputHeader = namedtuple('SensorDataOutput', ['DataOutputConfigMask', 'TimeStamp', 'SystemInfo'])

ID_SENSOR_DATA_OUTPUT = 0x91

GESTURES = [
	'No gesture',
	'Garbage model',
	'Flick West to East',
	'Flick East to West',
	'Flick South to North',
	'Flick North to South',
	'Circle clockwise',
	'Circle counter-clockwise',
]

def process_message( buffer ):
	if len(buffer) < 4:
		return # not enough data to get whole header

	# length - byte (not including magic)
	# flags - byte
	# seq - byte
	# id - byte
	header = Header( *struct.unpack( 'BBBB', buffer[:4] ) )

	if len(buffer) < header.length:
		print 'WARN: Truncated data'
		return # not enoguh data to get whole message

	message = buffer[:header.length]
	payload = message[4:]

	if header.msgid == ID_SENSOR_DATA_OUTPUT:
		data = SensorDataOutputHeader( *struct.unpack( 'HBB', payload[:4] ) )
		rest = payload[4:]

		if data.DataOutputConfigMask & 1:
			# contains DSPInfo, ignore
			rest = rest[2:]
		if data.DataOutputConfigMask & 2:
			# contains GestureInfo
			(GestureInfo,) = struct.unpack( 'I', rest[:4])
			Gesture = GestureInfo & 0xff
			rest = rest[4:]

			print json.dumps({'gesture': GESTURES[Gesture]})
			sys.stdout.flush()
		if data.DataOutputConfigMask & 4:
			rest = rest[4:]
		if data.DataOutputConfigMask & 8:
			(AirWheelInfo,crap) = struct.unpack( 'BB', rest[:2] )
			rest = rest[2:]

			print json.dumps({'airwheel': AirWheelInfo})
			sys.stdout.flush()
		if data.DataOutputConfigMask & 16:
			(xPos, yPos, zPos) = struct.unpack( 'HHH', rest[:6] )
			rest = rest[6:]

			print json.dumps({'position': {'x': xPos, 'y': yPos, 'z': zPos}})
			sys.stdout.flush()
###

# currently the kernel driver itself doesn't handle pin muxing
os.system( 'echo 2f > /sys/kernel/debug/omap_mux/mii1_rxdv' ) # mii1_col on EVK

f = os.open( '/dev/gestic', os.O_RDWR )

while True:
	ret = select.select([f],[f],[f], 1)
	if len(ret[0]) > 0:
		print 'Got data ready to read!'
		data = os.read( f, 255 )
		#header = Header( *struct.unpack( 'BBBB', data[:4] ) )
		#print header
		process_message( data )

os.close( f )