package zabbixstdoutpublisher

import (
	"fmt"

	"github.com/mathpl/canary/pkg/sampler"
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
	switch e := m.Error.(type) {
	case *sampler.StatusCodeError:
		m.Sample.StatusCode = e.StatusCode
	default:
		fmt.Printf("%+V\n", e)
		if m.Error != nil {
			m.Sample.StatusCode = 0
		}
	}

	fmt.Printf(
		"%s = %d\n",
		m.Target.Key,
		m.Sample.StatusCode,
	)
	return
}
