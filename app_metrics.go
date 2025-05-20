package prometheus

import (
	"context"
	"runtime"
	"time"
)

type AppMetrics struct {
	prom *Prometheus
	ctx  context.Context
	stop context.CancelFunc

	GoRuntimeGoroutines *Counter
	GoMemTotalAlloc     *Counter
	GoMemMalloc         *Counter
	GoMemFrees          *Counter
	GoMemHeapAlloc      *Counter
	GoMemHeapSys        *Counter
	GoMemHeapObjects    *Counter
	GoGCPauseTotal      *Counter
	GoGCNum             *Counter
	GoGCFraction        *Counter
}

func NewAppMetrics(prom *Prometheus, ctx context.Context) *AppMetrics {
	if prom == nil {
		panic("prometheus is nil")
	}

	appCtx, cancel := context.WithCancel(ctx)

	return &AppMetrics{
		prom: prom,
		ctx:  appCtx,
		stop: cancel,

		GoRuntimeGoroutines: prom.NewCounter("go_runtime_goroutines_count", "Count of goroutines", "gauge", nil),
		GoMemTotalAlloc:     prom.NewCounter("go_mem_alloc_total", "Count of total allocations in bytes", "counter", nil),
		GoMemMalloc:         prom.NewCounter("go_mem_malloc", "Count of heap objects allocated", "counter", nil),
		GoMemFrees:          prom.NewCounter("go_mem_frees", "Count of freed objects", "counter", nil),
		GoMemHeapAlloc:      prom.NewCounter("go_mem_heap_alloc", "Count bytes of allocated heap objects", "gauge", nil),
		GoMemHeapSys:        prom.NewCounter("go_mem_heap_sys", "bytes of heap memory obtained from the OS", "gauge", nil),
		GoMemHeapObjects:    prom.NewCounter("go_mem_heap_objects", "Count of objects in heap", "gauge", nil),
		GoGCPauseTotal:      prom.NewCounter("go_gc_pause_total", "Total GC pause in ns", "counter", nil),
		GoGCNum:             prom.NewCounter("go_gc_num", "The number of completed GC cycles", "counter", nil),
		GoGCFraction:        prom.NewCounter("go_gc_fraction", "The fraction of this program's available CPU time used by the GC since the program started", "gauge", nil),
	}
}

func (am *AppMetrics) Start(dumpPeriod time.Duration) {
	runtimestats_ticker := time.NewTicker(dumpPeriod)

	for {
		select {
		case <-am.ctx.Done():
			runtimestats_ticker.Stop()
			return
		case <-runtimestats_ticker.C:
			am.Collect()
		}
	}
}

func (am *AppMetrics) Stop() {
	am.stop()
}

func (am *AppMetrics) Collect() {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)

	am.GoRuntimeGoroutines.Set(uint64(runtime.NumGoroutine()))
	am.GoMemTotalAlloc.Set(m.TotalAlloc)
	am.GoMemMalloc.Set(m.Mallocs)
	am.GoMemFrees.Set(m.Frees)
	am.GoMemHeapAlloc.Set(m.HeapAlloc)
	am.GoMemHeapSys.Set(m.HeapSys)
	am.GoMemHeapObjects.Set(m.HeapObjects)
	am.GoGCPauseTotal.Set(m.PauseTotalNs)
	am.GoGCNum.Set(uint64(m.NumGC))
	am.GoGCFraction.Set(uint64(m.GCCPUFraction * 100))
}
