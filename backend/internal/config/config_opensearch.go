package config

type IPersisterOpenSearchConfig interface {
	Address() []string
	Username() string
	Password() string
}

type PersisterOpenSearchConfig struct{}

func NewPersisterOpenSearchConfig() *PersisterOpenSearchConfig {
	return &PersisterOpenSearchConfig{}
}

func (c *PersisterOpenSearchConfig) Address() []string {

	return []string{
		getEnv("OPEN_SEARCH_ADDRESS", "http://192.168.2.204:9200"),
	}
}

func (c *PersisterOpenSearchConfig) Username() string {
	return getEnv("OPEN_SEARCH_USERNAME", "")
}

func (c *PersisterOpenSearchConfig) Password() string {
	return getEnv("OPEN_SEARCH_PASSWORD", "")
}
