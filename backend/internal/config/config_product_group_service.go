package config

import "os"

type IProductGroupServiceConfig interface {
	ProductHost() string
}

type ProductGroupServiceConfig struct{}

func NewProductGroupServiceConfig() *ProductGroupServiceConfig {
	return &ProductGroupServiceConfig{}
}

func (cfg *ProductGroupServiceConfig) ProductHost() string {
	return os.Getenv("PRODUCT_GROUP_SERVICE_PRODUCT_HOST")
}
