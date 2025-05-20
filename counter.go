package prometheus

type Counter struct {
	valInt   uint64
	name     string
	help     string
	datatype string
	metadata map[string]string
}

func NewCounter(name, help, datatype string, metadata map[string]string) *Counter {
	return &Counter{
		name:     name,
		help:     help,
		datatype: datatype,
		metadata: metadata,
	}
}

func (c *Counter) Inc() {
	c.valInt++
}

func (c *Counter) IncC(count uint64) {
	c.valInt += count
}

func (c *Counter) Set(value uint64) {
	c.valInt = value
}
