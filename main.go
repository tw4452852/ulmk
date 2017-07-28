package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

var (
	VERSION = "unknown"
)

func main() {
	help := flag.Bool("h", false, "show help")
	threshold := flag.String("t", "64M", "set threshold")
	version := flag.Bool("v", false, "show version")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *version {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	thres, err := parseThreshold(*threshold)
	if err != nil {
		log.Fatalf("parseThreshold failed: %s\n", err)
	}

	eventFile, err := setupEventfd(thres)
	if err != nil {
		log.Fatalf("setupEventfd failed: %s\n", err)
	}
	defer eventFile.Close()

	for {
		buf := make([]byte, 8)
		_, err := eventFile.Read(buf)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("TRESHOLD CROSSED: %v", buf)
	}

}

const (
	K = 1024
	M = 1024 * 1024
	G = 1024 * 1024 * 1024
)

var validUnits = [...]struct {
	symbol     string
	multiplier int64
}{
	{"k", K},
	{"m", M},
	{"g", G},
	{"", 1},
}

func parseThreshold(threshold string) (int64, error) {
	threshold = strings.ToLower(threshold)

	for _, unit := range validUnits {
		if strings.HasSuffix(threshold, unit.symbol) {
			size, err := strconv.ParseInt(threshold[:len(threshold)-len(unit.symbol)], 10, 64)
			if err != nil {
				return -1, err
			}
			return size * unit.multiplier, nil
		}
	}
	return -1, fmt.Errorf("can't parse %q", threshold)
}

func setupEventfd(threshold int64) (f *os.File, err error) {
	host, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	const memcgPrefix = "/sys/fs/cgroup/memory/lxc"
	memoryUsagePath := filepath.Join(memcgPrefix, host, "memory.usage_in_bytes")
	eventControlPath := filepath.Join(memcgPrefix, host, "cgroup.event_control")
	totalMemoryPath := filepath.Join(memcgPrefix, host, "memory.limit_in_bytes")

	content, err := ioutil.ReadFile(totalMemoryPath)
	if err != nil {
		return nil, err
	}
	totalMemory, err := strconv.ParseInt(string(content[:len(content)-1]), 10, 64)
	if err != nil {
		return nil, err
	}
	log.Printf("total memory: %d\n", totalMemory)
	threshold = totalMemory - threshold
	log.Printf("set threshold: %d\n", threshold)

	sysEventfd, _, syserr := syscall.RawSyscall(
		syscall.SYS_EVENTFD2, 0, syscall.FD_CLOEXEC, 0,
	)
	if syserr != 0 {
		return nil, syserr
	}

	eventFile := os.NewFile(sysEventfd, "eventfd")
	if eventFile == nil {
		return nil, fmt.Errorf("invalid eventfd[%d]", sysEventfd)
	}
	defer func() {
		if err != nil {
			eventFile.Close()
		}
	}()

	memoryUsageFile, err := os.Open(memoryUsagePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			memoryUsageFile.Close()
		}
	}()

	eventControlData := fmt.Sprintf(
		"%d %d %d",
		eventFile.Fd(), memoryUsageFile.Fd(), threshold,
	)

	err = ioutil.WriteFile(
		eventControlPath,
		[]byte(eventControlData),
		0222,
	)
	if err != nil {
		return nil, err
	}

	log.Printf(eventControlData)
	return eventFile, nil
}
