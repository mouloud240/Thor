package storage

import (
	"os"
	"testing"
)

func BenchmarkEngine_StoreLog(b *testing.B) {
	os.RemoveAll("test_dump")
	engine, err := NewEngine("test_dump", 1024*1024*100, 50)
	if err != nil {
		b.Fatal(err)
	}
	engine.StartWorkers()

	logLine := "2026-03-13T23:36:46Z|INFO|api-server|Test log message\n"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := engine.StoreLog(logLine)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
