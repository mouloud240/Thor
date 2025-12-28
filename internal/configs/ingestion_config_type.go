package configs

import "fmt"
type IngestionConfig struct{
	Server serverConfig  `yaml:"server"`
	Pipeline pipeLineConfig `yaml:"pipeline"`
	Storage storageConfig    `yaml:"storage"`
}
type serverConfig struct {
	TcpPort int `yaml:"tcp_port"`
}
type pipeLineConfig struct {
	NumWorkers int `yaml:"parser_workers"`
}
type storageConfig struct {
	LogDir string `yaml:"log_directory"`
	SegmentSize float64 `yaml:"segment_size"` //In bytes
}
func (c *IngestionConfig) String() string {
	return fmt.Sprintf(
		"IngestionConfig{\n"+
			"  Server: { TcpPort: %d },\n"+
			"  Pipeline: { NumWorkers: %d },\n"+
			"  Storage: { LogDir: %s, SegmentSize: %.0f bytes },\n"+
			"}",
		c.Server.TcpPort,
		c.Pipeline.NumWorkers,
		c.Storage.LogDir,
		c.Storage.SegmentSize,
	)
}

