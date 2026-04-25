package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const serverAddr = "127.0.0.1:9090"

func receiveMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("\n[!] Соединение с сервером разорвано.")
			os.Exit(0)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "FROM:"):
			rest := line[5:]
			idx := strings.Index(rest, ":")
			if idx != -1 {
				sender := rest[:idx]
				text := rest[idx+1:]
				fmt.Printf("\n[Сообщение от %s]: %s\n", sender, text)
			}
		case strings.HasPrefix(line, "ERR:"):
			fmt.Printf("[Ошибка] %s\n", line[4:])
		case strings.HasPrefix(line, "OK:"):
			info := line[3:]
			if !strings.Contains(info, "доставлено") {
				fmt.Printf("[Статус] %s\n", info)
			}
		}
	}
}

func main() {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Printf("Не удалось подключиться к серверу %s. Убедитесь, что сервер запущен.\n", serverAddr)
		os.Exit(1)
	}
	defer conn.Close()

	stdinReader := bufio.NewReader(os.Stdin)

	fmt.Print("Введите ваш ник: ")
	nick, _ := stdinReader.ReadString('\n')
	nick = strings.TrimSpace(nick)
	if nick == "" {
		fmt.Println("Ник не может быть пустым.")
		os.Exit(1)
	}

	fmt.Fprintf(conn, "NICK:%s\n", nick)

	serverReader := bufio.NewReader(conn)
	response, err := serverReader.ReadString('\n')
	if err != nil {
		fmt.Println("Ошибка соединения с сервером.")
		os.Exit(1)
	}
	response = strings.TrimSpace(response)

	if strings.HasPrefix(response, "ERR:") {
		fmt.Printf("[Ошибка] %s\n", response[4:])
		os.Exit(1)
	}
	if strings.HasPrefix(response, "OK:") {
		fmt.Printf("[Статус] %s\n", response[3:])
	}

	go receiveMessages(conn)

	fmt.Println("Для отправки сообщения введите: <ник получателя> <сообщение>")
	fmt.Println("Для выхода нажмите Ctrl+C")
	fmt.Println()

	for {
		input, err := stdinReader.ReadString('\n')
		if err != nil {
			break
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		parts := strings.SplitN(input, " ", 2)
		if len(parts) < 2 {
			fmt.Println("[Подсказка] <ник получателя> <сообщение>")
			continue
		}
		target := parts[0]
		text := parts[1]

		fmt.Fprintf(conn, "MSG:%s:%s\n", target, text)
	}
}
