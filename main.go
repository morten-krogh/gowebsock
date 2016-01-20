package main

import (
	"fmt"
	"os"
)

func printUsage() {
	usageFormatString := "Usage\n  %s client <web-socket-uri>\n  %s server <port>\n"
	fmt.Printf(usageFormatString, os.Args[0], os.Args[0])
}

func main() {

	if len(os.Args) != 3 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "client":
		wsUri := os.Args[2]
		client(wsUri)
	case "server":
		port := os.Args[2]
		server(port)
	default:
		printUsage()
		os.Exit(1)
	}
}