package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/ninjasphere/driver-go-gestic/gestic"
	"github.com/ninjasphere/go-ninja"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {

	// configure the agent logger
	logger := ninja.GetLogger("agent")

	// main logic here
	conn, err := ninja.Connect("com.ninjablocks.gestic")

	if err != nil {
		logger.Errorf("Connect failed: %v", err)
		return 1
	}

	pwd, err := os.Getwd()

	if err != nil {
		logger.Errorf("Connect failed: %v", err)
		return 1
	}

	_, err = conn.AnnounceDriver("com.ninjablocks.gestic", "driver-gestic", pwd)
	if err != nil {
		logger.Errorf("Could not get driver bus: %v", err)
		return 1
	}

	reader := gestic.NewReader(conn, logger)
	go reader.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// Block until a signal is received.
	s := <-c
	fmt.Println("Got signal:", s)

	return 0
}
