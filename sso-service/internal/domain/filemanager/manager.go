package filemanager

type FileManager interface {
	LoadFile(path string) ([]byte, error)
}
