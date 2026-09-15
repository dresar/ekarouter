package headroom

import (
	"runtime"
	"time"
)

type Stats struct {
	CPUCores    int     `json:"cpu_cores"`
	GoRoutines  int     `json:"goroutines"`
	HeapAlloc   uint64  `json:"heap_alloc_bytes"`
	HeapSys     uint64  `json:"heap_sys_bytes"`
	HeapFree    uint64  `json:"heap_free_bytes"`
	MemUsagePct float64 `json:"mem_usage_pct"`
	GCPauseNs   uint64  `json:"gc_pause_ns"`
	Timestamp   int64   `json:"timestamp"`
}

func Collect() *Stats {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	heapFree := ms.HeapSys - ms.HeapAlloc
	usagePct := 0.0
	if ms.HeapSys > 0 {
		usagePct = float64(ms.HeapAlloc) / float64(ms.HeapSys) * 100
	}
	return &Stats{
		CPUCores:    runtime.NumCPU(),
		GoRoutines:  runtime.NumGoroutine(),
		HeapAlloc:   ms.HeapAlloc,
		HeapSys:     ms.HeapSys,
		HeapFree:    heapFree,
		MemUsagePct: usagePct,
		GCPauseNs:   ms.PauseNs[(ms.NumGC+255)%256],
		Timestamp:   time.Now().Unix(),
	}
}

func IsConstrained() bool {
	s := Collect()
	return s.MemUsagePct > 90 || s.GoRoutines > 10000
}
