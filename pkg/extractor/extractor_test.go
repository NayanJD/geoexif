package extractor

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	// "log"
	"sync/atomic"

	// "runtime"
	"os"
	"strings"
	"sync"
	"testing"
)

var uniqueImagePaths []string = []string{
	"../../images/anubis.jpg",
	"../../images/rock.jpg",
	"../../images/bird.jpeg",
	"../../images/more_images/dog.png",
	"../../images/more_images/wax-card.jpg",
}

func randomisedImagePaths(size int) []string {
	uniqueImageCount := len(uniqueImagePaths)

	if size <= uniqueImageCount {
		return uniqueImagePaths[0:size]
	}

	times := size / uniqueImageCount

	if size%uniqueImageCount != 0 {
		times++
	}

	imagePaths := make([]string, 0)

	for i := 0; i < times; i++ {
		var selectedPaths []string
		if i == times-1 {
			selectedPaths = uniqueImagePaths[0 : size%uniqueImageCount]

			// imagePaths = append(imagePaths, uniqueImagePaths[0:size%uniqueImageCount]...)
		} else {
			// imagePaths = append(imagePaths, uniqueImagePaths...)
			selectedPaths = uniqueImagePaths
		}

		for _, path := range selectedPaths {
			fileName := filepath.Base(path)

			splittedFileName := strings.Split(fileName, ".")
			fileNameWithoutExt := strings.Join(splittedFileName[0:len(splittedFileName)-1], "")

			newFileName := fmt.Sprintf("%v-%d.%v", fileNameWithoutExt, i, splittedFileName[len(splittedFileName)-1])

			imagePaths = append(imagePaths, filepath.Join("../../test_images", newFileName))

			input, _ := ioutil.ReadFile(path)

			ioutil.WriteFile(filepath.Join("../../test_images", newFileName), input, 0644)
		}
	}

	return imagePaths
}

func clearFiles(paths []string) {
	for _, path := range paths {
		os.Remove(path)
	}
}

func BenchmarkExtractor(b *testing.B) {

	testInputs := []int{100, 500, 1e3, 1e4}
	// testInputs := []int{100}

	for _, testInput := range testInputs {
		b.Run(fmt.Sprintf("file_count_%d", testInput), func(b *testing.B) {
			paths := randomisedImagePaths(testInput)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				doneCh := make(chan bool)
				gracefulShutdownWg := &sync.WaitGroup{}

				gracefulShutdownWg.Add(1)

				allocationCount := atomic.Int32{}

				p := &sync.Pool{
					New: func() interface{} {
						mem := make([]byte, 4*1024)

						allocationCount.Add(1)
						// log.Println("New buffer created")
						return &mem
					},
				}
				e := NewExtractor(paths, 8, doneCh, gracefulShutdownWg, p)

				go func() {
					for range e.GetResultChan() {
					}
				}()

				b.ResetTimer()

				e.Run()

				// log.Printf("Allocations: %d\n", allocationCount.Load())

				clearFiles(paths)
			}

		})
	}

}
