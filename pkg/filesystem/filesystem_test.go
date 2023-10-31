package filesystem

import (
	// "path/filepath"
	"testing"
)

// Need to mock the file path walker instead of using the
// path/filepath Walk function to prevent directly reading directory
// and better check edge cases

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
	t.Run("test Fetch", func(t *testing.T) {
		r := NewBuiltInPathReader()

		paths, err := r.Fetch("/Users/nayan/Sites/projects/experiments/go/geoexif/images")

		if err != nil {
			t.Errorf("Got error: %v", err)
		}

		if len(paths) != 9 {
			t.Errorf("Length of the slice should b %v, got %v", 9, len(paths))
		}
	})

	t.Run("test FetchWithChannel", func(t *testing.T) {
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
			case e, ok := <-errCh:
				if ok {
					t.Errorf("Halted due to error: %v", e)
				}
				break L
			}
		}

		if fileCount != 9 {
			t.Errorf("File count should be %v, got %v", 9, fileCount)
		}

	})

}
