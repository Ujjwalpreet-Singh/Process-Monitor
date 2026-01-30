package main

import (
	"context"
	"fmt"
	"time"
	"encoding/json"
	"os"
	"path/filepath"
)

/* =======================
   Event & Core Types
   ======================= */

type Source string

const (
	SourceProcess Source = "process"
)

type ProcessInfo struct {
	PID  int
	PPID int
	Name string
	StartTime uint64
	OsTime time.Time
}

type ProcessKey struct {
	PID            int
	StartTime time.Time
}

type Event struct {
	Timestamp time.Time
	Source    Source
	Message   string
	Process   *ProcessInfo
}

type ProcessSnapshot struct {
    Timestamp time.Time
    Processes []ProcessInfo
}

type ProcessNode struct {
	Info     ProcessInfo
	Children []*ProcessNode
}

type SnapshotDiff struct {
	Started []ProcessInfo
	Exited  []ProcessInfo
}

/* =======================
   Collector Interface
   ======================= */

type Collector interface {
	Run(ctx context.Context, out chan<- Event)
}

type SnapshotCollector interface {
    Snapshot(ctx context.Context) (ProcessSnapshot, error)
}

/* =======================
   Agent
   ======================= */

type Agent struct {
	collectors []Collector
	events     chan Event
}

func NewAgent(collectors []Collector, buffer int) *Agent {
	return &Agent{
		collectors: collectors,
		events:     make(chan Event, buffer),
	}
}

func (a *Agent) Run(ctx context.Context) {
	// Start collectors
	for _, c := range a.collectors {
		go c.Run(ctx, a.events)
	}

	// Event processing loop
	for {
		select {
		case <-ctx.Done():
			return

		case e := <-a.events:
			a.handleEvent(e)
		}
	}
}

func (a *Agent) handleEvent(e Event) {
	if e.Process != nil {
		fmt.Printf(
			"[%s] %s PID=%d PPID=%d NAME=%s\n",
			e.Timestamp.Format(time.RFC3339),
			e.Message,
			e.Process.PID,
			e.Process.PPID,
			e.Process.Name,
		)
		return
	}

	fmt.Printf("[%s] %s\n", e.Timestamp.Format(time.RFC3339), e.Message)
}

func (a *Agent) TakeSnapshot(ctx context.Context) (*ProcessSnapshot, error) {
    for _, c := range a.collectors {
        if sc, ok := c.(SnapshotCollector); ok {
            snap, err := sc.Snapshot(ctx)
            if err != nil {
                return nil, err
            }
            return &snap, nil
        }
    }
    return nil, fmt.Errorf("no snapshot-capable collector found")
}

/* =======================
	APIs
	====================== */

func BuildProcessGraph(snapshot ProcessSnapshot) []*ProcessNode {
	// PID -> node
	nodes := make(map[int]*ProcessNode)

	// Pass 1: create nodes
	for _, p := range snapshot.Processes {
		nodes[p.PID] = &ProcessNode{
			Info:     p,
			Children: nil,
		}
	}

	// Pass 2: link children to parents
	for _, node := range nodes {
		ppid := node.Info.PPID
		parent, exists := nodes[ppid]
		if exists {
			parent.Children = append(parent.Children, node)
		}
	}

	// Pass 3: collect roots
	var roots []*ProcessNode
	for _, node := range nodes {
		ppid := node.Info.PPID
		if ppid == 0 {
			roots = append(roots, node)
			continue
		}
		if _, exists := nodes[ppid]; !exists {
			roots = append(roots, node)
		}
	}

	return roots
}

func PrintProcessTree(nodes []*ProcessNode, prefix string,depth int) {
	if depth == 0 {
		return
	}

	for i, node := range nodes {
		isLast := i == len(nodes)-1

		branch := "├── "
		pipe := "│   "


		if isLast {
			branch = "└── "
			pipe = "    "
		}

		fmt.Printf("%s%sPID=%d NAME=%s\n",
			prefix,
			branch,
			node.Info.PID,
			node.Info.Name,
		)

		if len(node.Children) > 0 {
			PrintProcessTree(node.Children, prefix+pipe,depth-1)
		}
	}
}

func SaveProcessGraphToFile(roots []*ProcessNode, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(roots)
}

func DiffSnapshots(oldSnap, newSnap ProcessSnapshot) SnapshotDiff {
	oldMap := make(map[ProcessKey]ProcessInfo)
	newMap := make(map[ProcessKey]ProcessInfo)

	for _, p := range oldSnap.Processes {
		oldMap[ProcessKeyFromInfo(p)] = p
	}
	for _, p := range newSnap.Processes {
		newMap[ProcessKeyFromInfo(p)] = p
	}

	diff := SnapshotDiff{}

	for k, p := range newMap {
		if _, ok := oldMap[k]; !ok {
			diff.Started = append(diff.Started, p)
		}
	}

	for k, p := range oldMap {
		if _, ok := newMap[k]; !ok {
			diff.Exited = append(diff.Exited, p)
		}
	}

	return diff
}


func SaveSnapshotToFile(snapshot ProcessSnapshot, path string) error {
	cleanpath := filepath.Clean(path)

	dir := filepath.Dir(cleanpath)
	if dir != "." && dir != ""{
		if _,err := os.Stat(dir); os.IsNotExist(err){
			err := os.MkdirAll(dir,0755)
			if err != nil{
				return err
			}
		}
	}
	fullpath := cleanpath
    file, err := os.Create(fullpath)
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")

    return encoder.Encode(snapshot)
}

func LoadSnapshotFromFile(path string) (*ProcessSnapshot, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var snapshot ProcessSnapshot
    decoder := json.NewDecoder(file)

    if err := decoder.Decode(&snapshot); err != nil {
        return nil, err
    }

    return &snapshot, nil
}

func ProcessKeyFromInfo(p ProcessInfo) ProcessKey {
	return ProcessKey{
		PID:       p.PID,
		StartTime: p.OsTime,
	}
}
