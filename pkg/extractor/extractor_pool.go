package extractor

import (
	"geoexif/pkg/image"
	// "log"
	"os"
	"sync"
)

type ExtractorWithBuffferPool struct {
	SimpleExtractor

	bufferPool *sync.Pool
}

func (e *ExtractorWithBuffferPool) Run() {

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

func (e *ExtractorWithBuffferPool) ExtractExif(from, to int, wg *sync.WaitGroup) {
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

func (e *ExtractorWithBuffferPool) ExtractExifHelper(from, to int, isDoneCh chan<- interface{}) {
	defer close(isDoneCh)

	// fileCloserCh := make(chan *os.File)

	// defer close(fileCloserCh)

	// go func() {

	// 	for file := range fileCloserCh {
	// 		file.Close()
	// 	}
	// }()

	for ; from < to; from++ {
		// log.Println("With pool")

		path := e.paths[from]
		file, err := os.Open(path)
		if err != nil {
			e.resultChan <- &ExtractedResult{ImagePath: path, Error: err}
			return
		}

		defer file.Close()

		fileInfo, err := file.Stat()

		if err != nil {
			e.resultChan <- &ExtractedResult{ImagePath: path, Error: err}

			return
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

		if e.shouldStopNow {
			return
		}
		e.resultChan <- &ExtractedResult{ImagePath: path, Data: data, Error: err}
	}

}
