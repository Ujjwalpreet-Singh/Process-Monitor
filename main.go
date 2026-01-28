package main

import (
	"context"
	"flag"
	"fmt"
	"time"
)

func main() {
	ctx := context.Background()
	
	events := flag.Int("events",100,"Maximum amount of events to show")
	mode := flag.String("mode", "snapshot", "snapshot | graph | diff")
	interval := flag.Int("interval", 3, "seconds between snapshots (diff mode)")
	out := flag.String("out","","output file for snapshot (JSON)")
	oldpath := flag.String("old","","old snapshot file (JSON)")
	newpath := flag.String("new","","new snapshot file (JSON)")
	flag.Parse()

	pc := NewProcessCollector()
	agent := NewAgent([]Collector{pc}, *events)

	switch *mode {

	case "snapshot":
		snapshot, err := agent.TakeSnapshot(ctx)
		if err != nil {
			panic(err)
		}
		
		if *out != "" {
			if err := SaveSnapshotToFile(*snapshot,*out); err != nil{
				panic(err)
			}
			fmt.Println("Snapshot saved to",*out)
			return
		}
		
		fmt.Printf("Snapshot at %s\n", snapshot.Timestamp)
		for _, p := range snapshot.Processes {
			fmt.Printf("PID=%d PPID=%d NAME=%s\n", p.PID, p.PPID, p.Name)
		}

	case "graph":
		
		var snapshot *ProcessSnapshot
		var err error
		if *oldpath != "" {
			snapshot,err = LoadSnapshotFromFile(*oldpath)
			if err != nil {
				panic(err)
			}
		} else {
			snapshot, err = agent.TakeSnapshot(ctx)
			if err != nil {
				panic(err)
			}
		}

		roots := BuildProcessGraph(*snapshot)
		fmt.Println("Process Tree:")
		PrintProcessTree(roots, "",10)

	case "diff":
	
		if *oldpath != "" && *newpath != "" {
			oldSnap,err := LoadSnapshotFromFile(*oldpath)
			if err != nil{
				panic(err)
			}
			
			newSnap,err := LoadSnapshotFromFile(*newpath)
			if err != nil{
				panic(err)
			}
			
			diff := DiffSnapshots(*oldSnap, *newSnap)

			fmt.Println("\nProcesses started:")
			for _, p := range diff.Started {
				fmt.Printf("PID=%d PPID=%d NAME=%s\n", p.PID, p.PPID, p.Name)
			}

			fmt.Println("\nProcesses exited:")
			for _, p := range diff.Exited {
				fmt.Printf("PID=%d PPID=%d NAME=%s\n", p.PID, p.PPID, p.Name)
				
			return
			}
		fmt.Println("Taking first snapshot...")
		snap1, err := agent.TakeSnapshot(ctx)
		if err != nil {
			panic(err)
		}

		time.Sleep(time.Duration(*interval) * time.Second)

		fmt.Println("Taking second snapshot...")
		snap2, err := agent.TakeSnapshot(ctx)
		if err != nil {
			panic(err)
		}

		diff = DiffSnapshots(*snap1, *snap2)

		fmt.Println("\nProcesses started:")
		for _, p := range diff.Started {
			fmt.Printf("PID=%d PPID=%d NAME=%s\n", p.PID, p.PPID, p.Name)
		}

		fmt.Println("\nProcesses exited:")
		for _, p := range diff.Exited {
			fmt.Printf("PID=%d PPID=%d NAME=%s\n", p.PID, p.PPID, p.Name)
		}
		}
	default:
		fmt.Println("Unknown mode:", *mode)
		fmt.Println("Use --mode snapshot | graph | diff")
	}
}

