package analyzer

import (
	"bifrost/pkg/models"
)

func CalculateStats(cfg models.TestConfig, raw models.RawTestOutput) models.ComputedStats {
	var stats models.ComputedStats
	stats.EstimatedOverhead = 5.4

	if raw.ActualTCPDurSec > 0 {
		//stats.GoodPutMbps compute
	}

	if cfg.LinkCapacityMbps > 0 {
		//stats.CapacityUtilization compute
	}

	return stats

}
