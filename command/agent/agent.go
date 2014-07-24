package agent

import (
	"net/url"

	"github.com/juju/loggo"
	"github.com/wolfeidau/datbus"
)

type Agent struct {
	bus    *datbus.Bus
	logger loggo.Logger
	config *Config
}

func CreateAgent(config *Config) (*Agent, error) {

	logger := loggo.GetLogger("agent")

	localUrl, err := url.Parse(config.LocalUrl)

	if err != nil {
		return nil, err
	}

	busConfig := &datbus.Configuration{
		MqttUrl:  localUrl,
		ClientId: "gestic-go-driver",
	}

	bus, err := datbus.NewBus(busConfig)

	if err != nil {
		return nil, err
	}

	return &Agent{bus: bus, logger: logger, config: config}, nil
}
