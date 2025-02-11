package main

import (
	"os"
	"os/signal"

	"git.terah.dev/imterah/goevdi/libdisplayconfig"
	lib "git.terah.dev/imterah/goevdi/libevdi"
	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v3"
)

type Displays struct {
	DisplayList []struct {
		Modes []libdisplayconfig.Mode
	} `yaml:"displays"`
}

func main() {
	log.Info("OpenVD - based on GoEvdi")

	logLevel := os.Getenv("OPENVD_LOG_LEVEL")

	if logLevel != "" {
		switch logLevel {
		case "debug":
			log.SetLevel(log.DebugLevel)

		case "info":
			log.SetLevel(log.InfoLevel)

		case "warn":
			log.SetLevel(log.WarnLevel)

		case "error":
			log.SetLevel(log.ErrorLevel)

		case "fatal":
			log.SetLevel(log.FatalLevel)
		}
	}

	lib.SetupLogger(&lib.EvdiLogger{
		Log: func(message string) {
			log.Debugf("evdi: %s", message)
		},
	})

	if len(os.Args) <= 1 {
		log.Fatalf("Illegal arguments! Usage: openvd [configuration_file]")
	}

	config, err := os.ReadFile(os.Args[1])

	if err != nil {
		log.Fatalf("Failed to read configuration file: %s", err.Error())
	}

	displays := &Displays{}
	err = yaml.Unmarshal(config, displays)

	if err != nil {
		log.Fatalf("Failed to parse configuration file: %s", err.Error())
	}

	evdiInstances := []*lib.EvdiNode{}

	for displayID, display := range displays.DisplayList {
		log.Infof("Setting up display #%d", displayID+1)

		if len(display.Modes) == 0 {
			log.Fatalf("Failed to set up display: no modes specified")
		}

		edid, err := libdisplayconfig.GenerateEDID(display.Modes)

		if err != nil {
			log.Fatalf("Failed to generate EDID: %s", err.Error())
		}

		log.Debug("Attempting to add EVDI device...")
		evdi, err := lib.Open(nil)

		if err != nil {
			log.Fatalf("Failed to add EVDI device: %s", err.Error())
		}

		evdiInstances = append(evdiInstances, evdi)

		log.Debug("Attempting to configure display...")
		evdi.Connect(edid, uint(display.Modes[0].Width), uint(display.Modes[0].Height), uint(display.Modes[0].Refresh))

		log.Infof("Succesfully set up display #%d", displayID+1)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for range c {
		log.Info("Quitting.")

		for _, evdi := range evdiInstances {
			evdi.Disconnect()
			evdi.Close()
		}

		os.Exit(0)
	}
}
