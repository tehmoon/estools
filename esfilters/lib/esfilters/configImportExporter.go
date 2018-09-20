package esfilters

type ConfigImportExporter interface {
	ExportConfig() ([]byte, error)
	ImportConfig([]byte) (error)
	RLock()
	Runlock()
	Lock()
	Unlock()
}
