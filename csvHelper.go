package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var csvMutex sync.Mutex
var csvWriter *bufio.Writer

func init() {

	csvMutex.Lock()
	file, err := os.OpenFile("data-"+time.Now().Format("2006-01-02 15:04:05")+".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvWriter = bufio.NewWriter(file)

	csvMutex.Unlock()
}

func writeCsvLine(line string) {

	if !strings.HasSuffix(line, "\n") {
		line += "\n"
	}

	csvMutex.Lock()

	_, err := csvWriter.WriteString(line)
	if err != nil {
		fmt.Printf("Write CSV: %v\n", err)
	}
	err = csvWriter.Flush()
	if err != nil {
		fmt.Printf("Write CSV Flush: %v\n", err)
	}

	csvMutex.Unlock()
}
