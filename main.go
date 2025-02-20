package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
)

const (
	Name    = "shell"
	Version = "0.1.0"
)

type Config struct {
	InputPath string
}

func main() {
	var isHelp bool
	var isVersion bool
	var inputPath string
	flag.BoolVar(&isHelp, "h", false, "print help message")
	flag.BoolVar(&isHelp, "H", false, "print help message")
	flag.BoolVar(&isHelp, "?", false, "print help message")
	flag.BoolVar(&isHelp, "help", false, "print help message")
	flag.BoolVar(&isVersion, "v", false, "print version")
	flag.BoolVar(&isVersion, "V", false, "print version")
	flag.BoolVar(&isVersion, "version", false, "print version")
	flag.StringVar(&inputPath, "i", "", "input file path")
	flag.Parse()

	if isHelp {
		flag.CommandLine.Usage()
		return
	}

	if isVersion {
		fmt.Printf("%s %s\n", Name, Version)
		return
	}

	config := &Config{inputPath}
	Run(config)
}

func Run(config *Config) {
	var err error
	var data []byte
	if config.InputPath == "" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(config.InputPath)
	}

	if err != nil {
		panic(err)
	}

	raw := trim(data)
	prettyPrint(raw)
}

func trim(data []byte) [][]byte {
	raw := bytes.Split(data, []byte("\n"))
	raw = raw[2 : len(raw)-1]

	for i := range raw {
		raw[i] = append([]byte("0"), raw[i][3:len(raw[i])-1]...)
	}

	return raw
}

func prettyPrint(raw [][]byte) {
	const header string = `XX:                1               2               3
XX:0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF`

	fmt.Println(header)
	for i, line := range raw {
		fmt.Printf("%02X:%s\n", i, string(append(line[1:], []byte("0")...)))
	}
}
