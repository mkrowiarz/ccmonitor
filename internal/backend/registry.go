package backend

import "fmt"

var backends = map[string]Backend{}

// Register adds a backend to the global registry.
func Register(b Backend) {
	backends[b.Name()] = b
}

// Get returns a registered backend by name.
func Get(name string) (Backend, error) {
	b, ok := backends[name]
	if !ok {
		return nil, fmt.Errorf("unknown backend: %q", name)
	}
	return b, nil
}

// List returns the names of all registered backends.
func List() []string {
	names := make([]string, 0, len(backends))
	for name := range backends {
		names = append(names, name)
	}
	return names
}
