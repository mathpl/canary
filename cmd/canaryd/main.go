package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"strconv"

	"github.com/mathpl/canary"
	"github.com/mathpl/canary/pkg/libratopublisher"
	"github.com/mathpl/canary/pkg/opentsdbstdoutpublisher"
	"github.com/mathpl/canary/pkg/sampler"
	"github.com/mathpl/canary/pkg/sensor"
	"github.com/mathpl/canary/pkg/stdoutpublisher"
	"github.com/mathpl/canary/pkg/transportsampler"
	"github.com/mathpl/canary/pkg/zabbixpublisher"
	"github.com/mathpl/canary/pkg/zabbixstdoutpublisher"
)

type config struct {
	ManifestURL   string
	DefaultSampleInterval int
	RampupSensors bool
	PublisherList []string
}

// builds the app configuration via ENV
func getConfig() (c config, err error) {
	c.ManifestURL = os.Getenv("MANIFEST_URL")
	if c.ManifestURL == "" {
		err = fmt.Errorf("MANIFEST_URL not defined in ENV")
	}

	list := os.Getenv("PUBLISHERS")
	if list == "" {
		list = "stdout"
	}
	c.PublisherList = strings.Split(list, ",")

	interval := os.Getenv("DEFAULT_SAMPLE_INTERVAL")
	// if the variable is unset, an empty string will be returned
	if interval == "" {
		interval = "1"
	}
	c.DefaultSampleInterval, err = strconv.Atoi(interval)
	if err != nil {
		err = fmt.Errorf("DEFAULT_SAMPLE_INTERVAL is not a valid integer")
	}

	// Set RampupSensors if RAMPUP_SENSORS is set to 'yes'
	rampUp := os.Getenv("RAMPUP_SENSORS")
	if rampUp == "yes" {
		c.RampupSensors = true
	} else {
		c.RampupSensors = false
	}

	return
}

func main() {
	conf, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	manifest, err := canary.GetManifest(conf.ManifestURL)
	if err != nil {
		log.Fatal(err)
	}

	if conf.RampupSensors {
		manifest.GenerateRampupDelays(conf.DefaultSampleInterval)
	}

	// output chan
	c := make(chan sensor.Measurement)

	var publishers []canary.Publisher

	// spinup publishers
	for _, publisher := range conf.PublisherList {
		switch publisher {
		case "stdout":
			p := stdoutpublisher.New()
			publishers = append(publishers, p)
		case "librato":
			p, err := libratopublisher.NewFromEnv()
			if err != nil {
				log.Fatal(err)
			}
			publishers = append(publishers, p)
		case "opentsdbstdout":
			p := opentsdbstdoutpublisher.New()
			if err != nil {
				log.Fatal(err)
			}
			publishers = append(publishers, p)
		case "zabbixstdout":
			p := zabbixstdoutpublisher.New()
			if err != nil {
				log.Fatal(err)
			}
			publishers = append(publishers, p)
		case "zabbix":
			p, err := zabbixpublisher.NewFromEnv()
			if err != nil {
				log.Fatal(err)
			}
			publishers = append(publishers, p)
		default:
			log.Printf("Unknown publisher: %s", publisher)
		}
	}

	// spinup a sensor for each target
	for index, target := range manifest.Targets {
		// Determine whether to use target.Interval or conf.DefaultSampleInterval
		var interval int;
		// Targets that lack an interval value in JSON will have their value set to zero. in this case,
		// use the DefaultSampleInterval
		if target.Interval == 0 {
			interval = conf.DefaultSampleInterval
		} else {
			interval = target.Interval
		}
		sensor := sensor.Sensor{
			Target:  target,
			C:       c,
			Sampler: sampler.New(),
		}
		go sensor.Start(interval, manifest.StartDelays[index])
	}

	// publish each incoming measurement
	for m := range c {
		for _, p := range publishers {
			p.Publish(m)
		}
	}
}
