package config

import (
	"strings"
)

type IHttpConfig interface {
	Port() string
	PathPrefix() string
	IgnoreLogUrls() []string
	CORS() []string
}

type HttpConfig struct{}

func NewHttpConfig() IHttpConfig {
	return &HttpConfig{}
}

func (c *HttpConfig) Port() string {

	servicePort := getEnv("SERVICE_PORT", "")
	if servicePort != "" {
		return servicePort
	}

	return getEnv("HTTP_PORT", "8080")
}

func (c *HttpConfig) PathPrefix() string {

	pathPrefix := getEnv("PATH_PREFIX", "")
	if pathPrefix != "" {
		return pathPrefix
	}
	return getEnv("HTTP_PATH_PREFIX", "")
}

func (c *HttpConfig) IgnoreLogUrls() []string {

	ignoreLogUrls := getEnv("HTTP_IGNORE_LOG_URLS", "/healthz")
	if ignoreLogUrls == "" {
		return []string{}
	}

	ignoreLists := strings.Split(ignoreLogUrls, " ")
	return ignoreLists
}

func (c *HttpConfig) CORS() []string {
	rawCORS := getEnv("HTTP_CORS", "*")
	return strings.Split(rawCORS, " ")
}
