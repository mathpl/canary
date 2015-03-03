package canary

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mathpl/canary/pkg/sampler"
)

// Manifest represents configuration data.
type Manifest struct {
	Targets []sampler.Target
	StartDelays []float64
}

// GenerateRampupDelays generates an even distribution of sensor start delays
// based on the passed number of interval seconds and the number of targets.
func (m *Manifest) GenerateRampupDelays(intervalSeconds int) {
	var intervalMilliseconds = float64(intervalSeconds*1000)
	var chunkSize = float64(intervalMilliseconds/float64(len(m.Targets)))
	for i := 0.0; i < intervalMilliseconds; i = i + chunkSize {
		m.StartDelays[int((i/chunkSize))] = i
	}
}

// GetManifest retreives a manifest from a given URL.
func GetManifest(url string) (manifest Manifest, err error) {
	if strings.HasPrefix(url, "http") {
		resp, err := http.Get(url)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}

		err = json.Unmarshal(body, &manifest)
		if err != nil {
			return
		}
	} else if strings.HasPrefix(url, "zbx") {
		zc, err := active_zabbix.NewZabbixActiveClient(url, 5000, 5000)
		if err != nil {
			return nil, err
		}

		// Fetch active check for current hostname
		host, err := os.Hostname()
		if err != nil {
			return nil, err
		}

		host_keys, err := zc.FetchActiveChecks(host)
		if err != nil {
			return nil, err
		}

		// Get the regexp to extract data from zabbix key ready
		http_regexp := regexp.MustCompile("healthcheck\\[(http://[^\\]]+)\\]\\[(\\d+)\\]\\[(.*)\\]")

		targets := make([]Target, 0)
		for host_key, check_interval := range host_keys {
			matches := http_regexp.FindAllStringSubmatch(host_key, -1)
			if len(matches) > 0 {
				// Extract url from zabbix key. Format
				//   http.healthcheck[<http://url/>][<timeout in ms>][<healthcheck name>]
				// Check interval is provided by zabbix
				timeout, err := strconv.Atoi(matches[0][2])
				if err != nil {
					return nil, err
				}

				timeout_dur := time.Duration(timeout) * time.Millisecond
				t := Target{URL: matches[0][1], Name: matches[0][3],
					Key: host_key, Type: "http.healthcheck",
					CheckInterval: check_interval,
					Timeout:       timeout_dur}
				targets = append(targets, t)
			}
		}
		manifest.Targets = targets
	}

	// Initialize manifest.StartDelays to zeros
	manifest.StartDelays = make([]float64, len(manifest.Targets))
	for i := 0; i < len(manifest.Targets); i++ {
		manifest.StartDelays[i] = 0.0
	}

	return
}
