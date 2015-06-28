package main

import (
	"flag"
	"fmt"
	"github.com/janelia-flyem/DVIDServicesServer/Server"
	"os"
)

const defaultPort = 15000

var (
	portNum  = flag.Int("port", defaultPort, "")
	showHelp = flag.Bool("help", false, "")
        configFile = flag.String("config", "", "")
)

const helpMessage = `
Launches service manager for Spark-based EM services.

Usage: DVIDServicesServer <config-file>
  Provide config file for remote cluster access (otherwise local machine can access the cluster) and web front-end
      -port     (number)        Port for HTTP server
  -h, -help     (flag)          Show help message
`

func main() {
	flag.BoolVar(showHelp, "h", false, "Show help message")
	flag.Parse()

	if *showHelp {
		fmt.Printf(helpMessage)
		os.Exit(0)
	}

	if flag.NArg() != 1 {
                fmt.Println("Must provide a config file for web front-end and location of DVIDServices")
                fmt.Println(helpMessage)
                os.Exit(0)
        }

	Server.Serve(*portNum, flag.Arg(0))
}
