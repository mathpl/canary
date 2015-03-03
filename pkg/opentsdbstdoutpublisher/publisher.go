package opentsdbstdoutpublisher

import (
	"fmt"

	"github.com/mathpl/canary/pkg/sensor"
)

// Publisher implements canary.Publisher, and is our
// gateway for delivering canary.Measurement data to STDOUT.
type Publisher struct{}

// New returns a pointer to a new Publsher.
func New() *Publisher {
	return &Publisher{}
}

// Publish takes a canary.Measurement and emits data to STDOUT.
func (p *Publisher) Publish(m sensor.Measurement) (err error) {
	duration := m.Sample.T2.Sub(m.Sample.T1).Seconds() * 1000

	if m.Error != nil {
		m.Sample.StatusCode = -1
	}

	fmt.Printf(
		"%s %d %f status=%d check=%s\n",
		m.Target.Type,
		m.Sample.T2.Unix(),
		duration,
		m.Sample.StatusCode,
		m.Target.Name,
	)
	return
}
