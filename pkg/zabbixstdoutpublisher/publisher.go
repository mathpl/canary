package zabbixstdoutpublisher

import (
	"fmt"

	"github.com/mathpl/canary"
)

// Publisher implements canary.Publisher, and is our
// gateway for delivering canary.Measurement data to STDOUT.
type Publisher struct{}

// New returns a pointer to a new Publsher.
func New() *Publisher {
	return &Publisher{}
}

// Publish takes a canary.Measurement and emits data to STDOUT.
func (p *Publisher) Publish(m canary.Measurement) (err error) {
	if m.Error != nil {
		m.Sample.StatusCode = -1
	}

	fmt.Printf(
		"%s = %d\n",
		m.Target.Key,
		m.Sample.StatusCode,
	)
	return
}
