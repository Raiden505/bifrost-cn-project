package analyzer

import (
	"bifrost/pkg/models"
	"math"
	"slices"
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

	var jitter float64

	for i := 1; i < len(validRTTs); i++ {
		diff := math.Abs(validRTTs[i] - validRTTs[i-1])
		jitter = jitter + (diff-jitter)/16.0
	}
	stats.JitterRFC3550Ms = jitter

	var RTTSum float64
	var varianceSum float64

	if len(validRTTs) > 0 {
		for i := range len(validRTTs) {
			RTTSum = RTTSum + validRTTs[i]
		}
		stats.RTTMeanMs = RTTSum / float64(len(validRTTs))

		for i := range len(validRTTs) {
			diff := validRTTs[i] - stats.RTTMeanMs
			varianceSum = varianceSum + (diff * diff)
		}
		stats.RTTStdDevMs = math.Sqrt(varianceSum / float64(len(validRTTs)))

		slices.Sort(validRTTs)
		stats.RTTMinMs = validRTTs[0]
		stats.RTTMaxMs = validRTTs[len(validRTTs)-1]
		stats.RTTMedianMs = validRTTs[len(validRTTs)/2]
		stats.RTTP95Ms = validRTTs[int(float64(len(validRTTs))*0.95)]

	}

	return stats

}
