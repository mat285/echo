package types

// File represents a name + content pair.
type File struct {
	Name     string `json:"name" yaml:"name"`
	Contents string `json:"contents" yaml:"contents"`
}
