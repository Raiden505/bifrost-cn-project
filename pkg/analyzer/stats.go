package analyzer

import (
	"bifrost/pkg/models"
)

func CalculateStats(cfg models.TestConfig, raw models.RawTestOutput) models.ComputedStats {
	var stats models.ComputedStats
	stats.EstimatedOverhead = 5.4

	if raw.ActualTCPDurSec > 0 {
		//stats.GoodPutMbps computes
		stats.GoodputMbps = (float64(raw.TotalTCPBytes) * 8.0) / (raw.ActualTCPDurSec * 1e6)

	}

	if cfg.LinkCapacityMbps > 0 {
		//stats.CapacityUtilization compute
		stats.CapacityUtilization = (stats.GoodputMbps / cfg.LinkCapacityMbps) * 100.0
	}

	stats.ProbesSent = len(raw.UDPSamples)

	var validRTTs []float64
	var idleSum, loadedSum float64
	var idleCount, loadedCount int

	for _, sample := range raw.UDPSamples {
		if sample.Lost {
			continue
		}

		validRTTs = append(validRTTs, sample.RTTMs)

		if sample.Phase == "idle" {
			idleSum = idleSum + sample.RTTMs
			idleCount++
		}

		if sample.Phase == "loaded" {
			loadedSum = loadedSum + sample.RTTMs
			loadedCount++
		}

	}

	stats.ProbesReceived = len(validRTTs)

	if stats.ProbesSent > 0 {
		lostProbes := stats.ProbesSent - stats.ProbesReceived
		stats.ProbeLossPct = float64(lostProbes) / float64(stats.ProbesSent) * 100
	}

	return stats

}
