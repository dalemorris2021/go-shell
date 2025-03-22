package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	Name       = "shell"
	Version    = "0.1.0"
	numBytes   = 32
	numColumns = 64
	header     = `XX:                1               2               3
XX:0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF`
	numRows = 32
)

type ShellAction int

const (
	Disk = iota
	Type
	Dir
)

type Config struct {
	InputPath string
	Action    ShellAction
	TypePath  string
}

type cluster interface {
}

type emptyCluster struct {
	nextEmpty int
}

type damagedCluster struct {
	nextDamaged int
}

type fileDataCluster struct {
	content  string
	nextData int
}

type fileHeaderCluster struct {
	name       string
	content    string
	nextHeader int
	nextData   int
}

type rootCluster struct {
	name    string
	empty   int
	damaged int
	headers int
}

func main() {
	var isHelp bool
	var isVersion bool
	var isDir bool
	var isType bool
	var typePath string
	var inputPath string

	flag.BoolVar(&isHelp, "h", false, "print help message")
	flag.BoolVar(&isHelp, "H", false, "print help message")
	flag.BoolVar(&isHelp, "?", false, "print help message")
	flag.BoolVar(&isHelp, "help", false, "print help message")
	flag.BoolVar(&isVersion, "v", false, "print version")
	flag.BoolVar(&isVersion, "V", false, "print version")
	flag.BoolVar(&isVersion, "version", false, "print version")
	flag.BoolVar(&isDir, "dir", false, "print files in root directory")
	flag.StringVar(&typePath, "type", "", "prints contents of file")
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

	if typePath != "" {
		isType = true
	}

	var action ShellAction
	if isDir {
		action = Dir
	} else if isType {
		action = Type
	} else {
		action = Disk
	}

	config := &Config{inputPath, action, typePath}
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

	stringData := string(data)
	raw := trim(stringData)
	clusters := rawToClusters(raw)

	switch config.Action {
	case Disk:
		printDisk(raw)
	case Type:
		printContent(clusters, config.TypePath)
	case Dir:
		printFiles(clusters)
	}
}

func trim(stringData string) [][]byte {
	var err error
	const hexConversion int64 = 16

	lines := strings.Split(stringData, "\n")
	lines = lines[2 : len(lines)-1]

	for i := range lines {
		lines[i] = "0" + lines[i][3:len(lines[i])-1]
	}

	raw := make([][]byte, numRows)
	for i := range raw {
		raw[i] = make([]byte, numBytes)
		var bigEnd int64
		var littleEnd int64
		for j := range raw[i] {
			bigEnd, err = strconv.ParseInt(string(lines[i][2*j]), 16, 64)
			if err != nil {
				panic(err)
			}

			littleEnd, err = strconv.ParseInt(string(lines[i][2*j+1]), 16, 64)
			if err != nil {
				panic(err)
			}

			raw[i][j] = byte(hexConversion*bigEnd + littleEnd)
		}
	}

	return raw
}

func printDisk(raw [][]byte) {
	var stringData []string = make([]string, numRows)
	var temp string
	for i := range raw {
		stringData[i] = ""
		for j, b := range raw[i] {
			if j == 0 {
				temp = fmt.Sprintf("%X", b)
			} else {
				temp = fmt.Sprintf("%02X", b)
			}

			stringData[i] += temp
		}
		stringData[i] += "0"
	}

	fmt.Println(header)
	for i := range raw {
		fmt.Printf("%02X:%s\n", i, stringData[i])
	}
}

func printContent(clusters []cluster, fileName string) {
	fileFound := false
	for _, cluster := range clusters {
		fileHeader, ok := cluster.(*fileHeaderCluster)
		if ok && fileHeader.name == fileName {
			fmt.Print(fileHeader.content)
			if fileHeader.nextData != 0 {
				fileData, ok := clusters[fileHeader.nextData].(*fileDataCluster)
				if ok {
					fmt.Print(fileData.content)
					for fileData.nextData != 0 {
						fileData, ok := clusters[fileData.nextData].(*fileDataCluster)
						if ok {
							fmt.Print(fileData.content)
						}
					}
				}
			}
			fmt.Println()
			fileFound = true
			break
		}
	}

	if !fileFound {
		fmt.Fprintf(os.Stderr, "[Error] File not found: %s\n", fileName)
	}
}

func printFiles(clusters []cluster) {
	for _, cluster := range clusters {
		fileHeader, ok := cluster.(*fileHeaderCluster)
		if ok {
			fmt.Println(fileHeader.name)
		}
	}
}

func rawToClusters(raw [][]byte) []cluster {
	clusters := make([]cluster, numRows)
	for i, line := range raw {
		clusters[i] = rawToCluster(line)
	}

	return clusters
}

func rawToCluster(raw []byte) cluster {
	var c cluster
	switch clusterType := raw[0]; clusterType {
	case 0:
		var buffer bytes.Buffer
		for _, val := range createRange(4, numBytes) {
			b := raw[val]
			if b == 0 {
				break
			} else {
				buffer.WriteByte(b)
			}
		}

		name := buffer.String()
		empty := int(raw[1])
		damaged := int(raw[2])
		headers := int(raw[3])
		c = &rootCluster{name, empty, damaged, headers}
	case 1:
		nextEmpty := int(raw[1])
		c = &emptyCluster{nextEmpty}
	case 2:
		nextDamaged := int(raw[1])
		c = &damagedCluster{nextDamaged}
	case 3:
		var buffer bytes.Buffer
		contentStart := numBytes

		for _, val := range createRange(3, numBytes) {
			b := raw[val]
			if b == 0 {
				contentStart = val + 1
				break
			} else {
				buffer.WriteByte(b)
			}
		}

		name := buffer.String()
		buffer.Reset()

		for _, val := range createRange(contentStart, numBytes) {
			b := raw[val]
			if b == 0 {
				break
			} else {
				buffer.WriteByte(b)
			}
		}

		content := buffer.String()
		nextHeader := int(raw[1])
		nextData := int(raw[2])
		c = &fileHeaderCluster{name, content, nextHeader, nextData}
	case 4:
		var buffer bytes.Buffer
		for _, val := range createRange(2, numBytes) {
			b := raw[val]
			if b == 0 {
				break
			} else {
				buffer.WriteByte(b)
			}
		}

		content := buffer.String()
		nextData := int(raw[1])
		c = &fileDataCluster{content, nextData}
	}

	return c
}
