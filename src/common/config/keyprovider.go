package config

import "io/ioutil"

// KeyProvider provides the key used to encrypt and decrypt attrs
type KeyProvider interface {
	// Get returns the key
	// params can be used to pass parameters in different implements
	Get(params map[string]interface{}) (string, error)
}

// FileKeyProvider reads key from file
type FileKeyProvider struct {
	path string
}

// NewFileKeyProvider returns an instance of FileKeyProvider
// path: where the key should be read from
func NewFileKeyProvider(path string) KeyProvider {
	return &FileKeyProvider{
		path: path,
	}
}

// Get returns the key read from file
func (f *FileKeyProvider) Get(params map[string]interface{}) (string, error) {
	b, err := ioutil.ReadFile(f.path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// PresetKeyProvider returns the preset key disregarding the parm, this is for testing only
type PresetKeyProvider struct {
	Key string
}

// Get ...
func (p *PresetKeyProvider) Get(params map[string]interface{}) (string, error) {
	return p.Key, nil
}
