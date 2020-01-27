package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"application/webserver"
)

func main() {
	fmt.Printf("\nwelcome\n\n")

	var printUsage = false
	var help = flag.Bool("help", false, "show this help")
	var port = flag.Int("port", 8080, "webserver listening port")

	flag.Parse()

	if *help {
		printUsage = true
	}

	verbs := flag.Args()

	if len(verbs) == 0 {
		fmt.Println("not enough parameters, use '-help' !")
		printUsage = true
	}

	printHelp := func() {
		fmt.Printf("\nsesame usage :\n\n  sesame [OPTIONS] verbs...\n\nOPTIONS :\n\n")
		flag.PrintDefaults()
		fmt.Printf("\nVERBS :\n\n  serve\n        serves the Web UI\n")
		fmt.Printf("\nEXAMPLES :\n\n  sesame serve\n  sesame serve .\n  sesame serve ~/repos/my-git-repo\n")
	}

	if printUsage {
		printHelp()
		return
	}

	// execute the verb
	switch verbs[0] {
	case "serve":
		relativeWorkdir := ".git-docs"
		if len(verbs) > 1 {
			relativeWorkdir = verbs[1]
		}

		workingDir, err := filepath.Abs(relativeWorkdir)
		if err != nil {
			fmt.Printf("cannot find working directory, abort (%v)\n", err)
			return
		}

		fmt.Printf("working directory: %s\n", workingDir)

		fmt.Println()

		webserver.Start(*port)
		break

	default:
		printHelp()
	}

	// parse command line for those actions :
	// * serve -port 8098 -insecure
	// 		=> future options : multi repositories
	// * document list -remoteUri=local
	// * document create  => opens a file with a template, creates the files, commit the changes, optionnally push
	// * document update DOCUMENT_ID   => same flow : file, modify the document files, commit the changes, optionnally push

	// option '-remoteUri=local' can be used to talk through the REST api of another git-docs server (http://...)
}
