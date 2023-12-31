package extractor

import (
	// "fmt"
	// "bufio"
	"geoexif/pkg/image"
	// "log"
	// "os"
	"sync"
	// "time"
)

type Extractor interface {
	Run()
	GetResultChan() <-chan *ExtractedResult
}

// Extractor represents a type with a fixed file paths
// slice. By fixed, we mean that to extract image EXIF tags
// of another slice you need to create another instance of this
// struct.
type SimpleExtractor struct {

	// The number of go routines this extractor is supposed to
	// spawn. Anything less than 1 would set this to 1.
	concurrencyN int

	// The set of image paths whose EXIF data needs to be extracted.
	paths []string

	// The channel which would stop the execution of this extractor
	done <-chan bool

	resultChan             chan *ExtractedResult
	gracefulCloseWaitGroup *sync.WaitGroup

	shouldStopNow bool
}

type ExtractedResult struct {

	// Path of the image
	ImagePath string

	// Geo data in String if present
	Data string

	// Error faced while extracting EXIF data
	Error error
}

func NewExtractor(paths []string, concurrency int, done <-chan bool, gracefulWg *sync.WaitGroup, p *sync.Pool) Extractor {
	resultChan := make(chan *ExtractedResult)

	if p == nil {
		return &SimpleExtractor{
			concurrencyN:           concurrency,
			paths:                  paths,
			resultChan:             resultChan,
			done:                   done,
			gracefulCloseWaitGroup: gracefulWg,
		}
	} else {
		return &ExtractorWithBuffferPool{
			SimpleExtractor: SimpleExtractor{
				concurrencyN:           concurrency,
				paths:                  paths,
				resultChan:             resultChan,
				done:                   done,
				gracefulCloseWaitGroup: gracefulWg,
			},
			bufferPool: p,
		}
	}

}

func (e *SimpleExtractor) Run() {

	// fmt.Printf("len paths: %v\n", len(e.paths))

	defer e.gracefulCloseWaitGroup.Done()
	defer close(e.resultChan)

	concurrency := e.concurrencyN

	if concurrency < 1 {
		concurrency = 1
	}

	if concurrency > len(e.paths) {
		concurrency = Max(len(e.paths), 1)
	}

	steps := len(e.paths) / concurrency

	times := concurrency + len(e.paths)%concurrency

	wg := sync.WaitGroup{}

	for i := 0; i < times; i++ {
		wg.Add(1)

		go e.ExtractExif(steps*i, Min(steps*(i+1), len(e.paths)), &wg)
	}

	wg.Wait()

	// fmt.Println("Extractor run dnoe")
}

func (e *SimpleExtractor) ExtractExif(from, to int, wg *sync.WaitGroup) {
	defer wg.Done()

	isDoneCh := make(chan interface{})

	go e.ExtractExifHelper(from, to, isDoneCh)
L:
	for {
		select {
		case <-e.done:
			// fmt.Printf("Got shutdown request")
			e.shouldStopNow = true
		case <-isDoneCh:
			break L
		}
	}
}

func (e *SimpleExtractor) ExtractExifHelper(from, to int, isDoneCh chan<- interface{}) {
	defer close(isDoneCh)

	for ; from < to; from++ {

		// log.Printf("Extracting exif: %v\n", e.paths[from])

		data, err := image.GetGeoData(e.paths[from])

		if e.shouldStopNow {
			return
		}

		e.resultChan <- &ExtractedResult{ImagePath: e.paths[from], Data: data, Error: err}
	}

}

func (e *SimpleExtractor) GetResultChan() <-chan *ExtractedResult {
	return e.resultChan
}

func Min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func Max(a, b int) int {
	if a < b {
		return b
	} else {
		return a
	}
}
