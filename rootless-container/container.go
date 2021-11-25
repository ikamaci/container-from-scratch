package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("help")
	}
}

func run() {
	fmt.Printf("Running %v as pid %d\n", os.Args[2:], os.Getpid())
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID | // ==> new process ID given with isolated from host
			syscall.CLONE_NEWUSER | // ==> new user namespace for isolating user privileges
			syscall.CLONE_NEWNS, // ==> new mount namespace
		Credential: &syscall.Credential{Uid: 0, Gid: 0}, // ==> be the root in container
		UidMappings: []syscall.SysProcIDMap{
			// Uid 0 in the container(root) corresponds to my current userId in the Host
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			// Gid 0 in the container(root) corresponds to my current userId in the Host
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
	}
	must(cmd.Run())
}

func child() {
	fmt.Printf("Running %v as pid %d\n", os.Args[2:], os.Getpid())
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(syscall.Sethostname([]byte("container")))
	/**
	PART 2
	BEGIN
	*/
	must(syscall.Chroot("/home/kmc/chroot-ubuntu")) // container image
	must(syscall.Chdir("/"))

	must(syscall.Mount("proc", "proc", "proc", 0, ""))
	/**
	PART 2
	END
	*/
	must(cmd.Run())

	must(syscall.Unmount("proc", 0))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
