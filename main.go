// Main entry point for Rancher.
//
// Notes on code in Main: some of the code in runMain is copied from the copy-
// right holder, Mitchell Hashimoto (github.com/mitchellh), as I am using his
// cli package.
package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mohae/cli"
	"github.com/mohae/rancher/ranchr"
	jww "github.com/spf13/jwalterweatherman"
)

func main() {
	cpus := runtime.NumCPU()
	if cpus > 1 {
		runtime.GOMAXPROCS(runtime.NumCPU() - 1)
	}
	os.Exit(realMain())
}

func realMain() int {
	// runMain parses the Flag for glog, sets up CLI stuff for the supported
	// subcommands and runs Rancher.
	err := ranchr.SetEnv()
	if err != nil {
		fmt.Printf("An error while processing Rancher Environment variables: %s\n", err)
		return -1
	}
	// Logging setup
	ranchr.SetLogging()
	args := os.Args[1:]
	cli := &cli.CLI{
		Name:     "rancher",
		Version:  Version,
		Args:     args,
		Commands: Commands,
		HelpFunc: cli.BasicHelpFunc("rancher"),
	}
	exitCode, err := cli.Run()
	if err != nil {
		jww.ERROR.Printf("Rancher encountered an error: %s\n", err)

	}
	return exitCode
}
