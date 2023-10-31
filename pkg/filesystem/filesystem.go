package filesystem

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

// This interface is to define a common set of functionality we
// expect from a package which can walk through a directory and its
// subdirectories and return the known file paths in it without
// reading the file.
type PathReader interface {

	// This functions keeps on sending the file names in dir directory
	// to a write only channel ch until it receives an item in done channel.
	// It sends the error to the e channel.
	FetchWithChannel(dir string, ch chan<- string, e chan<- error)
}

type BuiltInPathReader struct {

	// This has been stubbed out to better write the test
	// cases with mock
	Walk func(string, filepath.WalkFunc) error
}

func (bipr *BuiltInPathReader) FetchWithChannel(dir string, ch chan<- string, e chan<- error) {

	err := bipr.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error while reading metadata of file %q: %v\n", path, err)
			return err
		}

		if !info.IsDir() {
			ch <- info.Name()
		}

		return nil
	})

	if err != nil {
		e <- err
	}

	close(e)

	return
}

func NewBuiltInPathReader() PathReader {
	return &BuiltInPathReader{
		Walk: filepath.Walk,
	}
}
