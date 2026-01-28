//go:build linux
// +build linux

package main

func NewProcessCollector() Collector {
	return &LinuxProcessCollector{}
}
