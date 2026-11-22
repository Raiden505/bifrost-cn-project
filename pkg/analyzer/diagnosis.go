package analyzer

import (
	"fmt"

	"bifrost/pkg/models"
)

func Diagnose(cfg models.TestConfig, stats models.ComputedStats) models.DiagnosisResult {
	var result models.DiagnosisResult

	if cfg.BufferbloatMode && stats.QueueingInflationMs > 20.0 {
		// Rule 1: Bufferbloat
		result.Condition = "Bufferbloat / Queueing Delay Under Load"
		result.Severity = "warning"
		result.Summary = "Saturating the link caused severe queueing delay inside an oversized router buffer."
		result.Explanations = []string{
			fmt.Sprintf("Idle mean RTT was %.2f ms, which spiked to %.2f ms under load (+%.2f ms inflation, exceeding the 20.0 ms threshold).", stats.IdleMeanRTTMs, stats.LoadedMeanRTTMs, stats.QueueingInflationMs),
			fmt.Sprintf("Application Goodput reached %.2f Mbps (%.1f%% of %.1f Mbps link capacity).", stats.GoodputMbps, stats.CapacityUtilization, cfg.LinkCapacityMbps),
		}
	} else if stats.ProbeLossPct >= 1.0 && stats.JitterRFC3550Ms < 8.0 {
		// Rule 2: Random Link Packet Loss
		result.Condition = "Random Packet Loss (Non-Congestive)"
		result.Severity = "critical"
		result.Summary = "High non-congestive packet loss is forcing TCP to back off and reducing Goodput."
		result.Explanations = []string{
			fmt.Sprintf("UDP Probe Loss is %.2f%% (exceeds 1.0%% threshold) while RFC 3550 Jitter remained low at %.2f ms (< 8.0 ms).", stats.ProbeLossPct, stats.JitterRFC3550Ms),
			"Packets are being dropped randomly by a faulty link or policer rather than a full queue.",
			fmt.Sprintf("TCP Goodput dropped to %.2f Mbps (%.1f%% of %.1f Mbps link capacity) due to congestion window reductions.", stats.GoodputMbps, stats.CapacityUtilization, cfg.LinkCapacityMbps),
		}
	} else if stats.JitterRFC3550Ms >= 5.0 || (stats.RTTStdDevMs > 5.0 && stats.CapacityUtilization < 70.0) {
		// Rule 3: Network Congestion
		result.Condition = "Network Congestion / Competing Traffic"
		result.Severity = "critical"
		result.Summary = "High latency variance and degraded throughput indicate active queue congestion."
		result.Explanations = []string{
			fmt.Sprintf("RFC 3550 Jitter reached %.2f ms with an RTT standard deviation of %.2f ms, showing packets fighting competing traffic in a bottleneck queue.", stats.JitterRFC3550Ms, stats.RTTStdDevMs),
			fmt.Sprintf("Application Goodput degraded to %.2f Mbps (%.1f%% of %.1f Mbps expected capacity).", stats.GoodputMbps, stats.CapacityUtilization, cfg.LinkCapacityMbps),
		}
	} else if stats.RTTMeanMs >= 40.0 && stats.JitterRFC3550Ms < 3.0 && stats.ProbeLossPct < 1.0 {
		// Rule 4: High Intrinsic Propagation Delay
		result.Condition = "High Intrinsic Propagation Delay"
		result.Severity = "warning"
		result.Summary = "The network link is stable and lossless, but physically distant."
		result.Explanations = []string{
			fmt.Sprintf("Mean RTT is elevated at %.2f ms (>= 40.0 ms), but jitter is rock-steady at %.2f ms (< 3.0 ms) with %.2f%% packet loss.", stats.RTTMeanMs, stats.JitterRFC3550Ms, stats.ProbeLossPct),
			"Consistent latency with zero queueing delay indicates physical propagation distance rather than congestion.",
		}
	} else if stats.CapacityUtilization < 50.0 && stats.RTTMeanMs < 20.0 && stats.ProbeLossPct < 1.0 {
		// Rule 5: Bandwidth Limitation
		result.Condition = "Bandwidth Bottleneck / Rate Cap"
		result.Severity = "warning"
		result.Summary = "Latency and packet loss are clean, but throughput is strictly capped below expected capacity."
		result.Explanations = []string{
			fmt.Sprintf("Goodput is capped at %.2f Mbps (%.1f%% of %.1f Mbps expected reference capacity).", stats.GoodputMbps, stats.CapacityUtilization, cfg.LinkCapacityMbps),
			fmt.Sprintf("Latency (%.2f ms mean RTT), jitter (%.2f ms), and loss (%.2f%%) are completely clean, pointing to a hard traffic shaper or rate cap.", stats.RTTMeanMs, stats.JitterRFC3550Ms, stats.ProbeLossPct),
		}
	} else {
		// Rule 6: Healthy Network
		result.Condition = "Healthy Network"
		result.Severity = "healthy"
		result.Summary = "Throughput, latency, jitter, and packet loss are all within normal operating limits."
		result.Explanations = []string{
			fmt.Sprintf("Achieved %.2f Mbps Goodput (%.1f%% capacity utilization, ~%.1f%% protocol overhead).", stats.GoodputMbps, stats.CapacityUtilization, stats.EstimatedOverhead),
			fmt.Sprintf("Mean RTT is %.2f ms (P95: %.2f ms) with %.2f ms jitter and %.2f%% packet loss.", stats.RTTMeanMs, stats.RTTP95Ms, stats.JitterRFC3550Ms, stats.ProbeLossPct),
		}
	}

	return result
}
