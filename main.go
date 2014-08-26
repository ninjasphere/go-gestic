package main

import (
	"fmt"
	"os"
	"os/signal"

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

	log.Infof("resetting gestic device")

	err = writetofile("/sys/kernel/debug/omap_mux/mii1_rxdv", "2f")

	if err != nil {
		log.Errorf("Unable to reset gestic device: %v", err)
		return 1
	}

	err = writetofile("/sys/class/gpio/export", "100")

	if err != nil {
		log.Warningf("Unable to write to export pin: %v", err)
	}

	err = writetofile("/sys/class/gpio/gpio100/direction", "out")

	if err != nil {
		log.Errorf("Unable to reset gestic device: %v", err)
		return 1
	}

	err = writetofile("/sys/class/gpio/gpio100/value", "0")

	if err != nil {
		log.Errorf("Unable to reset gestic device: %v", err)
		return 1
	}

	err = writetofile("/sys/class/gpio/gpio100/value", "1")

	if err != nil {
		log.Errorf("Unable to reset gestic device: %v", err)
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

	reader := gestic.NewReader(conn, log)
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
