package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/pflag"
)

var (
	// todo implement textMatrix/htmlMatrix/mergeOuput
	// textMatrix  = pflag.Bool("summary-matrix", false, "Render multiple result files to the console")
	// htmlMatrix  = pflag.String("summary-matrix", "", "Generate an HTML report matrix")
	// mergeOuput  = pflag.String("merge", "", "Merge multiple test results into one file")
	maxFailures = pflag.Int("max-failures", 0, "Exit non-zero if FAILURES or more test cases are failures (has no effect with --merge)")
	maxSkipped  = pflag.Int("max-skipped", 0, "Exit non-zero if SKIPPED or more test cases are skipped (has no effect with --merged)")
)

func Usage() {
	fmt.Printf("Usage: %s [OPTIONS] INPUT_FILE OUTPUT_FILE \n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	pflag.Parse()
	pflag.Usage = Usage
	if pflag.NArg() != 2 {
		pflag.Usage()
		os.Exit(1)
	}
	inputFile, outputFile := pflag.Arg(0), pflag.Arg(1)
	err := run(inputFile, outputFile)
	if err != nil {
		fmt.Println("convert failed: ", err)
	} else {
		fmt.Println("convert success")
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
