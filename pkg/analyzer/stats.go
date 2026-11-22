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

	return stats

}
