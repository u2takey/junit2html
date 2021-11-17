package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/spf13/pflag"
)

var (
	// todo implement textMatrix/htmlMatrix/mergeOuput
	// textMatrix  = pflag.Bool("summary-matrix", false, "Render multiple result files to the console")
	// htmlMatrix  = pflag.String("summary-matrix", "", "Generate an HTML report matrix")
	// mergeOuput  = pflag.String("merge", "", "Merge multiple test results into one file")
	// maxFailures = pflag.Int("max-failures", 0, "Exit non-zero if FAILURES or more test cases are failures (has no effect with --merge)")
	// maxSkipped  = pflag.Int("max-skipped", 0, "Exit non-zero if SKIPPED or more test cases are skipped (has no effect with --merged)")
	open = pflag.Bool("open", false, "Open output file after rendering")
)

func Usage() {
	fmt.Printf("Usage: %s INPUT_FILE OUTPUT_FILE \n", os.Args[0])
	fmt.Printf("       go test -v ./... | %s OUTPUT_FILE \n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	pflag.Parse()
	pflag.Usage = Usage
	var done int32 = 0
	if pflag.NArg() == 1 {
		// read go test result from stdin
		outputFile := pflag.Arg(0)
		go func() {
			err := Parse(os.Stdin, "")
			if err != nil {
				fmt.Println("Parse go test result failed: ", err)
				os.Exit(1)
			}
			atomic.AddInt32(&done, 1)
			fmt.Println("Test Done, Exiting...")
		}()

		go func() {
			if *open {
				time.Sleep(1 * time.Second)
				_ = exec.Command("open", outputFile).Run()
			}
		}()

		var indexHtml []byte
		for range time.Tick(time.Second) {
			title := "Test Ongoing"
			replace1 := `99999999`
			replace2 := `2.9999999`
			if atomic.LoadInt32(&done) > 0 {
				title = "Test Done"
				tmp := replace2
				replace2 = replace1
				replace1 = tmp
			}
			junit, err := JUnitReportXMLModel(report, "", title)
			if err != nil {
				fmt.Printf("Error writing html: %s\n", err)
				os.Exit(1)
			}
			data, err := junit.Html()
			if err != nil {
				fmt.Printf("Error writing html: %s\n", err)
				os.Exit(1)
			}
			indexHtml = bytes.ReplaceAll(data, []byte(replace1), []byte(replace2))
			f, err := os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
			if err != nil {
				fmt.Printf("Error open file: %s\n", err)
				os.Exit(1)
			}
			_, err = f.Write(indexHtml)
			if err != nil {
				fmt.Printf("Error write file: %s\n", err)
				os.Exit(1)
			}
			_ = f.Close()

			if atomic.LoadInt32(&done) > 0 {
				os.Exit(0)
			}
		}
	} else if pflag.NArg() == 2 {
		inputFile, outputFile := pflag.Arg(0), pflag.Arg(1)
		err := run(inputFile, outputFile)
		if err != nil {
			fmt.Println("convert failed: ", err)
		} else {
			fmt.Println("convert success")
		}
		if *open {
			_ = exec.Command("open", outputFile).Run()
		}
	} else {
		pflag.Usage()
		os.Exit(1)
	}
}

func run(inputFile, outputFile string) error {
	fmt.Printf("runing with input file: %v, output file: %s\n", inputFile, outputFile)
	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return err
	}
	jUnit, err := NewJunit(data, inputFile)
	if err != nil {
		return err
	}
	outData, err := jUnit.Html()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(outputFile, outData, 0777)
}
