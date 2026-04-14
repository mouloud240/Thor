package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

func main() {
	connections := 100
	messagesPerConn := 1000
	totalMessages := connections * messagesPerConn
	
	address := "localhost:9000"
	
	fmt.Printf("Starting persistent load test...\nConnections: %d\nMessages/Conn: %d\nTotal: %d\n", connections, messagesPerConn, totalMessages)
	
	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(connections)
	
	for i := 0; i < connections; i++ {
		go func(connID int) {
			defer wg.Done()
			conn, err := net.Dial("tcp", address)
			if err != nil {
				fmt.Printf("Dial error: %v\n", err)
				return
			}
			defer conn.Close()
			
			for j := 0; j < messagesPerConn; j++ {
				msg := fmt.Sprintf("2026-03-13T23:36:46Z|INFO|api-server|Test persistent log message %d from conn %d\n", j, connID)
				_, err := conn.Write([]byte(msg))
				if err != nil {
					fmt.Printf("Write error: %v\n", err)
					return
				}
				
				// Read response (basic reading to prevent buffer full)
				buf := make([]byte, 1024)
				_, _ = conn.Read(buf)
			}
		}(i)
	}
	
	wg.Wait()
	duration := time.Since(start)
	qps := float64(totalMessages) / duration.Seconds()
	
	fmt.Printf("=== Persistent Test Complete ===\nDuration: %v\nMessages sent: %d\nThroughput: %.2f messages/sec\n", duration, totalMessages, qps)
}
