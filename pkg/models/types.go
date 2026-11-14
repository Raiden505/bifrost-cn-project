package models

import "time"

// 1. INPUT: Sent from Zaynab's Web UI -> Ayaan's Client Engine
type TestConfig struct {
	ServerIP         string  `json:"server_ip"`          // Default: "10.0.0.2"
	DurationSec      int     `json:"duration_sec"`       // Default: 5
	ProbeIntervalMs  int     `json:"probe_interval_ms"`  // Default: 20 (50 probes/sec)
	LinkCapacityMbps float64 `json:"link_capacity_mbps"` // Default: 100.0 (for comparison)
	BufferbloatMode  bool    `json:"bufferbloat_mode"`   // False = Sequential, True = Idle then Loaded
	ImpairmentPreset string  `json:"impairment_preset"`  // Optional UI preset: "none", "rate10", "delay50", "loss3"
}

// 2. RAW DATA: Produced by Ayaan's Networking Engine -> Sent to Mustafa's Analyzer
type RawProbeSample struct {
	SeqNum   uint64  `json:"seq_num"`
	SentTime int64   `json:"sent_time_ns"`
	RecvTime int64   `json:"recv_time_ns"`
	RTTMs    float64 `json:"rtt_ms"`
	Lost     bool    `json:"lost"`
	Phase    string  `json:"phase"` // "normal", "idle", or "loaded"
}

type RawTCPInterval struct {
	IntervalIndex int     `json:"interval_index"` // 1, 2, 3... (every 250ms)
	ElapsedSec    float64 `json:"elapsed_sec"`    // 0.25, 0.50, 0.75...
	BytesRead     uint64  `json:"bytes_read"`     // Bytes read in this 250ms window
	GoodputMbps   float64 `json:"goodput_mbps"`   // Instantaneous goodput for this window
}

type RawTestOutput struct {
	TotalTCPBytes   uint64           `json:"total_tcp_bytes"`
	ActualTCPDurSec float64          `json:"actual_tcp_dur_sec"`
	TCPIntervals    []RawTCPInterval `json:"tcp_intervals"`
	UDPSamples      []RawProbeSample `json:"udp_samples"`
}

// 3. COMPUTED STATS & DIAGNOSIS: Produced by Mustafa's Analyzer -> Sent to Zaynab's UI
type ComputedStats struct {
	GoodputMbps         float64 `json:"goodput_mbps"`
	CapacityUtilization float64 `json:"capacity_utilization_pct"` // Goodput / LinkCapacity * 100
	EstimatedOverhead   float64 `json:"estimated_overhead_pct"`   // ~5.4% for standard TCP/IP/Eth
	ProbesSent          int     `json:"probes_sent"`
	ProbesReceived      int     `json:"probes_received"`
	ProbeLossPct        float64 `json:"probe_loss_pct"`
	RTTMinMs            float64 `json:"rtt_min_ms"`
	RTTMeanMs           float64 `json:"rtt_mean_ms"`
	RTTMedianMs         float64 `json:"rtt_median_ms"`
	RTTP95Ms            float64 `json:"rtt_p95_ms"`
	RTTMaxMs            float64 `json:"rtt_max_ms"`
	RTTStdDevMs         float64 `json:"rtt_stddev_ms"`
	JitterRFC3550Ms     float64 `json:"jitter_rfc3550_ms"`
	// Bufferbloat Extension Stats
	IdleMeanRTTMs       float64 `json:"idle_mean_rtt_ms"`
	LoadedMeanRTTMs     float64 `json:"loaded_mean_rtt_ms"`
	QueueingInflationMs float64 `json:"queueing_inflation_ms"` // LoadedMean - IdleMean
}

type DiagnosisResult struct {
	Condition    string   `json:"condition"`    // e.g., "Healthy Network", "Random Packet Loss"
	Severity     string   `json:"severity"`     // "healthy" (green), "warning" (yellow), "critical" (red)
	Summary      string   `json:"summary"`      // One-sentence plain English verdict
	Explanations []string `json:"explanations"` // Bullet points with exact numbers backing the verdict
}

// 4. FINAL REPORT: The Master JSON sent to Zaynab's Web Dashboard
type FinalReport struct {
	Timestamp string          `json:"timestamp"`
	Config    TestConfig      `json:"config"`
	Raw       RawTestOutput   `json:"raw"`
	Stats     ComputedStats   `json:"stats"`
	Diagnosis DiagnosisResult `json:"diagnosis"`
}

// HELPER FOR ZAYNAB: Generates fake data so Web UI can be built on Saturday!
func GetMockReport() FinalReport {
	return FinalReport{
		Timestamp: time.Now().Format("15:04:05"),
		Config: TestConfig{
			ServerIP: "10.0.0.2", DurationSec: 3, ProbeIntervalMs: 500, LinkCapacityMbps: 100,
		},
		Raw: RawTestOutput{
			TotalTCPBytes: 11250000, ActualTCPDurSec: 3.0,
			TCPIntervals: []RawTCPInterval{
				{1, 0.5, 1875000, 30.0}, {2, 1.0, 1800000, 28.8}, {3, 1.5, 1950000, 31.2},
				{4, 2.0, 1750000, 28.0}, {5, 2.5, 1900000, 30.4}, {6, 3.0, 1975000, 31.6},
			},
			UDPSamples: []RawProbeSample{
				{1, 0, 0, 10.2, false, "normal"}, {2, 0, 0, 10.5, false, "normal"},
				{3, 0, 0, 0.0, true, "normal"}, {4, 0, 0, 10.1, false, "normal"},
				{5, 0, 0, 10.4, false, "normal"}, {6, 0, 0, 10.3, false, "normal"},
			},
		},
		Stats: ComputedStats{
			GoodputMbps: 30.0, CapacityUtilization: 30.0, EstimatedOverhead: 5.4,
			ProbesSent: 6, ProbesReceived: 5, ProbeLossPct: 16.67,
			RTTMinMs: 10.1, RTTMeanMs: 10.3, RTTMedianMs: 10.3, RTTP95Ms: 10.5, RTTMaxMs: 10.5,
			RTTStdDevMs: 0.15, JitterRFC3550Ms: 0.21,
		},
		Diagnosis: DiagnosisResult{
			Condition: "Random Packet Loss Detected",
			Severity:  "critical",
			Summary:   "High non-congestive packet loss is causing TCP congestion backoff.",
			Explanations: []string{
				"UDP Probe Loss is 16.67% (exceeds 1.0% threshold).",
				"Base RTT remained stable at 10.30 ms with low jitter (0.21 ms < 5.0 ms threshold), ruling out queue congestion.",
				"Application Goodput dropped to 30.0 Mbps (30.0% of 100 Mbps link capacity) because TCP reduced its sending window upon detecting dropped packets.",
			},
		},
	}
}
