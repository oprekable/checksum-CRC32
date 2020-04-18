//go:generate go run generator.go
// From https://github.com/koddr/example-embed-static-files-go/blob/master/internal/box/box.go

package helper

type embedBox struct {
	storage map[string][]byte
}

// Create new box for embed files
func newEmbedBox() *embedBox {
	return &embedBox{storage: make(map[string][]byte)}
}

// Add a file to box
func (e *embedBox) Add(file string, content []byte) {
	e.storage[file] = content
}

// Get file's content
// Always use / for looking up
// For example: /init/README.md is actually configs/init/README.md
func (e *embedBox) Get(file string) []byte {
	if f, ok := e.storage[file]; ok {
		return f
	}
	return nil
}

// Find for a file
func (e *embedBox) Has(file string) bool {
	if _, ok := e.storage[file]; ok {
		return true
	}
	return false
}

// Box Embed box expose
var Box = newEmbedBox()

// Add a file content to box
func Add(file string, content []byte) {
	Box.Add(file, content)
}

// Get a file from box
func Get(file string) []byte {
	return Box.Get(file)
}

// Has a file in box
func Has(file string) bool {
	return Box.Has(file)
}
