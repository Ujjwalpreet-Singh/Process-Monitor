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
		if len(parts) < 4 {
			continue
		}

		ppid, err := strconv.Atoi(parts[3])
		if err != nil {
			continue
		}

		name := strings.Trim(parts[1], "()")

		snapshot.Processes = append(snapshot.Processes, ProcessInfo{
			PID:  pid,
			PPID: ppid,
			Name: name,
		})
	}

	return snapshot, nil
}
