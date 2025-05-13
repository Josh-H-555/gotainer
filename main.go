package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

// docker run image <cmd> <params>
// desired equivalence
// go run main.go run <cmd> <params>

func main() {
	switch os.Args[1] {
	case "run":
		run()

	case "child":
		child()

	default:
		panic("bad command")
	}
}

// only works on Linux
func run() {
	fmt.Printf("Running %v\n", os.Args[2:], os.Getpid())

	// sets up the commands
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)

	// adds input/output/error
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// setup namespaces
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	cmd.Run()
}

// only works on Linux
func child() {
	// set this to a mounted file system
	new_root := "/temp"

	fmt.Printf("Running %v\n", os.Args[2:], os.Getpid())

	syscall.Sethostname([]byte("container"))

	// set the new root
	syscall.Chroot(new_root)
	syscall.Chdir("/")

	// mount /proc for process visibility
	syscall.Mount("proc", "proc", "proc", 0, "")

	// sets up the commands
	cmd := exec.Command(os.Args[2], os.Args[3:]...)

	// adds input/output/error
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()

	syscall.Unmount("/proc", 0)
}

func cg() {
	// setup cgroup information for the container
	cgroups := "/sys/fs/cgroup"
	pids := filepath.Join(cgroups, "pids")
	err := os.Mkdir(filepath.Join(pids, "container"), 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	// write the cgroup information
	must(os.WriteFile(filepath.Join(pids, "container/pids.max"), []byte("20"), 0700))
	must(os.WriteFile(filepath.Join(pids, "container/notify_on_release"), []byte("1"), 0700))
	must(os.WriteFile(filepath.Join(pids, "container/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
