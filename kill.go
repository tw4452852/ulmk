package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type task struct {
	pid    int
	name   string
	rss    int
	oomAdj int
}

func (t *task) String() string {
	return fmt.Sprintf("[pid: %d, name: %q, rss: %d, oomAdj: %d]",
		t.pid, t.name, t.rss, t.oomAdj)
}

type victims []*task

func (v victims) Len() int      { return len(v) }
func (v victims) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v victims) Less(i, j int) bool {
	// compare oomAdj at first
	if v[i].oomAdj > v[j].oomAdj {
		return true
	}
	if v[i].oomAdj < v[j].oomAdj {
		return false
	}

	// compare rss secondly
	if v[i].rss > v[j].rss {
		return true
	}
	if v[i].rss < v[j].rss {
		return false
	}

	// compare pid in the end
	if v[i].pid > v[j].pid {
		return true
	}

	return false
}

func killOne() {
	// exclude init, myself and my parent
	excludes := map[int]struct{}{
		1:            {},
		os.Getpid():  {},
		os.Getppid(): {},
	}

	vs := findVictims(excludes)

	if len(vs) == 0 {
		log.Println("no victim found")
		return
	}

	victim := vs[0]
	log.Printf("try to kill %s\n", victim)
	p, err := os.FindProcess(victim.pid)
	if err != nil {
		log.Printf("can't find victim: %s\n", err)
		return
	}

	err = p.Kill()
	if err != nil {
		log.Printf("kill failed: %s\n", err)
		return
	}
}

func findVictims(excludes map[int]struct{}) victims {
	processPath := filepath.Join(memcgPrefix, host, "cgroup.procs")
	f, err := os.Open(processPath)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	var vs victims
	for s.Scan() {
		ps := s.Text()
		pid, err := strconv.Atoi(ps)
		if err != nil {
			log.Printf("can't convert %q to number: %s\n", ps, err)
			continue
		}

		// skip excludes
		if _, ok := excludes[pid]; ok {
			continue
		}

		task, err := getTask(pid)
		if err != nil {
			log.Printf("can't get %d task: %s\n", pid, err)
			continue
		}
		vs = append(vs, task)
	}

	if err := s.Err(); err != nil {
		log.Println(err)
	}

	sort.Sort(vs)
	return vs
}

func getTask(pid int) (*task, error) {
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	content, err := ioutil.ReadFile(statPath)
	if err != nil {
		return nil, err
	}

	fields := strings.Fields(string(content))
	rss, err := strconv.Atoi(fields[23])
	if err != nil {
		return nil, err
	}

	oomAdjPath := fmt.Sprintf("/proc/%d/oom_adj", pid)
	content, err = ioutil.ReadFile(oomAdjPath)
	if err != nil {
		return nil, err
	}
	oomAdj, err := strconv.Atoi(string(content[:len(content)-1]))
	if err != nil {
		return nil, err
	}

	return &task{
		pid:    pid,
		name:   fields[1],
		rss:    rss,
		oomAdj: oomAdj,
	}, nil
}
