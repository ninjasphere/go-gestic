package agent

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/juju/loggo"
	"github.com/mitchellh/cli"
	"github.com/ninjasphere/go-ninja"
)

type Command struct {
	Ui         cli.Ui
	ShutdownCh <-chan struct{}
	args       []string
	logger     loggo.Logger
	debug      bool
	agent      *Agent
}

func (c *Command) handleSignals(config *Config) int {
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	var sig os.Signal
	select {
	case s := <-signalCh:
		sig = s
	case <-c.ShutdownCh:
		sig = os.Interrupt
	}
	c.Ui.Output(fmt.Sprintf("Caught signal: %v", sig))

	return 0
}

func (c *Command) readConfig(args []string) *Config {
	var cmdConfig Config
	cmdFlags := flag.NewFlagSet("agent", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	cmdFlags.BoolVar(&c.debug, "debug", false, "Enable debug level logging for the agent.")
	cmdFlags.StringVar(&cmdConfig.LocalUrl, "localurl", "tcp://localhost:1883", "cloud url to connect to")
	cmdFlags.StringVar(&cmdConfig.SerialNo, "serial", "unknown", "the serial number of the device")

	if err := cmdFlags.Parse(c.args); err != nil {
		return nil
	}

	if c.debug {
		// set the root logger to debug
		loggo.GetLogger("").SetLogLevel(loggo.DEBUG)
	} else {
		loggo.GetLogger("").SetLogLevel(loggo.INFO)
	}

	return &cmdConfig
}

func (c *Command) Run(args []string) int {

	var err error

	c.args = args
	config := c.readConfig(args)

	if config == nil {
		return 1
	}

	// configure the agent logger
	c.logger = loggo.GetLogger("agent")
	if c.agent, err = CreateAgent(config); err != nil {
		c.logger.Errorf("Unable to init agent : %v", err)
		return 1
	}

	// main logic here
	conn, err := ninja.Connect("com.ninjablocks.gestic")

	if err != nil {
		c.logger.Errorf("Connect failed: %v", err)
		return 1
	}

	pwd, err := os.Getwd()

	if err != nil {
		c.logger.Errorf("Connect failed: %v", err)
		return 1
	}

	_, err = conn.AnnounceDriver("com.ninjablocks.gestic", "driver-gestic", pwd)
	if err != nil {
		c.logger.Errorf("Could not get driver bus: %v", err)
		return 1
	}

	reader := NewReader(conn)

	go reader.Start()

	return c.handleSignals(config)
}

func (c *Command) Synopsis() string {
	return "Runs a appname agent"
}

func (c *Command) Help() string {
	helpText := `
Usage: driver-go-gestic agent [options]

  Starts the appname agent and runs until an interrupt is received.

Options:

  -debug              Enable debug level logging for the agent.
  -localurl=tcp://localhost:1883      URL for the local broker.
  -serial=123123                      Configure the Serial number of the device.
`
	return strings.TrimSpace(helpText)
}
