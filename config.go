package prometheus

import (
	"time"

	configjson "github.com/akademic/go-config-json"
)

type Config struct {
	ProjectName  string
	DumpInterval configjson.Duration
	DumpPath     string
}

var DefaultDumpInterval = configjson.Duration{Duration: 60 * time.Second}
