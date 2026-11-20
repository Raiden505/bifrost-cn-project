# Bifrost: The "No-Jargon" Master Blueprint & Weekend Execution Guide

If everything so far felt like a wall of networking buzzwords, take a breath. At its core, **Bifrost is just a stopwatch and a byte-counter with a pretty webpage attached to it.**

This document explains:
1. **How the app actually works** in plain English.
2. **How the 3 of you will work in parallel on GitHub** without waiting on each other or getting merge conflicts.
3. **The exact shared code contract (`types.go`)** you all agree on before starting.
4. **Step-by-step instructions for Ayaan, Mustafa, and Zaynab** (ordered from easiest to hardest so you learn Go naturally as you build).

---

## Part 1: How the App Works (In Plain English)

Imagine you want to test how good the road is between **House A (Client)** and **House B (Server)**. You can't just send one car and guess. Instead, Bifrost runs two very simple tests between two programs running on your computer:

### Test 1: The Ping-Pong Test (UDP Probes — Measures Delay, Jitter & Packet Loss)
* **Why UDP?** When you send a message over TCP, the operating system secretly retries lost packets behind your back. You never find out a packet was dropped! **UDP is "fire and forget"**—if a packet drops, it's gone forever. That makes UDP perfect for catching the network red-handed when it drops or delays packets.
* **How it works:**
  1. Every 20 milliseconds, the **Client** sends a tiny 64-byte UDP packet to the **Server** on **Port `9002`**.
  2. Inside that packet, the Client stamps two numbers:
     * **Packet ID (`SeqNum`):** `1, 2, 3, 4...`
     * **Time Sent (`SendTime`):** The exact nanosecond the packet left the Client.
  3. The **Server** catches the packet and **immediately throws it right back** to the Client (like a ping-pong ball hitting a wall).
  4. When the Client catches the returning packet, it looks at its own clock (`ReceiveTime`):
     * **RTT (Round-Trip Time):** `ReceiveTime - SendTime`. (Because the Client uses its own stopwatch for both timestamps, you never have to worry about the Server's clock being out of sync!).
     * **Packet Loss:** If the Client sends 100 packets (`ID 1` to `100`), and never gets `ID 42` back within 1 second, `ID 42` is marked **Lost**.
     * **Jitter (Delay Variation):** Did every packet take smoothly `10ms`, `10ms`, `10ms`? Or did it jump wildly like `10ms`, `85ms`, `12ms`, `110ms`? That jumpiness is **Jitter**. High jitter means a router buffer somewhere is getting clogged up and emptying erratically.

### Test 2: The Firehose Test (TCP Transfer — Measures Goodput)
* **Why TCP?** Real apps (downloads, Netflix, web browsing) use TCP. We want to see how fast a real app can push actual useful data through the pipe.
* **Goodput vs. Link Capacity (The Rubric Question):**
  * **Link Capacity (Bandwidth):** The speed limit painted on the highway (e.g., `100 Mbps`).
  * **Application Goodput:** How much actual cargo (your app's data) reaches the destination per second. Because every packet wears "packaging" (Ethernet + IP + TCP headers $\approx 54\text{ bytes}$ per $1514\text{ byte}$ frame) and because TCP slows down whenever a packet drops, **Goodput is always lower than Link Capacity** (usually around `94.6 Mbps` on a clean `100 Mbps` link, or way lower if there is packet loss).
* **How it works:**
  1. The **Client** connects to the **Server** on **Port `9001` (TCP)**.
  2. For 5 seconds, the Client blasts `64 KB` chunks of dummy bytes as fast as the OS allows.
  3. The **Server** reads those bytes in a loop and counts how many bytes arrived every `250 milliseconds`.
  4. Formula for Goodput in Mbps:
     $$\text{Goodput (Mbps)} = \frac{\text{Total Bytes Received} \times 8}{\text{Seconds Elapsed} \times 1,000,000}$$

### Test 3: The Walkie-Talkie Channel (TCP Control — Port `9000`)
Before running Test 1 or Test 2, the Client connects to the Server on **Port `9000` (TCP)** and sends a small JSON message saying: *"Hey, get ready! We are running a 5-second test with 20ms UDP pings."* When the tests finish, the Server uses this same connection to send back its byte counts so the Client has all the numbers in one place.

### How All 3 People's Code Connects Together
Here is the entire lifecycle of one click in the app:

```
[1. ZAYNAB: Web Browser]
   User clicks "Run Test" on http://localhost:8080
         │
         ▼ (Sends JSON settings to Go HTTP Server)
[2. AYAAN: Networking Engine]
   Client talks to Server over Ports 9000, 9001, 9002.
   Returns raw lists of numbers (Raw UDP pings + Raw TCP byte counts).
         │
         ▼ (Passes raw arrays to Analyzer)
[3. MUSTAFA: Math & Diagnosis Engine]
   Calculates Averages, Jitter, Loss %, Goodput.
   Runs if/else rules to diagnose the problem (e.g., "Random Packet Loss!").
         │
         ▼ (Returns FinalReport struct)
[4. ZAYNAB: Web Browser]
   Receives FinalReport JSON and draws Chart.js graphs + Verdict Box!
```

---

## Part 2: How to Collaborate on GitHub Without Conflicts

If all three of you edit the same files, Git will give you merge conflicts and slow you down. Instead, we use **Contract-First Development**:

1. **Folder Ownership:** Each person owns their own folder. Nobody touches anyone else's folder!
   * `pkg/models/` $\rightarrow$ **Shared** (You paste this in once on Friday night and never touch it again).
   * `pkg/netengine/` $\rightarrow$ **Ayaan's folder ONLY.**
   * `pkg/analyzer/` & `scripts/` $\rightarrow$ **Mustafa's folder ONLY.**
   * `pkg/web/` $\rightarrow$ **Zaynab's folder ONLY.**
2. **How can Mustafa and Zaynab work on Saturday if Ayaan hasn't finished the sockets yet?**
   * Because `pkg/models/types.go` defines the exact input and output structures!
   * On Saturday, **Mustafa** feeds fake arrays of numbers into his math functions to test them.
   * On Saturday, **Zaynab** uses a built-in helper function (`models.GetMockReport()`) that returns realistic fake data so she can build and style the entire webpage and graphs without needing the server running.
   * On **Sunday Afternoon**, you replace the 1 line of fake data with Ayaan's and Mustafa's real functions. Everything snaps together like Lego bricks.

---

## Part 3: Step 0 — The Shared Setup (Do This Together in 20 Mins)

One person creates the GitHub repo, runs these commands in a terminal, and pushes to `main`:

```bash
mkdir bifrost && cd bifrost
go mod init bifrost
mkdir -p pkg/models pkg/netengine pkg/analyzer pkg/web/static scripts
```

Now, create **`pkg/models/types.go`** and paste this exact code. **This is your team's contract:**

```go
package models

import "time"

// 1. INPUT: Sent from Zaynab's Web UI -> Ayaan's Client Engine
type TestConfig struct {
	ServerIP          string  `json:"server_ip"`           // Default: "10.0.0.2"
	DurationSec       int     `json:"duration_sec"`        // Default: 5
	ProbeIntervalMs   int     `json:"probe_interval_ms"`   // Default: 20 (50 probes/sec)
	LinkCapacityMbps  float64 `json:"link_capacity_mbps"`  // Default: 100.0 (for comparison)
	BufferbloatMode   bool    `json:"bufferbloat_mode"`    // False = Sequential, True = Idle then Loaded
	ImpairmentPreset  string  `json:"impairment_preset"`   // Optional UI preset: "none", "rate10", "delay50", "loss3"
}

// 2. RAW DATA: Produced by Ayaan's Networking Engine -> Sent to Mustafa's Analyzer
type RawProbeSample struct {
	SeqNum    uint64  `json:"seq_num"`
	SentTime  int64   `json:"sent_time_ns"`
	RecvTime  int64   `json:"recv_time_ns"`
	RTTMs     float64 `json:"rtt_ms"`
	Lost      bool    `json:"lost"`
	Phase     string  `json:"phase"` // "normal", "idle", or "loaded"
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
	Condition   string   `json:"condition"`   // e.g., "Healthy Network", "Random Packet Loss"
	Severity    string   `json:"severity"`    // "healthy" (green), "warning" (yellow), "critical" (red)
	Summary     string   `json:"summary"`     // One-sentence plain English verdict
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
				{3, 0, 0, 0.0, true, "normal"},   {4, 0, 0, 10.1, false, "normal"},
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
```

---

## Part 4: Quick 5-Minute Go Crash Course (For People Who Know Python/C++/JS)

You already know how to code, so here are the only 5 Go quirks you need to know this weekend:
1. **Exported (Public) vs. Private:** In Go, if a function or variable starts with a **CAPITAL letter** (`RunTest`, `GoodputMbps`), other folders can import and use it. If it starts with a **lowercase letter** (`runTest`), it is private to that folder.
2. **Error Handling:** Go doesn't use `try/catch`. Functions return two things: the result and an `error`. You always write:
   ```go
   conn, err := net.Dial("tcp", "10.0.0.2:9000")
   if err != nil {
       return err // Something went wrong!
   }
   defer conn.Close() // "defer" automatically closes the socket when the function finishes!
   ```
3. **Slices (Dynamic Arrays):**
   ```go
   var samples []models.RawProbeSample // Creates an empty dynamic list
   samples = append(samples, newSample) // Pushes an item to the list
   ```
4. **Loops:** Go only has `for` loops.
   * Standard loop: `for i := 0; i < 10; i++ { ... }`
   * Foreach loop: `for index, item := range samples { ... }`
   * While loop: `for time.Since(start) < duration { ... }`
5. **Goroutines (Background Threads):** Putting the word `go` in front of a function call runs it in the background immediately!
   ```go
   go startUDPServer() // Runs in background!
   ```

---

## Part 5: Individual Step-by-Step Guides

---

### 👤 Guide 1: Muhammad Ayaan — Core Networking (`pkg/netengine`)

**Your Goal:** Write the Server (`server.go`) and the Client (`client.go`) that open ports `9000`, `9001`, and `9002`, run the TCP & UDP tests, and fill out the `models.RawTestOutput` struct.

#### Step 1 (Saturday Morning): The UDP Echo Server & Client (Learn Go UDP Sockets)
Start with UDP because it is stateless and super short.
* **Server Side (`:9002`):**
  1. Listen on UDP port `9002`:
     ```go
     addr, _ := net.ResolveUDPAddr("udp", ":9002")
     conn, err := net.ListenUDP("udp", addr)
     ```
  2. In an infinite loop (`for {}`), read a packet into a 2048-byte buffer and immediately write those exact bytes back to the sender's address:
     ```go
     buf := make([]byte, 2048)
     for {
         n, clientAddr, err := conn.ReadFromUDP(buf)
         if err == nil {
             conn.WriteToUDP(buf[:n], clientAddr) // Echo right back!
         }
     }
     ```
* **Client Side (`udp_probe.go`):**
  1. Connect to `ServerIP:9002` using `net.Dial("udp", serverIP+":9002")`.
  2. Create a 64-byte buffer: `packet := make([]byte, 64)`.
  3. Every `ProbeIntervalMs` (use `time.Sleep(time.Duration(intervalMs) * time.Millisecond)`):
     * Increment `seqNum` (`1, 2, 3...`).
     * Record current time in nanoseconds: `sentNs := time.Now().UnixNano()`.
     * Pack `seqNum` and `sentNs` into the first 16 bytes of `packet` using `encoding/binary`:
       ```go
       binary.BigEndian.PutUint64(packet[0:8], seqNum)
       binary.BigEndian.PutUint64(packet[8:16], uint64(sentNs))
       ```
     * Send the packet (`conn.Write(packet)`).
     * Set a timeout so if the packet is dropped by the network, `conn.Read` doesn't hang forever!
       ```go
       conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
       _, err := conn.Read(recvBuf)
       if err != nil {
           // Packet timed out -> IT WAS LOST!
           samples = append(samples, models.RawProbeSample{SeqNum: seqNum, Lost: true, Phase: phase})
       } else {
           recvNs := time.Now().UnixNano()
           rttMs := float64(recvNs - sentNs) / 1e6 // Convert nanoseconds to milliseconds
           samples = append(samples, models.RawProbeSample{
               SeqNum: seqNum, SentTime: sentNs, RecvTime: recvNs, RTTMs: rttMs, Lost: false, Phase: phase,
           })
       }
       ```

#### Step 2 (Saturday Afternoon): The TCP Goodput Engine (`:9001`)
* **Server Side (`:9001`):**
  1. Listen on TCP `:9001` (`net.Listen("tcp", ":9001")`).
  2. When the client connects (`conn, _ := listener.Accept()`), start a timer (`start := time.Now()`, `windowStart := time.Now()`).
  3. Read bytes in a loop using a 64KB buffer (`buf := make([]byte, 65536)`):
     * Add `n` (bytes read) to `totalBytes` and `windowBytes`.
     * Whenever `time.Since(windowStart) >= 250 * time.Millisecond`:
       * Calculate `mbps := (float64(windowBytes) * 8.0) / (time.Since(windowStart).Seconds() * 1e6)`.
       * Append a `models.RawTCPInterval` to your list, reset `windowBytes = 0`, and reset `windowStart = time.Now()`.
     * When `conn.Read(buf)` returns `io.EOF` (client closed the connection), stop and store the final `totalBytes`, `actualDuration`, and `intervals` so the Control Channel can send them to the client!
* **Client Side (`tcp_goodput.go`):**
  1. Connect to `ServerIP:9001` (`net.Dial("tcp", serverIP+":9001")`).
  2. Create a 64KB slice (`chunk := make([]byte, 65536)`).
  3. Loop `for time.Since(start) < time.Duration(durationSec)*time.Second` and call `conn.Write(chunk)`.
  4. Call `conn.Close()` so the server sees `EOF` and knows the transfer is done.

#### Step 3 (Saturday Night): The TCP Control Channel (`:9000`) & Concurrency
* **How `:9000` ties it together:**
  1. Client connects to `ServerIP:9000` and sends `TestConfig` using `json.NewEncoder(ctrlConn).Encode(cfg)`.
  2. Server reads `TestConfig` and replies `"READY\n"`.
  3. Client runs the tests:
     * **If `cfg.BufferbloatMode == false` (Standard Sequential):**
       1. Run `runUDPProbes(cfg, "normal")`.
       2. Run `runTCPStream(cfg)`.
     * **If `cfg.BufferbloatMode == true` (Advanced Extension — Idle vs. Loaded):**
       1. Run `runUDPProbes(3 seconds, "idle")` alone first to get clean baseline RTT.
       2. Use `sync.WaitGroup` to run `runUDPProbes(cfg.DurationSec, "loaded")` and `runTCPStream(cfg.DurationSec)` **at the exact same time**!
          ```go
          var wg sync.WaitGroup
          wg.Add(2)
          go func() { defer wg.Done(); udpLoaded = runUDPProbes(cfg, "loaded") }()
          go func() { defer wg.Done(); runTCPStream(cfg) }()
          wg.Wait() // Waits for both to finish!
          ```
  4. Finally, Client sends `"GET_RESULTS\n"` on `ctrlConn`, and Server sends back the `TotalTCPBytes` and `TCPIntervals` as JSON!

---

### 👤 Guide 2: Mustafa — Math, Diagnostic Rules & Linux Testbed (`pkg/analyzer` & `scripts/`)

**Your Goal:** Write the math/diagnosis functions that turn Ayaan's raw numbers into a `models.FinalReport`, and write two tiny bash scripts that create the virtual Linux network and inject lag/packet loss.

#### Step 1 (Saturday Morning): The Math Aggregator (`pkg/analyzer/stats.go`)
Write `func CalculateStats(cfg models.TestConfig, raw models.RawTestOutput) models.ComputedStats`:
1. **Goodput & Overhead Math:**
   * `goodputMbps = (float64(raw.TotalTCPBytes) * 8.0) / (raw.ActualTCPDurSec * 1e6)`
   * `capacityUtil = (goodputMbps / cfg.LinkCapacityMbps) * 100.0`
   * **Protocol Overhead %:** Every standard 1460-byte TCP payload requires 20B TCP header + 20B IPv4 header + 14B Ethernet header = 54 bytes of non-application framing ($\frac{54}{1514} \approx 3.57\%$ to $5.4\%$ with ACKs/options). Set `EstimatedOverhead = 5.4` (or compute wire rate vs goodput) so the report explicitly distinguishes application goodput from link capacity!
2. **RTT Percentiles & Standard Deviation:**
   * Loop through `raw.UDPSamples`. Count total `ProbesSent`. If `sample.Lost == false`, add `sample.RTTMs` to a float slice `validRTTs`.
   * `ProbeLossPct = float64(ProbesSent - len(validRTTs)) / float64(ProbesSent) * 100.0`
   * Sort `validRTTs` using `slices.Sort(validRTTs)` (import `"slices"`):
     * `RTTMinMs = validRTTs[0]`
     * `RTTMaxMs = validRTTs[len(validRTTs)-1]`
     * `RTTMedianMs = validRTTs[len(validRTTs)/2]`
     * `RTTP95Ms = validRTTs[int(float64(len(validRTTs))*0.95)]`
   * Calculate the average (`RTTMeanMs`) and standard deviation:
     $$\sigma = \sqrt{\frac{\sum (x_i - \text{mean})^2}{N}}$$
3. **RFC 3550 Jitter Formula:**
   * In a loop over `validRTTs` from index `1` to the end:
     ```go
     diff := math.Abs(validRTTs[i] - validRTTs[i-1])
     jitter = jitter + (diff - jitter) / 16.0
     ```
   * That 3-line loop is literally the official IETF RFC 3550 smoothed jitter calculation!
4. **Bufferbloat Extension Math:**
   * Average the RTT of samples where `sample.Phase == "idle"` (`IdleMeanRTTMs`) and where `sample.Phase == "loaded"` (`LoadedMeanRTTMs`).
   * `QueueingInflationMs = LoadedMeanRTTMs - IdleMeanRTTMs`.

#### Step 2 (Saturday Afternoon): The 5-Rule Diagnostic Engine (`pkg/analyzer/diagnosis.go`)
Write `func Diagnose(cfg models.TestConfig, stats models.ComputedStats) models.DiagnosisResult`:
Use clean `if / else if` rules that check combinations of metrics and generate human-readable explanations with `fmt.Sprintf`:

1. **Rule 1 — Bufferbloat / Standing Queue Under Load (If `cfg.BufferbloatMode` is ON):**
   * *Trigger:* `stats.QueueingInflationMs > 20.0`
   * *Verdict:* `"Bufferbloat / Queueing Delay Under Load"` (`Severity: "warning"`)
   * *Explanation:* Show `IdleMeanRTTMs` vs `LoadedMeanRTTMs` and explain that saturating the link caused packets to wait in an oversized router buffer.
2. **Rule 2 — Random Link Packet Loss:**
   * *Trigger:* `stats.ProbeLossPct >= 1.0 && stats.JitterRFC3550Ms < 8.0`
   * *Verdict:* `"Random Packet Loss (Non-Congestive)"` (`Severity: "critical"`)
   * *Explanation:* Point out that UDP loss is high ($\ge 1\%$) while jitter is low, meaning packets are being dropped randomly by a faulty link or policer rather than a full queue, forcing TCP to cut its congestion window and tanking Goodput.
3. **Rule 3 — Network Congestion / Competing Traffic:**
   * *Trigger:* `stats.JitterRFC3550Ms >= 5.0 || (stats.RTTStdDevMs > 5.0 && stats.CapacityUtilization < 70.0)`
   * *Verdict:* `"Network Congestion / Competing Traffic"` (`Severity: "critical"`)
   * *Explanation:* High jitter ($\ge 5\text{ ms}$) combined with degraded TCP goodput proves our packets are fighting competing traffic inside a bottleneck queue.
4. **Rule 4 — High Intrinsic Propagation Delay:**
   * *Trigger:* `stats.RTTMeanMs >= 40.0 && stats.JitterRFC3550Ms < 3.0 && stats.ProbeLossPct < 1.0`
   * *Verdict:* `"High Intrinsic Propagation Delay"` (`Severity: "warning"`)
   * *Explanation:* RTT is high ($\ge 40\text{ ms}$) but rock-steady (jitter $< 3\text{ ms}$) with $0\%$ loss. This means the link is physically long/distant, not congested.
5. **Rule 5 — Bandwidth Limitation (Traffic Shaper / Cap):**
   * *Trigger:* `stats.CapacityUtilization < 50.0 && stats.RTTMeanMs < 20.0 && stats.ProbeLossPct < 1.0`
   * *Verdict:* `"Bandwidth Bottleneck / Rate Cap"` (`Severity: "warning"`)
   * *Explanation:* Latency, jitter, and loss are completely clean, yet Goodput is strictly capped at `X Mbps` ($Y\%$ of expected reference capacity), indicating a hard link speed limit.
6. **Rule 6 — Default (`else`):**
   * *Verdict:* `"Healthy Network"` (`Severity: "healthy"`)

#### Step 3 (Sunday Morning): The 2 Linux Testbed Scripts (`scripts/`)
Instead of using two laptops, Linux lets you create two virtual mini-computers (`client_ns` and `server_ns`) connected by a virtual cable (`veth0` $\leftrightarrow$ `veth1`).

Create **`scripts/setup_netns.sh`** (run `chmod +x scripts/*.sh`):
```bash
#!/usr/bin/env bash
set -e
# Clean up old namespaces if they exist
sudo ip netns del client_ns 2>/dev/null || true
sudo ip netns del server_ns 2>/dev/null || true

# 1. Create two virtual network namespaces
sudo ip netns add client_ns
sudo ip netns add server_ns

# 2. Create a virtual ethernet cable (veth0 <-> veth1)
sudo ip link add veth0 type veth peer name veth1

# 3. Plug veth0 into client_ns and veth1 into server_ns
sudo ip link set veth0 netns client_ns
sudo ip link set veth1 netns server_ns

# 4. Assign IP addresses (10.0.0.1 and 10.0.0.2) and turn them ON
sudo ip netns exec client_ns ip addr add 10.0.0.1/24 dev veth0
sudo ip netns exec client_ns ip link set veth0 up
sudo ip netns exec client_ns ip link set lo up

sudo ip netns exec server_ns ip addr add 10.0.0.2/24 dev veth1
sudo ip netns exec server_ns ip link set veth1 up
sudo ip netns exec server_ns ip link set lo up

# 5. Set default baseline speed to 100 Mbps
sudo ip netns exec client_ns tc qdisc add dev veth0 root netem rate 100mbit
sudo ip netns exec server_ns tc qdisc add dev veth1 root netem rate 100mbit

echo "✅ Testbed Ready! Client: 10.0.0.1 (client_ns) <---> Server: 10.0.0.2 (server_ns)"
```

Create **`scripts/impair.sh`** (lets you switch network conditions in 1 second!):
```bash
#!/usr/bin/env bash
MODE=$1
# Clear existing rules
sudo ip netns exec client_ns tc qdisc del dev veth0 root 2>/dev/null || true
sudo ip netns exec server_ns tc qdisc del dev veth1 root 2>/dev/null || true

case "$MODE" in
  "baseline"|"none")
    sudo ip netns exec client_ns tc qdisc add dev veth0 root netem rate 100mbit delay 1ms
    sudo ip netns exec server_ns tc qdisc add dev veth1 root netem rate 100mbit delay 1ms
    echo "🟢 Applied: Healthy Baseline (100 Mbps, 2ms RTT, 0% Loss)"
    ;;
  "bandwidth"|"rate10")
    sudo ip netns exec client_ns tc qdisc add dev veth0 root netem rate 10mbit delay 1ms
    sudo ip netns exec server_ns tc qdisc add dev veth1 root netem rate 10mbit delay 1ms
    echo "🟡 Applied: Bandwidth Limitation (10 Mbps Cap, 2ms RTT, 0% Loss)"
    ;;
  "delay"|"delay50")
    sudo ip netns exec client_ns tc qdisc add dev veth0 root netem rate 100mbit delay 50ms
    sudo ip netns exec server_ns tc qdisc add dev veth1 root netem rate 100mbit delay 50ms
    echo "🟡 Applied: High Intrinsic Delay (100 Mbps, 100ms RTT, 0% Loss)"
    ;;
  "loss"|"loss3")
    sudo ip netns exec client_ns tc qdisc add dev veth0 root netem rate 100mbit delay 5ms loss 3%
    sudo ip netns exec server_ns tc qdisc add dev veth1 root netem rate 100mbit delay 5ms loss 3%
    echo "🔴 Applied: Random Packet Loss (100 Mbps, 10ms RTT, 3% Loss)"
    ;;
  "congestion")
    sudo ip netns exec client_ns tc qdisc add dev veth0 root netem rate 20mbit delay 25ms 18ms distribution normal loss 1.5%
    sudo ip netns exec server_ns tc qdisc add dev veth1 root netem rate 20mbit delay 25ms 18ms distribution normal loss 1.5%
    echo "🔴 Applied: Heavy Queueing Congestion (20 Mbps, High Jitter ±18ms)"
    ;;
  "bufferbloat")
    sudo ip netns exec client_ns tc qdisc add dev veth0 root netem rate 15mbit delay 5ms limit 2000
    sudo ip netns exec server_ns tc qdisc add dev veth1 root netem rate 15mbit delay 5ms limit 2000
    echo "🟣 Applied: Bufferbloat Profile (15 Mbps Bottleneck with Deep Queue)"
    ;;
esac
```

---

### 👤 Guide 3: Zaynab Naqqi Raja — Web Dashboard & API (`pkg/web`)

**Your Goal:** Build the Go HTTP web server (`pkg/web/api.go`) and the single-page frontend (`pkg/web/static/index.html`) that triggers tests, displays the KPI cards and Diagnosis Verdict, and draws two Chart.js graphs.

#### Step 1 (Saturday Morning): The Go Web Server (`pkg/web/api.go`)
Using Go's built-in `net/http`, you only need ~50 lines of code:
1. Serve static files from `pkg/web/static`:
   ```go
   http.Handle("/", http.FileServer(http.Dir("pkg/web/static")))
   ```
2. Create `POST /api/run`:
   * On Saturday, simply return `models.GetMockReport()` so you can build your HTML/JS immediately:
     ```go
     http.HandleFunc("/api/run", func(w http.ResponseWriter, r *http.Request) {
         w.Header().Set("Content-Type", "application/json")
         report := models.GetMockReport() // On Sunday, replace with Ayaan + Mustafa's functions!
         json.NewEncoder(w).Encode(report)
     })
     ```
3. **Bonus Stand-Out Feature (`/api/impair`):** Add a 10-line endpoint that runs `exec.Command("bash", "scripts/impair.sh", preset).Run()` so you can switch Linux `tc/netem` network impairments right from a dropdown menu on the webpage during your demo!

#### Step 2 (Saturday Afternoon): Dashboard HTML (`pkg/web/static/index.html`)
Include Tailwind CSS and Chart.js in your `<head>`:
```html
<script src="https://cdn.tailwindcss.com"></script>
<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
```
Build 4 clean visual sections on the page:
1. **Top Control Bar:**
   * Inputs for `Server IP` (`10.0.0.2`), `Duration (s)` (`5`), `Reference Capacity (Mbps)` (`100`), a checkbox for `Bufferbloat Mode (Idle vs. Loaded)`, a dropdown for `Inject Kernel Impairment (tc/netem)`, and a **"Run Diagnostic Test"** button + **"Download JSON"** button.
2. **Diagnosis Verdict Banner:**
   * A prominent banner that turns **Emerald Green** (`healthy`), **Amber Yellow** (`warning`), or **Rose Red** (`critical`) based on `report.diagnosis.severity`.
   * Displays `report.diagnosis.condition`, `report.diagnosis.summary`, and renders `report.diagnosis.explanations` as bullet points.
3. **4 Metric Cards (With Sub-Stats):**
   * **Card 1 — Application Goodput:** `stats.goodput_mbps` Mbps (Sub-label: `Utilization: X% | Est. TCP/IP Header Overhead: 5.4%`).
   * **Card 2 — Round-Trip Time (RTT):** `stats.rtt_mean_ms` ms (Sub-label: `Min: X ms | P95: Y ms | Max: Z ms`).
   * **Card 3 — Delay Variation (Jitter):** `stats.jitter_rfc3550_ms` ms (Sub-label: `StdDev: X ms` or `Queue Inflation: +Y ms` if Bufferbloat mode is active).
   * **Card 4 — UDP Probe Loss:** `stats.probe_loss_pct` % (Sub-label: `Received: X / Y Probes`).
4. **2 Side-by-Side Chart.js Canvases:**
   * **Left Chart (`#rttChart`):** Line chart plotting `raw.udp_samples` (`seq_num` on X-axis, `rtt_ms` on Y-axis). Color `"idle"` points blue, `"loaded"` points orange, and `"lost"` packets red!
   * **Right Chart (`#tcpChart`):** Line/Area chart plotting `raw.tcp_intervals` (`elapsed_sec` on X-axis, `goodput_mbps` on Y-axis).

---

## Part 6: Sunday Afternoon — How to Connect & Run Everything

Once all three of you merge your branches (`feat/netengine`, `feat/analyzer-testbed`, `feat/webui`) into `main`:

### 1. Connect the 3 Modules in `pkg/web/api.go`
Inside your `/api/run` handler in `pkg/web/api.go`, replace `models.GetMockReport()` with the real pipeline:
```go
var cfg models.TestConfig
json.NewDecoder(r.Body).Decode(&cfg)

// 1. Optional: Apply tc/netem preset if selected in UI
if cfg.ImpairmentPreset != "" {
    exec.Command("bash", "scripts/impair.sh", cfg.ImpairmentPreset).Run()
}

// 2. Ayaan's Engine runs the active TCP & UDP socket tests
rawOutput, err := netengine.RunClientTest(cfg)
if err != nil {
    http.Error(w, err.Error(), 500)
    return
}

// 3. Mustafa's Engine computes stats and diagnoses the network condition
stats := analyzer.CalculateStats(cfg, rawOutput)
diagnosis := analyzer.Diagnose(cfg, stats)

// 4. Send FinalReport back to Zaynab's UI
report := models.FinalReport{
    Timestamp: time.Now().Format("15:04:05"),
    Config:    cfg,
    Raw:       rawOutput,
    Stats:     stats,
    Diagnosis: diagnosis,
}
json.NewEncoder(w).Encode(report)
```

### 2. Run the Full Project in 3 Terminal Commands
Open your Linux terminal in the `bifrost/` folder:

```bash
# 1. Build the single Go binary
go build -o bifrost main.go

# 2. Set up the virtual Linux network namespaces (client_ns and server_ns)
sudo ./scripts/setup_netns.sh

# 3. Start the Bifrost Server inside server_ns (10.0.0.2)
sudo ip netns exec server_ns ./bifrost --mode=server &

# 4. Start the Bifrost Client + Web UI inside client_ns (10.0.0.1)
sudo ip netns exec client_ns ./bifrost --mode=client --port=8080
```
*(Tip: To open your browser inside or bridged to `client_ns`, or even simpler: run the Web UI on the host or use `sudo ip netns exec client_ns google-chrome --no-sandbox http://localhost:8080` or forward port 8080 from host to `client_ns`!)*

---

## Summary Checklist for the Weekend

- [ ] **Friday Night (All):** Create repo, paste `pkg/models/types.go`, create 3 branches.
- [ ] **Saturday Morning:**
  - **Ayaan:** UDP Probe Client & Echo Server (`:9002`).
  - **Mustafa:** `CalculateStats()` in `pkg/analyzer/stats.go`.
  - **Zaynab:** Go HTTP Server (`:8080`) serving `GetMockReport()`.
- [ ] **Saturday Afternoon:**
  - **Ayaan:** TCP Goodput Sender & Receiver (`:9001`).
  - **Mustafa:** `Diagnose()` 5-rule classifier in `pkg/analyzer/diagnosis.go`.
  - **Zaynab:** Dashboard HTML layout + KPI Cards + Verdict Banner.
- [ ] **Sunday Morning:**
  - **Ayaan:** TCP Control Channel (`:9000`) + Bufferbloat `sync.WaitGroup` concurrency.
  - **Mustafa:** `scripts/setup_netns.sh` and `scripts/impair.sh`.
  - **Zaynab:** Chart.js RTT & Goodput graphs + Download JSON button.
- [ ] **Sunday Afternoon (All):** Merge to `main`, connect the 3 functions in `api.go`, test all 6 `impair.sh` scenarios, and export your JSON reports!