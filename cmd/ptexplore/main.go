package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gianlucaborello/ptexplore"
)

func main() {
	flag.Usage = usage

	pid := flag.Int("pid", 0, "Pid of the process to analyze (e.g. 42)")
	areaFilter := flag.String("areas", "", "Comma separated list of memory areas (even patterns) to analyze (e.g. 'stack,heap,libc')")
	addressFilter := flag.String("address", "", "Analyze a single address (e.g. '0x7f66a002ab70')")
	quiet := flag.Bool("quiet", false, "Don't print page table details, just a summary of the memory areas")

	flag.Parse()

	if *pid == 0 {
		fmt.Fprintln(os.Stderr, "Pid not specified")
		os.Exit(1)
	}

	var address uint64
	var err error
	if *addressFilter != "" {
		address, err = getHexAddress(*addressFilter)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	state := &ptexplore.PtExplorerState{}

	err = state.ParseMemAreas(*pid)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = state.OpenSystemFiles(*pid)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = state.PrintAreas(*areaFilter, address, *quiet)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Explore page table of a process under Linux.\n")
	fmt.Fprintf(os.Stderr, "Works by attaching to a process and printing each memory area. Optionally, memory areas can be restricted via a filter.\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func getHexAddress(hex string) (address uint64, err error) {
	if strings.HasPrefix(hex, "0x") {
		hex = hex[2:]
	}

	address, err = strconv.ParseUint(hex, 16, 64)
	if err != nil {
		return 0, err
	}

	return address, nil
}
