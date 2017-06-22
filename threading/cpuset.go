package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

const (
	cpusetRoot  = "/sys/fs/cgroup/cpuset"
	cgroupProcs = "cgroup.procs"
	cgroupTasks = "tasks"
	cpusetCPUs  = "cpuset.cpus"
	cpusetMems  = "cpuset.mems"

	defaultDirPerm  = 0755
	defaultFilePerm = os.FileMode(0)
)

type cpuset struct {
	name string
	path string
	cpu  int
	node int
}

func (c *cpuset) create() error {
	c.path = filepath.Join(cpusetRoot, c.name)

	if err := os.MkdirAll(c.path, defaultDirPerm); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(c.path, cpusetCPUs), []byte(strconv.Itoa(c.cpu)), defaultFilePerm); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(c.path, cpusetMems), []byte(strconv.Itoa(c.node)), defaultFilePerm); err != nil {
		return err
	}

	return nil
}

func (c *cpuset) delete() error {
	return os.RemoveAll(c.path)
}

func (c *cpuset) addThread(tid int) error {
	procs := filepath.Join(c.path, cgroupTasks)
	
	return ioutil.WriteFile(procs, []byte(strconv.Itoa(tid)), defaultFilePerm)
}

func (c *cpuset) dump() {
	procs, err := ioutil.ReadFile(filepath.Join(c.path, cgroupProcs))
	if err != nil {
		return
	}

	cpus, err := ioutil.ReadFile(filepath.Join(c.path, cpusetCPUs))
	if err != nil {
		return
	}

	mems, err := ioutil.ReadFile(filepath.Join(c.path, cpusetMems))
	if err != nil {
		return
	}

	tasks, err := ioutil.ReadFile(filepath.Join(c.path, "tasks"))
	if err != nil {
		return
	}

	fmt.Printf("*** %s ***\n", c.path)
	fmt.Printf("\tcgroup procs: %s", string(procs))
	fmt.Printf("\ttasks: %s", string(tasks))
	fmt.Printf("\tcpuset CPUs: %s", string(cpus))
	fmt.Printf("\tcpuset memory node: %s", string(mems))
}
