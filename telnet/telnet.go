package telnet

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	PROT          = "tcp"
	TCP_PORT      = ":telnet" // Порт telnet = 23
	BUFFER        = 4096      // byte
	DIAL_TIMEOUT  = 1         // seconds
	DEAD_TIMEOUT  = 15        // seconds
	WRITE_TIMEOUT = 200       // milliseconds
)

type Auth struct {
	Login    string
	Password string
	Enable   string
}

type Node struct {
	Hostname string
	Addr     string
	Auth     Auth
}

type Host struct {
	conn net.Conn
	node Node
}

func Telnet(jsonData []byte, cmdSlice []string) (map[string]string, error) {
	// Разбираем json, сохраняем в nodes
	var nodes []Node
	if err := json.Unmarshal(jsonData, &nodes); err != nil {
		return nil, err
	}

	connections := make(chan Host, len(nodes))
	wg := &sync.WaitGroup{}
	for _, node := range nodes {
		wg.Add(1)
		go func(c chan Host, n Node, wg *sync.WaitGroup) {
			defer wg.Done()
			conn, err := connect(n.Addr + TCP_PORT)
			if err != nil {
				fmt.Println(err)
				return
			}
			connections <- Host{conn, n}
		}(connections, node, wg)
	}
	wg.Wait()
	close(connections)

	resultChan := make(chan [2]string, len(connections)*len(cmdSlice))
	for conn := range connections {
		wg.Add(1)
		go func(c Host, comms []string, wg *sync.WaitGroup) {
			defer wg.Done()
			if err := login(c.conn, c.node.Auth); err != nil {
				fmt.Println(err.Error())
				return
			}
			for _, cmd := range comms {
				res, err := exec(c.conn, cmd)
				if err != nil {
					fmt.Println(err.Error())
					continue
				}
				resultChan <- [2]string{c.conn.RemoteAddr().String(), res}
			}
		}(conn, cmdSlice, wg)
	}
	wg.Wait()
	close(resultChan)

	// сохраняем результат из resultChan []chan в resultMap map[IP]command_output
	var resultMap = make(map[string]string)
	for i := range resultChan {
		if _, ok := resultMap[i[0]]; ok {
			resultMap[i[0]] += i[1]
		} else {
			resultMap[i[0]] = i[1]
		}
	}
	return resultMap, nil
}

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
