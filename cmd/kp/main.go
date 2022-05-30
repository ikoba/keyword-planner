package main

import (
	"flag"
	"log"
	"os"

	kp "github.com/ikoba/keyword-planner"
)

// /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --remote-debugging-port=9222

func main() {
	var outDir string
	var port int
	var timeout int
	var retry int
	var debug bool

	flag.StringVar(&outDir, "out", "", "output directory (required)")
	flag.IntVar(&port, "port", 9222, "port number set in the chrome startup option `--remote-debugging-port`")
	flag.IntVar(&timeout, "timeout", 20, "number of seconds until timeout for each word")
	flag.IntVar(&retry, "retry", 3, "number of retries when timeout occurs for each word")
	flag.BoolVar(&debug, "debug", false, "show debug log")
	flag.Parse()

	words := flag.Args()
	if len(words) < 1 {
		log.Fatalln("argument 'WORD' is required")
	}
	if outDir == "" {
		log.Fatal("option '-out' is required")
	} else if f, err := os.Stat(outDir); os.IsNotExist(err) || !f.IsDir() {
		log.Fatalln("directory set by '-out' dose not exist")
	}
	r := &kp.Request{Words: words, OutDir: outDir, Port: port, Timeout: timeout, Retry: retry, Debug: debug}

	err := r.Execute()
	if err != nil {
		log.Fatalln(err)
	}
}
