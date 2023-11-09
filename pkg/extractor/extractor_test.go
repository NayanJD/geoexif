package extractor

import (
	"fmt"
	// "log"
	"sync/atomic"

	// "runtime"
	"sync"
	"testing"
)

var uniqueImagePaths []string = []string{
	"../../images/anubis.jpg",
	"../../images/rock.jpg",
	"../../images/bird.jpeg",
	"../../images/anubis.jpg",
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
		if i == times-1 {
			imagePaths = append(imagePaths, uniqueImagePaths[0:size%uniqueImageCount]...)
		} else {
			imagePaths = append(imagePaths, uniqueImagePaths...)
		}
	}

	return imagePaths
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
				extractor := NewExtractor(paths, 8, doneCh, gracefulShutdownWg, p)

				go func() {
					for range extractor.ResultChan {
					}
				}()

				b.ResetTimer()

				extractor.Run()

				// log.Printf("Allocations: %d\n", allocationCount.Load())
			}

		})
	}

}
