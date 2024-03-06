package bridge

import (
	"fmt"

	"network/data/configuration"
	"network/services/bridge/hooks"

	"github.com/charmbracelet/log"
	"github.com/hashicorp/mdns"
	"github.com/wind-c/comqtt/v2/mqtt"
	"github.com/wind-c/comqtt/v2/mqtt/listeners"
)

var Config *configuration.BridgeConfig
var MQTTServer *mqtt.Server
var MDNSServer *mdns.Server

func Initialize() error {
	log.Info("initializing bridge/v1")

	Config = &configuration.Config.Bridge

	MQTTServer = mqtt.New(&mqtt.Options{
		InlineClient: true,
	})

	tcp := listeners.NewTCP("t1", ":"+fmt.Sprint(Config.BrokerPort), nil)
	_ = MQTTServer.AddListener(tcp)

	if err := MQTTServer.AddHook(new(hooks.AuthenticationHook), &hooks.AuthenticationHookConfig{
		Server: MQTTServer,
	}); err != nil {
		return err
	}

	if err := MQTTServer.AddHook(new(hooks.InitializeDeviceHook), &hooks.InitializeDeviceHookConfig{
		Server: MQTTServer,
	}); err != nil {
		return err
	}

	if err := MQTTServer.AddHook(new(hooks.AuthorizateHook), &hooks.AuthorizateHookConfig{
		Server: MQTTServer,
	}); err != nil {
		return err
	}

	if err := InitializeMDNS(); err != nil {
		return err
	}

	_, err := InitializeLoader()
	if err != nil {
		return err
	}

	// Handle the cases when one or more plugins failed to load

	go func() {
		err := MQTTServer.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	return nil
}

func Cleanup() error {
	log.Info("cleaning up bridge/v1")

	MQTTServer.Close()
	MDNSServer.Shutdown()

	return nil
}
