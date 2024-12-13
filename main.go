package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

func worker(host string, portChan <-chan int, openPortsChan chan<- int, closedPortsChan chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()

	for port := range portChan {
		address := net.JoinHostPort(host, strconv.Itoa(port))
		conn, err := net.DialTimeout("tcp", address, time.Second)
		if err != nil {
			closedPortsChan <- port
		} else {
			conn.Close()
			openPortsChan <- port
		}
	}
}

func feedPorts(ports []int, portChan chan int) {
	for _, port := range ports {
		portChan <- port
	}
	close(portChan)
}

func addResults(portChan chan int, results *[]int, wg *sync.WaitGroup) {
	defer wg.Done()
	for port := range portChan {
		*results = append(*results, port)
	}
}

func scanner(host string, ports []int, workers int) ([]int, []int) {
	portChan := make(chan int)
	openPortsChan := make(chan int)
	closedPortsChan := make(chan int)

	var wg sync.WaitGroup
	var resultsWg sync.WaitGroup

	// Start scanning
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(host, portChan, openPortsChan, closedPortsChan, &wg)
	}

	go feedPorts(ports, portChan)

	// Start processing results while the workers are working
	var openPorts, closedPorts []int
	resultsWg.Add(2)
	go addResults(openPortsChan, &openPorts, &resultsWg)
	go addResults(closedPortsChan, &closedPorts, &resultsWg)

	// Wait for the scan to finishh to close the channels
	wg.Wait()
	close(openPortsChan)
	close(closedPortsChan)

	// Wait for the results
	resultsWg.Wait()

	return closedPorts, openPorts
}

func getPorts(n int) []int {
	var result []int
	for i := 1; i <= n; i++ {
		result = append(result, i)
	}
	return result
}

func main() {
	fmt.Println("Starting scan")
	closedPorts, openPorts := scanner("loremips.um", getPorts(65535), 500)
	fmt.Println("Closed Ports:", closedPorts)
	fmt.Println("Open Ports:", openPorts)
}
