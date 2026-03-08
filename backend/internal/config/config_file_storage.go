package config

type StorageFileConfig struct{}

func NewStorageFileConfig() *StorageFileConfig {
	return &StorageFileConfig{}
}

func (cfg *StorageFileConfig) StorageDataPath() string {
	return getEnv("STORAGE_DATA_PATH", "")
}

func (cfg *StorageFileConfig) StorageUriAtlas() string {
	return getEnv("STORAGE_DATA_URI", "")
}
