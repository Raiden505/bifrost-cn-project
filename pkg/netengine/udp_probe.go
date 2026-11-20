package netengine

import (
	"bifrost/pkg/models"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func RunUDPProbe(serverIP string, durationSec int, intervalMs int, phase string) []models.RawProbeSample {
	conn, err := net.Dial("udp", serverIP+":9002")
	if err != nil {
		fmt.Println("An error while connecting")
		return nil
	}

	defer conn.Close()

	var samples []models.RawProbeSample
	packet := make([]byte, 64)
	recvBuf := make([]byte, 64)

	var seqNum uint64 = 0
	endTime := time.Now().Add(time.Duration(durationSec) * time.Second)
	for time.Now().Before(endTime) {
		seqNum++
		sentNs := time.Now().UnixNano()
		binary.BigEndian.PutUint64(packet[0:8], seqNum)
		binary.BigEndian.PutUint64(packet[8:16], uint64(sentNs))

		conn.Write(packet)
		conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		_, err := conn.Read(recvBuf)
		recvNs := time.Now().UnixNano()
		if err != nil {
			samples = append(samples, models.RawProbeSample{SeqNum: seqNum, Lost: true, Phase: phase})
		} else {
			samples = append(samples, models.RawProbeSample{SeqNum: seqNum, SentTime: sentNs, RecvTime: recvNs, RTTMs: float64(recvNs-sentNs) / 1e6, Phase: phase})
		}

		time.Sleep(time.Duration(intervalMs) * time.Millisecond)
	}
	return samples
}
