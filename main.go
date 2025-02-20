package main

import (
	"bytes"
	"fmt"
	"os"
)

type Config struct {
	FilePath string
}

func main() {
	filePath := "data/emptydisk.txt"
	config := &Config{filePath}
	Run(config)
}

func Run(config *Config) {
	data, err := os.ReadFile(config.FilePath)
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
