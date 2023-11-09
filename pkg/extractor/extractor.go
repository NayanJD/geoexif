package extractor

import (
	// "fmt"
	// "bufio"
	"geoexif/pkg/image"
	// "log"
	"os"
	"sync"
	// "time"
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

	bufferPool *sync.Pool
}

type ExtractedResult struct {

	// Path of the image
	ImagePath string

	// Geo data in String if present
	Data string

	// Error faced while extracting EXIF data
	Error error
}

func NewExtractor(paths []string, concurrency int, done <-chan bool, gracefulWg *sync.WaitGroup, p *sync.Pool) *Extractor {
	resultChan := make(chan *ExtractedResult)

	return &Extractor{
		concurrencyN:           concurrency,
		paths:                  paths,
		ResultChan:             resultChan,
		resultChan:             resultChan,
		done:                   done,
		gracefulCloseWaitGroup: gracefulWg,
		bufferPool:             p,
	}
}

func (e *Extractor) Run() {

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

	// fileCloserCh := make(chan *os.File)

	// defer close(fileCloserCh)

	// go func() {

	// 	for file := range fileCloserCh {
	// 		file.Close()
	// 	}
	// }()

	for ; from < to; from++ {

		// log.Printf("Extracting exif: %v\n", e.paths[from])

		// fileCloserCh <- file

		// // if err != nil {
		// // 	log.Printf("path: %q, err: %v\n", e.paths[from], err)
		// // }

		// data, err := image.GetGeoData(e.paths[from])
		if e.shouldStopNow {
			return
		}
		e.resultChan <- e.GetExifData(e.paths[from])
	}

}

func (e *Extractor) GetExifData(path string) *ExtractedResult {
	file, err := os.Open(path)
	if err != nil {
		return &ExtractedResult{ImagePath: path, Error: err}
	}

	defer file.Close()

	fileInfo, err := file.Stat()

	if err != nil {
		return &ExtractedResult{ImagePath: path, Error: err}
	}

	sizeInBytes := fileInfo.Size()

	imageBytesBufferPtr := e.bufferPool.Get().(*[]byte)
	imageBytesBuffer := *imageBytesBufferPtr

	if int64(cap(imageBytesBuffer)) < sizeInBytes {
		extraByteSlice := make([]byte, int(sizeInBytes-int64(cap(imageBytesBuffer))))
		imageBytesBuffer = append(imageBytesBuffer, extraByteSlice...)
	} else {
		imageBytesBuffer = imageBytesBuffer[:sizeInBytes]
	}

	_, err = file.Read(imageBytesBuffer)
	// if err != nil {
	// 	log.Printf("path: %q, err: %v\n", e.paths[from], err)
	// }
	// log.Printf("size: %d, read %d\n", sizeInBytes, count)

	data, err := image.GetGeoDataFromBytes(imageBytesBuffer)

	*imageBytesBufferPtr = imageBytesBuffer
	e.bufferPool.Put(imageBytesBufferPtr)

	return &ExtractedResult{ImagePath: path, Data: data, Error: err}
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
