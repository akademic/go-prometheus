package prometheus

import (
	"context"
	"fmt"
	"os"
	"time"
)

type Prometheus struct {
	ctx  context.Context
	stop context.CancelFunc

	cfg            Config
	log            Logger
	counters       []*Counter
	populatedNames map[string]bool
}

func NewPrometheus(ctx context.Context, cfg Config, log Logger) *Prometheus {
	promCtx, stop := context.WithCancel(ctx)
	prom := &Prometheus{
		ctx:            promCtx,
		stop:           stop,
		cfg:            cfg,
		log:            log,
		counters:       make([]*Counter, 0),
		populatedNames: make(map[string]bool),
	}

	return prom
}

func (prom *Prometheus) NewCounter(name, help, datatype string, metadata map[string]string) *Counter {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["project"] = prom.cfg.ProjectName

	counter := NewCounter(name, help, datatype, metadata)

	prom.register(counter)

	return counter
}

func (prom *Prometheus) Populate() string {
	metrics_str := ""
	prom.populatedNames = make(map[string]bool)
	for _, item := range prom.counters {
		if _, ok := prom.populatedNames[item.name]; !ok {
			metrics_str += fmt.Sprintf("# HELP %s %s\n", item.name, item.help)
			metrics_str += fmt.Sprintf("# TYPE %s %s\n", item.name, item.datatype)
			prom.populatedNames[item.name] = true
		}
		metrics_str += fmt.Sprintf("%s", item.name)
		if len(item.metadata) != 0 {
			metrics_str += "{"
			sep := ""
			for md_key, md_val := range item.metadata {
				metrics_str += fmt.Sprintf("%s%s=\"%s\"", sep, md_key, md_val)
				sep = ", "
			}
			metrics_str += "}"
		}

		metrics_str += fmt.Sprintf(" %d\n", item.valInt)
	}

	return metrics_str
}

func (prom *Prometheus) register(counter *Counter) {
	prom.counters = append(prom.counters, counter)
}

func (prom *Prometheus) Start() {
	if prom.cfg.DumpPath == "" {
		prom.log.Info("dump path is empty, skip dump")
		return
	}

	prometheus_dump_ticker := time.NewTicker(prom.cfg.DumpInterval.Duration)

	for {
		select {
		case <-prom.ctx.Done():
			prom.log.Info("stop prometheus")
			prom.stop()
			return

		case <-prometheus_dump_ticker.C:
			prom.dumpPrometheusLog()
		}
	}
}

func (prom *Prometheus) Stop() {
	prom.log.Info("stop prometheus")
	prom.stop()
}

func (prom *Prometheus) dumpPrometheusLog() {
	if prom.cfg.DumpPath != "" {
		metrics_str := prom.Populate()

		prom.log.Info("dump prometheus file")
		err := os.WriteFile(prom.cfg.DumpPath, []byte(metrics_str), 0644)
		if err != nil {
			prom.log.Error("os.WriteFile(): %s", err)
		}
	}
}
