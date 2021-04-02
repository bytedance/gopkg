package gopool

const (
	defaultScalaThreshold = 1
)

type Config struct {
	// 扩容的阈值，当 len(task chan) > ScaleThreshold 时会新建 goroutine，默认为 defaultScalaThreshold
	ScaleThreshold int32
}

func NewConfig() *Config {
	c := &Config{
		ScaleThreshold: defaultScalaThreshold,
	}
	return c
}
