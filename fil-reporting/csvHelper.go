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

type CsvHelper struct {
	mutex     sync.Mutex
	writer    *bufio.Writer
	hasHeader bool
}

var timeLogCsv *CsvHelper
var blockLogCsv *CsvHelper

func init() {

	timeLogCsv = &CsvHelper{hasHeader: false}
	blockLogCsv = &CsvHelper{hasHeader: false}

	pwd, _ := os.Getwd()

	timeLogCsvFile := "data-time-" + time.Now().Format("20060102-150405") + ".csv"
	blockLogCsvFile := "data-block-" + time.Now().Format("20060102-150405") + ".csv"

	if len(os.Args) == 2 {
		timeLogCsvFile = "data-time-" + strings.Split(os.Args[1], ".")[0] + "-" + time.Now().Format("2006-01-02 15:04:05") + ".csv"
		blockLogCsvFile = "data-block-" + strings.Split(os.Args[1], ".")[0] + "-" + time.Now().Format("2006-01-02 15:04:05") + ".csv"
	}

	timeLogCsvFile = filepath.Join(pwd, timeLogCsvFile)
	blockLogCsvFile = filepath.Join(pwd, blockLogCsvFile)

	timeLogCsvF, err := os.OpenFile(timeLogCsvFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	blockLogCsvF, err := os.OpenFile(blockLogCsvFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	os.Chmod(timeLogCsvFile, 0666)
	os.Chmod(blockLogCsvFile, 0666)

	timeLogCsv.writer = bufio.NewWriter(timeLogCsvF)
	blockLogCsv.writer = bufio.NewWriter(blockLogCsvF)

}

func (csv *CsvHelper) writeCsvLine(line string, header string) {

	if !strings.HasSuffix(header, "\n") {
		line += "\n"
	}

	if !strings.HasSuffix(line, "\n") {
		line += "\n"
	}

	csv.mutex.Lock()

	if !csv.hasHeader {
		_, err := csv.writer.WriteString(header)
		if err != nil {
			fmt.Printf("Write CSV Header: %v\n", err)
		} else {
			csv.hasHeader = true
		}
	}

	_, err := csv.writer.WriteString(line)
	if err != nil {
		fmt.Printf("Write CSV: %v\n", err)
	}
	err = csv.writer.Flush()
	if err != nil {
		fmt.Printf("Write CSV Flush: %v\n", err)
	}

	csv.mutex.Unlock()
}
