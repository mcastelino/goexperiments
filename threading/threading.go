package main

/*
#define _GNU_SOURCE
#include <sched.h>
#include <pthread.h>
#include <signal.h>
#include <unistd.h>
#include <stdio.h>
#include <sys/types.h>
#include <sys/syscall.h>

int
checkAffinity(int cpu)
{
	int ret;
	cpu_set_t cpuset, desired_cpuset;


	CPU_ZERO(&desired_cpuset);
	CPU_SET(cpu , &desired_cpuset );

	CPU_ZERO(&cpuset);
	ret = sched_getaffinity(0, sizeof(cpuset), &cpuset);
	if (ret != 0) {
		return -1;
	}

	if (CPU_EQUAL(&cpuset, &desired_cpuset) == 0) {
		return -1;
	}
	return 0;
}

*/
import "C"

import (
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"time"
)

func myThread(name string, wg *sync.WaitGroup, cpu int, setAffinity bool, lockGoThread bool, expectLock bool) {
	defer wg.Done()

	if lockGoThread {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}

	tid := syscall.Gettid()
	fmt.Printf("%s is running on thread ID %d\n", name, tid)

	if setAffinity {
		cpuset := cpuset{
			name: "thread-" + name,
			cpu: cpu,
			node: 0,
		}

		if err := cpuset.create(); err != nil {
			fmt.Println(err)
			return
		}
		defer cpuset.delete()

		if err := cpuset.addThread(tid); err != nil {
			fmt.Println(err)
			return
		}

//		cpuset.dump()
	}

	time.Sleep(1 * time.Second)

	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)

		if C.checkAffinity(C.int(cpu)) != 0 {
			if expectLock {
				fmt.Printf("FATAL: Affinity failure [%v] Thread ID [%s] := %v \n", i, name, tid)
			} else {
				fmt.Printf("OK: Expected affinity failure [%v] Thread ID [%s] := %v \n", i, name, tid)
			}
		}
	}

}

func main() {
	var wg sync.WaitGroup
	tid := syscall.Gettid()
	fmt.Printf("My thread id: %d\n", tid)
	wg.Add(4)

	go myThread("locked-affinity", &wg, 1, true, true, true)
	go myThread("locked-no-affinity", &wg, 2, false, true, false)
	go myThread("unlocked-affinity", &wg, 3, true, false, false)
	go myThread("unlocked-no-affinity", &wg, 0, false, false, false)
	wg.Wait()
}
