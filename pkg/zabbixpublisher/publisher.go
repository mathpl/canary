package zabbixpublisher

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mathpl/active_zabbix"
	"github.com/mathpl/canary"
)

// Publisher implements canary.Publisher, and is our
// gateway for delivering canary.Measurement data to STDOUT.
type Publisher struct {
	zc   active_zabbix.ZabbixActiveClient
	host string
}

// New returns a pointer to a new Publsher.
func New(addr string, host string) (p *Publisher, err error) {
	p = &Publisher{}
	p.zc, err = active_zabbix.NewZabbixActiveClient(addr, 5000, 5000)
	p.host = host
	return p, err
}

func NewFromEnv() (*Publisher, error) {
	push_addr := os.Getenv("ZABBIX_PUSH_ADDR")
	if push_addr == "" {
		return nil, fmt.Errorf("ZABBIX_PUSH_ADDR not set in ENV")
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return New(push_addr, hostname)
}

// Publish takes a canary.Measurement and emits data to STDOUT.
func (p *Publisher) Publish(m canary.Measurement) (err error) {
	if m.Error != nil {
		m.Sample.StatusCode = -1
	}

	zm := active_zabbix.ZabbixMetricKeyJson{Host: p.host, Key: m.Target.Key,
		Value: fmt.Sprintf("%d", m.Sample.StatusCode), Clock: fmt.Sprintf("%d", m.Sample.T2.Unix())}

	data := make([]active_zabbix.ZabbixMetricKeyJson, 1)
	data[0] = zm

	zr := active_zabbix.ZabbixMetricRequestJson{Request: "agent data", Data: data}

	var marshalledJson []byte
	marshalledJson, err = json.Marshal(zr)

	return p.zc.ZabbixSendAndForget(marshalledJson)
}
