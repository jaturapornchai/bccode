package config

type IPersisterElkConfig interface {
	ElkAddress() []string
	Username() string
	Password() string
}

type PersisterElkConfig struct{}

func NewPersisterElkConfig() *PersisterElkConfig {
	return &PersisterElkConfig{}
}

func (c *PersisterElkConfig) ElkAddress() []string {

	return []string{
		getEnv("ELK_ADDRESS", "http://192.168.2.204:9200"),
	}
}

func (c *PersisterElkConfig) Username() string {
	return getEnv("ELK_USERNAME", "elastic")
}

func (c *PersisterElkConfig) Password() string {
	return getEnv("ELK_PASSWORD", "smlSoft2021")
}
