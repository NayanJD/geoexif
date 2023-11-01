package main

import (
	"flag"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	// "fmt"
	"geoexif/pkg/extractor"
	"geoexif/pkg/filesystem"
	"log"
	"os"
)

var fs = flag.NewFlagSet("geoexif", flag.ExitOnError)

var dir = fs.String("dir", ".", "directory on which to extract the exif data of images. Default: the current directory")
var outputFileName = fs.String("output-file", "output.csv", "the path to output file. Default: output.csv")
var htmlFormat = fs.Bool("html", false, "Specifies if the file format of output is HTML. Default is CSV.")

func main() {
	fs.Parse(os.Args[1:])

	pathReader := filesystem.NewBuiltInPathReader()

	paths, err := pathReader.Fetch(*dir)

	if err != nil {
		log.Fatalf("Error while walking the directory: %v", err)
	}

	done := make(chan bool, 1)

	gracefulShutdownWg := &sync.WaitGroup{}

	gracefulShutdownWg.Add(1)

	extractor := extractor.NewExtractor(paths, runtime.NumCPU(), done, gracefulShutdownWg)

	go extractor.Run()

	exit := make(chan os.Signal, 1)

	fileWriterChan := make(chan bool, 1)

	// outputFile, err := os.OpenFile(*outputFileName, os.O_CREATE|os.O_WRONLY, 0600)

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

	go func() {

		for r := range extractor.ResultChan {
			// log.Printf("main: Got result: %#v\n", r)
			resultWriter.WriteResult(r)
		}

		resultWriter.Sync()
		resultWriter.Close()
		fileWriterChan <- true

	}()

	go signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-exit:
			close(done)

			log.Println("Waiting for graceful shutdown")
			gracefulShutdownWg.Wait()
			log.Println("Gracefully shutted down")
			// break L2
		case <-fileWriterChan:
			os.Exit(0)
		}
	}
}
