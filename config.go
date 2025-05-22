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

const DefaultDumpInterval = 60 * time.Second
