package canary

import "time"

type Config struct {
	ManifestURL           string
	DefaultSampleInterval time.Duration
	ReloadInterval        time.Duration
	RampupSensors         bool
	PublisherList         []string
	Ppid                  int
}
