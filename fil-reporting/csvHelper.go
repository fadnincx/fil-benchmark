package fil_reporting

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var csvMutex sync.Mutex
var csvWriter *bufio.Writer
var hasHeader = false

func init() {

	csvMutex.Lock()
	pwd, _ := os.Getwd()
	dataFile := "data-" + time.Now().Format("20060102-150405") + ".csv"
	if len(os.Args) == 2 {
		dataFile = "data-" + strings.Split(os.Args[1], ".")[0] + "-" + time.Now().Format("2006-01-02 15:04:05") + ".csv"
	} else if len(os.Args) == 3 {
		dataFile = os.Args[2]
	}
	if !filepath.IsAbs(dataFile) {
		dataFile = filepath.Join(pwd, dataFile)
	}
	file, err := os.OpenFile(dataFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	os.Chmod(dataFile, 0666)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	csvWriter = bufio.NewWriter(file)

	hasHeader = false
	csvMutex.Unlock()

}

func writeCsvLine(line string, header string) {

	if !strings.HasSuffix(header, "\n") {
		line += "\n"
	}

	if !strings.HasSuffix(line, "\n") {
		line += "\n"
	}

	csvMutex.Lock()

	if !hasHeader {
		_, err := csvWriter.WriteString(header)
		if err != nil {
			fmt.Printf("Write CSV Header: %v\n", err)
		} else {
			hasHeader = true
		}
	}

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
