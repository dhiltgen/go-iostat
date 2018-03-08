package main

// Simple utility to display disk utilization

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	Interval = 5 * time.Second
)

type Requests struct {
	IOs     uint64 // Units: requests
	Merges  uint64 // Units: requests
	Sectors uint64 // Units: sectors
	Ticks   uint64 // Units: milliseconds
}

type BlockStat struct {
	Name        string
	Read        Requests
	Write       Requests
	InFlight    uint64 // Units: requests
	TotalTicks  uint64 // Units: milliseconds
	TimeInQueue uint64 // Units: milliseconds
}

func GetData(deviceFiles []string) ([]BlockStat, error) {
	res := []BlockStat{}
	for _, deviceFile := range deviceFiles {
		r, err := os.Open(deviceFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open %s: %s", deviceFile, err)
			return nil, err
		}
		nameSplit := strings.Split(deviceFile, "/")
		if len(nameSplit) != 5 {
			return nil, fmt.Errorf("Malforled device name: %s %v", deviceFile, nameSplit)
		}
		d := BlockStat{Name: nameSplit[3]}
		fmt.Fscanf(r, "%d %d %d %d %d %d %d %d %d %d %d",
			&d.Read.IOs,
			&d.Read.Merges,
			&d.Read.Sectors,
			&d.Read.Ticks,
			&d.Write.IOs,
			&d.Write.Merges,
			&d.Write.Sectors,
			&d.Write.Ticks,
			&d.InFlight,
			&d.TotalTicks,
			&d.TimeInQueue,
		)
		//fmt.Println(d.ToString()) // Debugging
		res = append(res, d)
	}
	return res, nil
}

func (d BlockStat) ToString() string {
	return fmt.Sprintf("%s %d %d %d %d %d %d %d %d %d %d %d",
		d.Name,
		d.Read.IOs,
		d.Read.Merges,
		d.Read.Sectors,
		d.Read.Ticks,
		d.Write.IOs,
		d.Write.Merges,
		d.Write.Sectors,
		d.Write.Ticks,
		d.InFlight,
		d.TotalTicks,
		d.TimeInQueue,
	)
}

func (d BlockStat) GetUtil(prev BlockStat, delta time.Duration) float64 {
	return (float64(d.TotalTicks) - float64(prev.TotalTicks)) / float64(delta/time.Millisecond) * 100
}

func main() {
	deviceFiles, err := filepath.Glob("/sys/block/sd*/stat")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load devices: %s", err)
		os.Exit(1)
	}
	if len(deviceFiles) == 0 {
		fmt.Println("No devices found")
		os.Exit(0)
	}
	lastData, err := GetData(deviceFiles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load device stats: %s", err)
		os.Exit(1)
	}
	lastTime := time.Now()
	ticker := time.NewTicker(Interval)
	fmt.Println(`DEV,%UTIL,TIME`)
	for {
		select {
		case <-ticker.C:
			newData, err := GetData(deviceFiles)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to load device stats: %s", err)
				continue
			}

			now := time.Now()
			for i, dev := range newData {
				fmt.Printf("%s,%0.2f,%s\n", dev.Name, dev.GetUtil(lastData[i], now.Sub(lastTime)), now.Format(time.RFC3339))
			}
			lastData = newData
			lastTime = now
		}
	}
}
