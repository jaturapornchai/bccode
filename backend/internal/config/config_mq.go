package config

// IMQConfig is mq configuration interface
type IMQConfig interface {
	URI() string
	SecurityProtocol() string
	SSLKeyFile() string
	SSLCAFile() string
	SSLCertFile() string
}

// MQ and Producer configuration
type MQConfig struct {
	uri      string
	protocol string
	sslca    string
	sslkey   string
	sslcert  string
}

func NewMQConfig() *MQConfig {

	uri := getEnv("KAFKA_SERVER_URL", "")             // localhost:9094
	protocol := getEnv("KAFKA_SECURITY_PROTOCOL", "") // SASL_SSL
	sslca := getEnv("KAFKA_SSL_CA_FILE", "")          // /path/to/ca
	sslkey := getEnv("KAFKA_SSL_KEY_FILE", "")        // /path/to/key
	sslcert := getEnv("KAFKA_SSL_CERT_FILE", "")      // /path/to/cert

	return &MQConfig{
		uri:      uri,
		protocol: protocol,
		sslca:    sslca,
		sslkey:   sslkey,
		sslcert:  sslcert,
	}
}

func (c *MQConfig) URI() string {
	return c.uri
}

func (c *MQConfig) SecurityProtocol() string {
	return c.protocol
}

func (c *MQConfig) SSLKeyFile() string {
	return c.sslkey
}

func (c *MQConfig) SSLCAFile() string {
	return c.sslca
}

func (c *MQConfig) SSLCertFile() string {
	return c.sslcert
}

func (*Config) MQConfig() IMQConfig {
	return NewMQConfig()
}
