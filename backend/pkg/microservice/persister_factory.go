package microservice

// NewFilePersister returns the default file persister (Cloudflare R2)
func NewFilePersister() IPersisterFile {
	return NewPersisterR2()
}
