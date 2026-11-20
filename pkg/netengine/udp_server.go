package netengine

import (
	"fmt"
	"net"
)

func StartUDPEchoServer() {
	addr, _ := net.ResolveUDPAddr("udp", ":9002")
	conn, err := net.ListenUDP("udp", addr)

	if err != nil {
		fmt.Println("Error starting UDP server: ", err)
		return
	}

	defer conn.Close()

	buf := make([]byte, 2048)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err == nil {
			conn.WriteToUDP(buf[:n], clientAddr)
		}
	}
}
