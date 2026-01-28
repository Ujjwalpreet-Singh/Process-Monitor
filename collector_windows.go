//go:build windows
// +build windows

package main

import (
	"context"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

type WindowsProcessCollector struct{}

func (w *WindowsProcessCollector) Run(ctx context.Context, out chan<- Event) {
	<-ctx.Done()
}

func (w *WindowsProcessCollector) Snapshot(ctx context.Context) (ProcessSnapshot, error) {
	snapshot := ProcessSnapshot{
		Timestamp: time.Now(),
		Processes: []ProcessInfo{},
	}

	handle, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return snapshot, err
	}
	defer windows.CloseHandle(handle)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	if err := windows.Process32First(handle, &entry); err != nil {
		return snapshot, err
	}

	for {
		name := windows.UTF16ToString(entry.ExeFile[:])
		snapshot.Processes = append(snapshot.Processes, ProcessInfo{
			PID:  int(entry.ProcessID),
			PPID: int(entry.ParentProcessID),
			Name: name,
		})

		if err := windows.Process32Next(handle, &entry); err != nil {
			break
		}
	}

	return snapshot, nil
}
