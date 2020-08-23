package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"telgo/telnet"
	"time"
)

// Каталог для вывода результатов. По умолчанию текущая дата в формате YYYY-MM-DD
const OUTPUT_DIR = ""

func main() {
	// Разбор аргументов командной строки
	// Формат: telcl.exe inventory_file.json command_file.txt
	// Переменная invFile - inventory_file.json
	// Формат:
	// [
	// {
	//     "hostname":"R1",
	//     "addr":"192.168.129.100",
	//     "auth": {
	//         "login": "cisco",
	//         "password": "cisco",
	//         "enable": "cisco"
	//     }
	// },
	// ..
	// Переменная cmdFile - command_file.txt
	// Файл с командами cisco, разделенных по строкам
	if len(os.Args) < 3 {
		log.Fatal("Usage: telcl.exe inventory_file.json command_file.txt")
	}
	invFile, cmdFile := os.Args[1], os.Args[2]
	// Читаем invFile и сохраняем []byte в jsonData
	jsonData, err := ioutil.ReadFile(invFile)
	if err != nil {
		log.Fatal(err)
	}
	// Читаем cmdFile и сохраняем []byte в cmdData
	cmdData, err := ioutil.ReadFile(cmdFile)
	if err != nil {
		log.Fatal(err)
	}
	// Разбиваем cmdData на строки и сохраняем как []string в cmdSlice
	cmdSlice := strings.Split(string(cmdData), "\n")
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
	// Выполняем команды из cmdSlice на устройствах jsonData,
	// резльтат rMap типа map[IP]command_output
	rMap, err := telnet.Telnet(jsonData, cmdSlice)
	if err != nil {
		log.Fatal(err)
	}
	// Сохраняем результат в файлы с именами по ip-адресам устройств
	for k, v := range rMap {
		fileName := strings.Split(k, ":")[0] + ".txt"
		fileName = filepath.Join(curDir, fileName)
		f, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		_, err = f.WriteString(v)
		if err != nil {
			fmt.Println(err.Error())
			f.Close()
			continue
		}
		f.Close()
	}
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
