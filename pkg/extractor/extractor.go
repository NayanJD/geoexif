package extractor

import (
	"fmt"
	"geoexif/pkg/image"
	"sync"
	"time"
)

// Extractor represents a type with a fixed file paths
// slice. By fixed, we mean that to extract image EXIF tags
// of another slice you need to create another instance of this
// struct.
type Extractor struct {

	// The number of go routines this extractor is supposed to
	// spawn. Anything less than 1 would set this to 1.
	concurrencyN int

	// The set of image paths whose EXIF data needs to be extracted.
	paths []string

	// The channel which would stop the execution of this extractor
	done <-chan bool

	// The channel to which the extractor would send the result to
	ResultChan <-chan *ExtractedResult

	resultChan             chan *ExtractedResult
	gracefulCloseWaitGroup *sync.WaitGroup

	shouldStopNow bool
}

type ExtractedResult struct {
	ImagePath string

	Data string

	Error error
}

func NewExtractor(paths []string, concurrency int, done <-chan bool, gracefulWg *sync.WaitGroup) *Extractor {
	resultChan := make(chan *ExtractedResult)

	return &Extractor{
		concurrencyN:           concurrency,
		paths:                  paths,
		ResultChan:             resultChan,
		resultChan:             resultChan,
		done:                   done,
		gracefulCloseWaitGroup: gracefulWg,
	}
}

func (e *Extractor) Run() {

	fmt.Printf("len paths: %v\n", len(e.paths))

	defer e.gracefulCloseWaitGroup.Done()
	defer close(e.resultChan)

	concurrency := e.concurrencyN

	if concurrency < 1 {
		concurrency = 1
	}

	if concurrency > len(e.paths) {
		concurrency = len(e.paths)
	}

	steps := len(e.paths) / concurrency

	times := concurrency + len(e.paths)%concurrency

	fmt.Printf("concurrency: %v, steps: %v, times: %v\n", concurrency, steps, times)

	wg := sync.WaitGroup{}

	for i := 0; i < times; i++ {
		wg.Add(1)

		go e.ExtractExif(steps*i, Min(steps*(i+1), len(e.paths)), &wg)
	}

	wg.Wait()

	fmt.Println("Extractor run dnoe")
}

func (e *Extractor) ExtractExif(from, to int, wg *sync.WaitGroup) {
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

func (e *Extractor) ExtractExifHelper(from, to int, isDoneCh chan<- interface{}) {
	defer close(isDoneCh)

	for ; from < to; from++ {

		fmt.Printf("Extracting exif: %v\n", e.paths[from])

		data, err := image.GetGeoData(e.paths[from])

		if err != nil {
			fmt.Printf("error type: %T, err: %v\n", err, err)
		}
		time.Sleep(1 * time.Second)
		if e.shouldStopNow {
			return
		}
		e.resultChan <- &ExtractedResult{ImagePath: e.paths[from], Data: data, Error: err}
	}

}
func Min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
