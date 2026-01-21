// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 Robin Jarry

package coverage

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"

	"github.com/openstack-k8s-operators/openstack-network-exporter/appctl"
	"github.com/openstack-k8s-operators/openstack-network-exporter/collectors/lib"
	"github.com/openstack-k8s-operators/openstack-network-exporter/config"
	"github.com/openstack-k8s-operators/openstack-network-exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func makeMetric(m lib.Metric, val float64) prometheus.Metric {
	if !config.MetricSets().Has(m.Set) {
		return nil
	}
	return prometheus.MustNewConstMetric(m.Desc(), m.ValueType, val)
}

type Collector struct{}

func (Collector) Name() string {
	return "coverage"
}

func (Collector) Metrics() []lib.Metric {
	var res []lib.Metric
	for _, m := range metrics {
		res = append(res, m)
	}
	return res
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	lib.DescribeEnabledMetrics(c, ch)
}

// "netdev_sent       967178.4/sec 966510.667/sec   880482.1181/sec   total: 21235468562413"
var coverageRe = regexp.MustCompile(`^(\w+)\s+.*\s+total: (\d+)$`)

func (Collector) Collect(ch chan<- prometheus.Metric) {
	buf := appctl.OvsVSwitchd("coverage/show")

	// Parse coverage/show output into a map of name -> value
	// OVS only reports non-zero counters, so we need to emit 0 for missing ones
	values := make(map[string]float64)
	scanner := bufio.NewScanner(strings.NewReader(buf))
	for scanner.Scan() {
		line := scanner.Text()
		match := coverageRe.FindStringSubmatch(line)
		if match != nil {
			name := match[1]
			val, err := strconv.ParseFloat(match[2], 64)
			if err != nil {
				log.Errf("%s: %s: %s", name, match[2], err)
				continue
			}
			values[name] = val
		}
	}

	// Emit all defined metrics, using 0 for any not present in output
	for name, m := range metrics {
		val := values[name]
		metric := makeMetric(m, val)
		if metric != nil {
			ch <- metric
		}
	}
}
