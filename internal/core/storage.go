package core

// Storage defines the interface for snippet persistence operations.
type Storage interface {
	ListFiles() ([]string, error)
	ReadFile(filepath string) ([]byte, error)
	WriteFile(filepath string, data []byte) error
	DeleteFile(filepath string) error
	FileExists(filepath string) bool
	GetSnippetsDir() string
}
