package main

import (
	"bufio"
	"errors"
	"net"
	"strings"
	"time"
)

const PROT = "tcp"
const BUFFER = 4096       // byte
const DIAL_TIMEOUT = 1    // seconds
const DEAD_TIMEOUT = 15   // seconds
const WRITE_TIMEOUT = 200 // milliseconds

func connect(addr string) (net.Conn, error) {
	conn, err := net.DialTimeout(PROT, addr, DIAL_TIMEOUT*time.Second)
	if err != nil {
		return conn, err
	}
	err = conn.SetDeadline(time.Now().Add(DEAD_TIMEOUT * time.Second))
	if err != nil {
		return conn, err
	}
	return conn, err
}

func write(conn net.Conn, bufs []byte) (int, error) {
	n, err := conn.Write(bufs)
	if err != nil {
		return n, err
	}
	time.Sleep(WRITE_TIMEOUT * time.Millisecond)
	return n, err
}

func writeWord(conn net.Conn, word string) error {
	var buf [BUFFER]byte
	_, err := conn.Read(buf[0:])
	if err != nil {
		return err
	}
	_, err = write(conn, []byte(word+"\n"))
	if err != nil {
		return err
	}
	return err
}

func login(conn net.Conn, id Auth) error {
	var buf [BUFFER]byte
	err := writeWord(conn, id.Login)
	if err != nil {
		return err
	}
	err = writeWord(conn, id.Password)
	if err != nil {
		return err
	}
	n, err := conn.Read(buf[0:])
	if err != nil {
		return err
	}
	if strings.HasSuffix(string(buf[0:n]), ">") {
		_, err = write(conn, []byte("enable\n"))
		if err != nil {
			return err
		}
		err = writeWord(conn, id.Enable)
		if err != nil {
			return err
		}
		_, err = conn.Read(buf[0:])
		if err != nil {
			return err
		}
	}
	_, err = write(conn, []byte("terminal length 0\n"))
	if err != nil {
		return err
	}
	_, err = conn.Read(buf[0:])
	if err != nil {
		return err
	}
	return err
}

func exec(conn net.Conn, cmd string) (string, error) {
	var buf [BUFFER]byte
	_, err := write(conn, []byte(cmd+"\n"))
	if err != nil {
		return "", err
	}
	reader := bufio.NewReader(conn)
	if reader == nil {
		return "", errors.New("Create reader failed.")
	}
	var output = make([]byte, 0, 16*BUFFER)
	for {
		n, err := reader.Read(buf[0:])
		if err != nil {
			return "", err
		}
		output = append(output, buf[0:n]...)
		if strings.HasSuffix(string(buf[0:n]), "#") {
			break
		}
	}
	return string(output), err
}

