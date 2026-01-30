//go:build linux
// +build linux

package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type LinuxProcessCollector struct{}

func (l *LinuxProcessCollector) Run(ctx context.Context, out chan<- Event) {
	<-ctx.Done()
}

func (l *LinuxProcessCollector) Snapshot(ctx context.Context) (ProcessSnapshot, error) {
	snapshot := ProcessSnapshot{
		Timestamp: time.Now(),
		Processes: []ProcessInfo{},
	}

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return snapshot, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		if err != nil {
			continue
		}

		parts := strings.Fields(string(data))
		if len(parts) < 22 {
			continue
		}

		ppid, err := strconv.Atoi(parts[3])
		if err != nil {
			continue
		}

		startTicks,err := strconv.ParseUint(parts[21],10,64)
		if err != nil{
			continue
		}

		bootTime, _ := GetBootTime()
		hz:= GetClockTicks()

		name := strings.Trim(parts[1], "()")

		converttime := ConvertStartTicksToTime(startTicks, bootTime, hz)

		snapshot.Processes = append(snapshot.Processes, ProcessInfo{
			PID:  pid,
			PPID: ppid,
			Name: name,
			StartTime: startTicks,
			OsTime: converttime,
		})
	}

	return snapshot, nil
}


func GetBootTime() (time.Time, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return time.Time{}, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "btime ") {
			parts := strings.Fields(line)
			if len(parts) != 2 {
				continue
			}

			sec, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return time.Time{}, err
			}

			return time.Unix(sec, 0), nil
		}
	}

	return time.Time{}, fmt.Errorf("btime not found")
}

func GetClockTicks() int64 {
	return 100
}

func ConvertStartTicksToTime(startTicks uint64, bootTime time.Time, hz int64) time.Time {
	seconds := float64(startTicks) / float64(hz)
	return bootTime.Add(time.Duration(seconds * float64(time.Second)))
}
