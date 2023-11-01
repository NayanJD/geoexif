package main

import (
	"flag"
	"geoexif/pkg/extractor"
	"geoexif/pkg/filesystem"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
)

var fs = flag.NewFlagSet("geoexif", flag.ExitOnError)

var dir = fs.String("dir", ".", "directory on which to extract the exif data of images. Default: the current directory")

var outputFileName = fs.String("output-file", "output.csv", "the path to output file. Default: output.csv")

var htmlFormat = fs.Bool("html", false, "Specifies if the file format of output is HTML. Default is CSV.")

func main() {
	fs.Parse(os.Args[1:])

	pathReader := filesystem.NewBuiltInPathReader()

	// Walk down directories and sub directories to get a slice
	// of all file paths
	paths, err := pathReader.Fetch(*dir)

	if err != nil {
		log.Fatalf("Error while walking the directory: %v", err)
	}

	// This channel would be used to signal all running go routines
	// of extractor (line 44) to start graceful shutdown
	done := make(chan bool, 1)

	// The waitgroup to wait for graceful shutdown of the go routines
	// Primarilty we need to wait for the extractor go routines to
	// close its ResultChan
	gracefulShutdownWg := &sync.WaitGroup{}
	gracefulShutdownWg.Add(1)

	// The struct whose Run() would actually start reading from paths
	// slice and send the image exif data to ResultChan read only channel
	extractor := extractor.NewExtractor(paths, runtime.NumCPU(), done, gracefulShutdownWg)
	go extractor.Run()

	// The channel which would signal that the CSV or HTML file writer has completed
	// successfully or handled inturruption gracefully
	fileWriterChan := make(chan bool, 1)

	// The file writer interface for CSV or HTML format
	var resultWriter filesystem.ResultWriter

	if *htmlFormat {
		resultWriter, err = filesystem.NewResultWriter(*outputFileName, filesystem.TypeHTMLWriter)

	} else {
		resultWriter, err = filesystem.NewResultWriter(*outputFileName, filesystem.TypeCSVWriter)

	}
	if err != nil {
		log.Fatalf("Error while opening %v: %v", *outputFileName, err)
	}

	resultWriter.WriteHeader()

	// Kick off the go routine to read from extractor's ResultChan and
	// write to the file
	go func() {

		for r := range extractor.ResultChan {
			// log.Printf("main: Got result: %#v\n", r)
			resultWriter.WriteResult(r)
		}

		// TODO: Handle errors while Sync or Close
		resultWriter.Sync()
		resultWriter.Close()
		fileWriterChan <- true

	}()

	// This is to capture values process interruption
	exit := make(chan os.Signal, 1)

	// Capture OS interruption
	go signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	// Hanlde Graceful Shutdown
	for {
		select {
		case <-exit:
			// Hanlde exit here. Closing done would signal all extractor's
			// go routines that it should stop further work and return immediately
			close(done)

			log.Println("Received shudown. Waiting for graceful shutdown")
			gracefulShutdownWg.Wait()
			log.Println("Gracefully shutted down")
			// break L2
		case <-fileWriterChan:
			// Once the file writer go-routine has fished gracefully, exit with
			// 0 code
			log.Println("File closed Successfully")
			os.Exit(0)
		}
	}
}
