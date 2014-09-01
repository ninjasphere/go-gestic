package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bitly/go-simplejson"
	"github.com/ninjasphere/driver-go-gestic/gestic"
	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-ninja/logger"
)

const driverName = "driver-gestic"

func main() {
	os.Exit(realMain())
}

func realMain() int {

	debug("DEBUG BUILD")

	// configure the agent logger
	log := logger.GetLogger(driverName)

	// main logic here
	conn, err := ninja.Connect("com.ninjablocks.gestic")

	if err != nil {
		log.Errorf("Connect failed: %v", err)
		return 1
	}

	pwd, err := os.Getwd()

	if err != nil {
		log.Errorf("Connect failed: %v", err)
		return 1
	}

	if err := gestic.ResetDevice(); err != nil {
		log.Errorf(err)
		return 1
	}

	_, err = conn.AnnounceDriver("com.ninjablocks.gestic", driverName, pwd)
	if err != nil {
		log.Errorf("Could not get driver bus: %v", err)
		return 1
	}

	statusJob, err := ninja.CreateStatusJob(conn, driverName)

	if err != nil {
		log.FatalErrorf(err, "Could not setup status job")
	}

	statusJob.Start()

	reader := gestic.NewReader(log, func(g *gestic.GesticData) {

		if g.Gesture.GestureVal > 0 {
			jsonmsg, _ := simplejson.NewJson([]byte(`{}`))
			jsonmsg.Set("gesture", g.Gesture.Name())
			jsonmsg.Set("seq", g.Event.Seq)
			r.conn.PublishRPCMessage("$client/gesture/gesture", jsonmsg)
		}

		if g.Touch.TouchVal > 0 {
			jsonmsg, _ := simplejson.NewJson([]byte(`{}`))
			jsonmsg.Set("touch", g.Touch.Name())
			jsonmsg.Set("seq", g.Event.Seq)
			r.conn.PublishRPCMessage("$client/gesture/touch", jsonmsg)
		}

		if g.AirWheel.AirWheelVal > 0 {
			jsonmsg, _ := simplejson.NewJson([]byte(`{}`))
			jsonmsg.Set("airwheel", g.AirWheel.AirWheelVal)
			jsonmsg.Set("seq", g.Event.Seq)
			r.conn.PublishRPCMessage("$client/gesture/airwheel", jsonmsg)
		}

		if g.Coordinates.X != 0 || g.Coordinates.Y != 0 || g.Coordinates.Z != 0 {
			jsonmsg, _ := simplejson.NewJson([]byte(`{}`))
			jsonmsg.Set("x", g.Coordinates.X)
			jsonmsg.Set("y", g.Coordinates.Y)
			jsonmsg.Set("z", g.Coordinates.Z)
			jsonmsg.Set("seq", g.Event.Seq)
			r.conn.PublishRPCMessage("$client/gesture/position", jsonmsg)
		}
	})

	go reader.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// Block until a signal is received.
	s := <-c
	fmt.Println("Got signal:", s)

	return 0
}

func writetofile(fn string, val string) error {

	df, err := os.OpenFile(fn, os.O_WRONLY|os.O_SYNC, 0666)

	if err != nil {
		return err
	}

	defer df.Close()

	if _, err = fmt.Fprintln(df, val); err != nil {
		return err
	}

	return nil
}
