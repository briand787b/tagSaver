package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"sync"
	// "github.com/fatih/color"
	"path/filepath"
)

const (
	port = ":3000"
)

var (
	ip          = flag.String("ip", "localhost", "the ip to which the formatted JSON should be sent")
	filePattern = flag.String("pattern", "*.json", "the pattern to match against for selecting files to send")
)

func main() {
	flag.Parse()
	// fmt.Println("ip: ", *ip)
	// fmt.Println("filePattern: ", *filePattern)

	fds, err := filepath.Glob(*filePattern)
	if err != nil {
		log.Fatalf("error matching pattern to files: %s", err)
	}

	conn, err := net.Dial("tcp", *ip+port)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	wg := sync.WaitGroup{}
	for _, fd := range fds {
		wg.Add(1)
		go writeJSONToConn(conn, &wg, fd)
	}

	wg.Wait()
	fmt.Println("program exiting...")
}

func writeJSONToConn(c net.Conn, wg *sync.WaitGroup, fileName string) {
	defer wg.Done()

	fd, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("error opening file (%s): %s\n", fileName, err)
		return
	}

	defer fd.Close()

	scnr := bufio.NewScanner(fd)
	scnr.Split(scanDoubleQuotations)
	for scnr.Scan() {
		_, err := c.Write([]byte(scnr.Text() + "\n"))
		if err != nil {
			fmt.Println("error writing to connection: ", err)
		}
	}
}

func scanDoubleQuotations(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	rgx := regexp.MustCompile(`"[\s,\S]+?"`)
	if byteRange := rgx.FindIndex(data); byteRange != nil {
		// fmt.Println("DEBUG: byteRange: ", byteRange)

		tag := dropCR(data[byteRange[0]+1 : byteRange[1]-1])
		// fmt.Println("DEBUG: tag: ", string(tag))

		rgxWhiteSpace := regexp.MustCompile(`\s+`)
		cleansedData := rgxWhiteSpace.ReplaceAll(tag, []byte{})
		// fmt.Println("DEBUG: cleansed data: ", string(cleansedData))
		return byteRange[1] + 1, cleansedData, nil
	}

	return 0, nil, nil
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}

	return data
}
