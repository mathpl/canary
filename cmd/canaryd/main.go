package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mathpl/canary"
	"github.com/mathpl/canary/pkg/libratopublisher"
	"github.com/mathpl/canary/pkg/opentsdbstdoutpublisher"
	"github.com/mathpl/canary/pkg/stdoutpublisher"
	"github.com/mathpl/canary/pkg/transportsampler"
	"github.com/mathpl/canary/pkg/zabbixstdoutpublisher"
)

type config struct {
	ManifestURL   string
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

	// output chan
	c := make(chan canary.Measurement)

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
		default:
			log.Printf("Unknown publisher: %s", publisher)
		}
	}

	// spinup a scheduler for each target
	for _, target := range manifest.Targets {
		scheduler := canary.Scheduler{
			Target:  target,
			C:       c,
			Sampler: transportsampler.New(target.Timeout),
		}
		go scheduler.Start(target.CheckInterval)
	}

	// publish each incoming measurement
	for m := range c {
		for _, p := range publishers {
			p.Publish(m)
		}
	}
}
