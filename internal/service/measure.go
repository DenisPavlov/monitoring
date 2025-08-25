// Package metrics provides functionality for collecting and processing
// system metrics including memory statistics, CPU utilization, and custom counters.
package metrics

import (
	"math/rand"
	"runtime"

	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

// Gauge collects and returns a comprehensive set of Go runtime memory statistics
// and a random value as gauge metrics.
//
// The function retrieves memory allocation data from runtime.MemStats and
// includes a random value for sampling purposes.
//
// Returns:
//   - map[string]float64: Map of gauge metric names to their float64 values
//
// Collected metrics include:
//   - Memory allocation statistics (Alloc, HeapAlloc, HeapSys, etc.)
//   - Garbage collection metrics (GCCPUFraction, NumGC, PauseTotalNs, etc.)
//   - Memory cache and span statistics (MCacheInuse, MSpanSys, etc.)
//   - System memory statistics (Sys, OtherSys, etc.)
//   - RandomValue: A random float64 value between 0 and 1 for sampling
//
// Example usage:
//
//	gauges := metrics.Gauge()
//	fmt.Printf("Memory allocated: %f\n", gauges["Alloc"])
//	fmt.Printf("Random value: %f\n", gauges["RandomValue"])
func Gauge() map[string]float64 {
	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)
	return map[string]float64{
		"Alloc":         float64(ms.Alloc),
		"BuckHashSys":   float64(ms.BuckHashSys),
		"Frees":         float64(ms.Frees),
		"GCCPUFraction": ms.GCCPUFraction,
		"GCSys":         float64(ms.GCSys),
		"HeapAlloc":     float64(ms.HeapAlloc),
		"HeapIdle":      float64(ms.HeapIdle),
		"HeapInuse":     float64(ms.HeapInuse),
		"HeapObjects":   float64(ms.HeapObjects),
		"HeapReleased":  float64(ms.HeapReleased),
		"HeapSys":       float64(ms.HeapSys),
		"LastGC":        float64(ms.LastGC),
		"Lookups":       float64(ms.Lookups),
		"MCacheInuse":   float64(ms.MCacheInuse),
		"MCacheSys":     float64(ms.MCacheSys),
		"MSpanInuse":    float64(ms.MSpanInuse),
		"MSpanSys":      float64(ms.MSpanSys),
		"Mallocs":       float64(ms.Mallocs),
		"NextGC":        float64(ms.NextGC),
		"NumForcedGC":   float64(ms.NumForcedGC),
		"NumGC":         float64(ms.NumGC),
		"OtherSys":      float64(ms.OtherSys),
		"PauseTotalNs":  float64(ms.PauseTotalNs),
		"StackInuse":    float64(ms.StackInuse),
		"StackSys":      float64(ms.StackSys),
		"Sys":           float64(ms.Sys),
		"TotalAlloc":    float64(ms.TotalAlloc),
		"RandomValue":   rand.Float64(),
	}
}

// AdditionalGauge collects and returns system-level metrics including
// memory information and CPU utilization using the gopsutil library.
//
// This function provides insights into system-wide resource usage beyond
// the Go runtime-specific metrics provided by Gauge().
//
// Returns:
//   - map[string]float64: Map of additional gauge metric names to their values
//
// Collected metrics include:
//   - TotalMemory: Total available system memory in bytes
//   - FreeMemory: Free system memory in bytes
//   - CPUutilization1: 1-minute load average (CPU utilization)
//
// Note: Errors from underlying system calls are ignored and zero values
// are returned for failed metrics.
//
// Example usage:
//
//	systemMetrics := metrics.AdditionalGauge()
//	fmt.Printf("Total memory: %f bytes\n", systemMetrics["TotalMemory"])
//	fmt.Printf("CPU load (1min): %f\n", systemMetrics["CPUutilization1"])
func AdditionalGauge() map[string]float64 {
	memory, _ := mem.VirtualMemory()
	avg, _ := load.Avg()

	return map[string]float64{
		"TotalMemory":     float64(memory.Total),
		"FreeMemory":      float64(memory.Free),
		"CPUutilization1": avg.Load1,
	}
}

// Count increments and returns counter metrics. Specifically, it increments
// the "PollCount" counter and returns the updated counters map.
//
// This function is used to track the number of metric collection cycles
// or other countable events in the system.
//
// Parameters:
//   - counters: Map of counter names to their current int64 values
//
// Returns:
//   - map[string]int64: Updated counters map with PollCount incremented by 1
//
// Example usage:
//
//	counters := make(map[string]int64)
//	counters["PollCount"] = 0
//
//	// After each poll
//	counters = metrics.Count(counters)
//	fmt.Printf("Poll count: %d\n", counters["PollCount"])
//
// The function modifies the input map in place and returns it for convenience.
func Count(counters map[string]int64) map[string]int64 {
	counters["PollCount"] = counters["PollCount"] + 1
	return counters
}
