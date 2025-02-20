package main

import (
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
	fmt.Print(string(data))
}
