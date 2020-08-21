package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// Порт telnet = 23
	TCP_PORT = ":telnet"
	// Каталог для вывода результатов. По умолчанию текущая дата в формате YYYY-MM-DD
	OUTPUT_DIR = "qwe"
)

type Host struct {
	conn net.Conn
	node Node
}

func main() {
	invFile, cmdFile, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return
	}

	var nodes Nodes
	err = nodes.GetNodes(invFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return
	}

	var commands Commands
	err = commands.GetCommangs(cmdFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return
	}

	// Задаем каталог для вывода результатов выполнения команд на устройствах.
	// Имя каталога можно изменить в константе OUTPUT_DIR.
	// По умолчанию - текущая дата в формате YYYY-MM-DD
	curDir := OUTPUT_DIR
	if curDir == "" {
		curDir = time.Now().Format("2006-01-02")
	}
	// Создаем каталог для вывода результатов выполнения команд на устройствах.
	// Если каталог уже существует, к имени нового каталога после знака "_"
	// будет добавлено случайное число от 0 до 10000
	if err := makeDir(curDir); err != nil {
		log.Fatal(err)
	}

	start := time.Now()

	connections := make(chan Host, len(nodes.Nodes))
	wg := &sync.WaitGroup{}
	for _, node := range nodes.Nodes {
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

	var mu sync.Mutex
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

				fileName := c.conn.RemoteAddr().String()
				fileName = strings.Split(fileName, ":")[0] + ".txt"
				fileName = filepath.Join(curDir, fileName)

				mu.Lock()
				f, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
				if err != nil {
					fmt.Println(err.Error())
					mu.Unlock()
					continue
				}
				_, err = f.WriteString(res)
				if err != nil {
					fmt.Println(err.Error())
					f.Close()
					mu.Unlock()
					continue
				}
				f.Close()
				mu.Unlock()
				runtime.Gosched()
			}
		}(conn, commands.Cmd, wg)
	}
	wg.Wait()

	duration := time.Since(start)
	fmt.Printf("Duration: %d ms\n", duration.Milliseconds())
}

func parseArgs() (string, string, error) {
	if len(os.Args) < 3 {
		return "", "", errors.New("Usage: telcl.exe inventory_file.json command_file.txt\n")
	}
	return os.Args[1], os.Args[2], nil
}

func makeDir(d string) (err error) {
	if err = os.Mkdir(d, os.ModeDir); err != nil {
		if !os.IsExist(err) {
			return
		}
		rand.Seed(time.Now().UTC().UnixNano())
		solt := rand.Intn(10000)
		return makeDir(d + "_" + strconv.Itoa(solt))
	}
	return
}
