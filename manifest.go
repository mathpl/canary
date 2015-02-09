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

	"github.com/mathpl/active_zabbix"
)

// Manifest represents configuration data.
type Manifest struct {
	Targets []Target
}

// GetManifest retreives a manifest from a given URL.
func GetManifest(url string) (*Manifest, error) {
	var manifest Manifest
	if strings.HasPrefix(url, "http") {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var jsonTargets []JsonTarget

		err = json.Unmarshal(body, &jsonTargets)
		if err != nil {
			return nil, err
		}

		for _, t := range jsonTargets {
			// Applying the current defaults for this manifest source
			t := Target{URL: t.URL, Name: t.Name, Key: t.Name, Type: "http_check",
				CheckInterval: time.Duration(10) * time.Second,
				Timeout:       time.Duration(10) * time.Second}
			manifest.Targets = append(manifest.Targets, t)
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

	return &manifest, nil
}
