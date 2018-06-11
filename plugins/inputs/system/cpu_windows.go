// +build windows

package system

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/shirou/gopsutil/cpu"
)

type CPUStats struct {
	ps        PS

	PerCPU         bool `toml:"percpu"`
	TotalCPU       bool `toml:"totalcpu"`
	CollectCPUTime bool `toml:"collect_cpu_time"`
	ReportActive   bool `toml:"report_active"`
}

func NewCPUStats(ps PS) *CPUStats {
	return &CPUStats{
		ps:             ps,
		CollectCPUTime: false,
		ReportActive:   true,
	}
}

func (_ *CPUStats) Description() string {
	return "Read metrics about cpu usage"
}

var sampleConfig = `
  ## Whether to report per-cpu stats or not
  percpu = true
  ## Whether to report total system cpu stats or not
  totalcpu = true
  ## If true, collect raw CPU time metrics. (Ignored on Windows)
  collect_cpu_time = false
  ## If true, compute and report the sum of all non-idle CPU states.
  report_active = false
`

func (_ *CPUStats) SampleConfig() string {
	return sampleConfig
}

func (s *CPUStats) Gather(acc telegraf.Accumulator) error {
	times, err := s.ps.CPUTimes(s.PerCPU, s.TotalCPU)
	if err != nil {
		return fmt.Errorf("error getting CPU info: %s", err)
	}
	now := time.Now()

	for _, cts := range times {
		tags := map[string]string{
			"cpu": cts.CPU,
		}

		fieldsG := map[string]interface{}{
			"usage_user":       cts.User,
			"usage_system":     cts.System,
			"usage_idle":       cts.Idle,
			"usage_nice":       cts.Nice,
			"usage_iowait":     cts.Iowait,
			"usage_irq":        cts.Irq,
			"usage_softirq":    cts.Softirq,
			"usage_steal":      cts.Steal,
			"usage_guest":      cts.Guest,
			"usage_guest_nice": cts.GuestNice,
		}
		if s.ReportActive {
			fieldsG["usage_active"] = activeCpuTime(cts)
		}
		acc.AddGauge("cpu", fieldsG, tags, now)
	}

	return err
}

func activeCpuTime(t cpu.TimesStat) float64 {
	active := t.User + t.System + t.Nice + t.Iowait + t.Irq + t.Softirq + t.Steal
	return active
}

func init() {
	inputs.Add("cpu", func() telegraf.Input {
		return &CPUStats{
			PerCPU:   true,
			TotalCPU: true,
			ps:       newSystemPS(),
		}
	})
}
