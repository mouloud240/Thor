package configs

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
	SegmentSize int64 `yaml:"segment_size"` //In bytes
}


