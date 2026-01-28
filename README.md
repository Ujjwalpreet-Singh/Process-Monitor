# Process Monitor

A cross-platform process monitoring and analysis tool written in Go, supporting **Windows and Linux**.

The project collects OS-level process snapshots, reconstructs parent–child process trees, and supports offline snapshot diffing to detect process creation and termination events.  
It is designed with portability, correctness, and forensic reproducibility in mind, inspired by blue-team and incident response tooling.

---

## Features

- Cross-platform support (Windows + Linux)
- Native process snapshot collection
  - Windows: Toolhelp API
  - Linux: `/proc`
- Snapshot persistence in JSON format
- Parent–child process tree reconstruction
- Offline snapshot diffing (detect started / exited processes)
- Clean separation between data collection and analysis
- Single static binary per platform

---

## Why this project?

Process telemetry is a foundational building block of endpoint security, forensics, and incident response systems.

This project explores:
- How process data can be collected natively across operating systems
- How snapshots can be normalized into a common format
- How offline analysis (graphs and diffs) enables reproducible investigations
- How Go’s build system enables clean cross-platform system tooling

The goal is correctness and clarity rather than stealth or evasion.

---

## Architecture Overview

- **Collectors**  
  OS-specific components responsible for gathering process data.

- **Snapshots**  
  A snapshot represents a point-in-time view of all running processes.

- **Analysis Layer**  
  OS-agnostic logic that:
  - Builds process trees from snapshots
  - Diffs snapshots to detect execution changes

- **CLI Interface**  
  Provides snapshot, graph, and diff functionality.

OS-specific code is isolated using Go build tags, while analysis logic remains portable.

---

## Installation

### Build from source

#### Linux / WSL
```bash
go build -o process-monitor
```
#### Windows
```bash
go build -o process-monitor.exe
```

---

## Usage

### Take a process snapshot
```bash
./process-monitor --mode snapshot
```

### Save a snapshot
```bash
./process-monitor --mode snapshot --out out.json
```

### Build a process tree (Live)
```bash
./process-monitor --mode graph
```

### Build a process tree from an old snapshot
```bash
./process-monitor --mode graph --old old.json
```

### Diff two saved snapshots
```bash
./process --mode diff --old old.json --new.json
```
