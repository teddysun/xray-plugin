// +build android

package main

import "C"

import (
	"log"
	"syscall"

	vinternet "github.com/xtls/xray-core/transport/internet"
)

func ControlOnConnSetup(network string, address string, s uintptr) error {
	fd := int(s)
	path := "protect_path"

	socket, err := syscall.Socket(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		log.Println(err)
		return err
	}

	defer syscall.Close(socket)

	C.set_timeout(C.int(socket))

	err = syscall.Connect(socket, &syscall.SockaddrUnix{Name: path})
	if err != nil {
		log.Println(err)
		return err
	}

	C.ancil_send_fd(C.int(socket), C.int(fd))

	dummy := []byte{1}
	n, err := syscall.Read(socket, dummy)
	if err != nil {
		log.Println(err)
		return err
	}
	if n != 1 {
		log.Println("Failed to protect fd: ", fd)
	}

	return nil
}

func registerControlFunc() {
	vinternet.RegisterDialerController(ControlOnConnSetup)
	vinternet.RegisterListenerController(ControlOnConnSetup)
}
