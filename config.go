package prometheus

import (
	configjson "github.com/akademic/go-config-json"
)

type Config struct {
	ProjectName  string
	DumpInterval configjson.Duration
	DumpPath     string
}
