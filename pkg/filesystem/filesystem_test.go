package filesystem

import (
	// "path/filepath"
	"testing"
)

// func MockFilePathWalker(dir string, fs filepath.WalkFunc) error {
// 	dirs := []string{
// 		"/topdir/subdir/file1.jpg",
// 		"/topdir/subdir/file2.jpg",
// 		"/topdir/subdir/file3.jpg",
// 	}

//	    for _, dir := range dirs {
//	        fs(dir, )
//	    }
//	}
func TestBuiltInPathReader(t *testing.T) {
	r := NewBuiltInPathReader()

	ch := make(chan string, 10)
	errCh := make(chan error)

	go r.FetchWithChannel("/Users/nayan/Sites/projects/experiments/go/geoexif/images", ch, errCh)

	fileCount := 0

L:
	for {
		select {
		case <-ch:
			fileCount++
		case e := <-errCh:
			t.Errorf("Halted due to error: %v", e)
			break L
		}
	}

	if fileCount != 9 {
		t.Errorf("File count should be %v, got %v", 9, fileCount)
	}

}
