package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

const addr = "127.0.0.1:9090"

var (
	clients = make(map[string]net.Conn)
	mu      sync.Mutex
)

func handleClient(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	line = strings.TrimSpace(line)

	if !strings.HasPrefix(line, "NICK:") {
		fmt.Fprintln(conn, "ERR:Неверный протокол")
		return
	}

	nick := line[5:]
	if nick == "" {
		fmt.Fprintln(conn, "ERR:Ник не может быть пустым")
		return
	}

	mu.Lock()
	if _, exists := clients[nick]; exists {
		mu.Unlock()
		fmt.Fprintln(conn, "ERR:Ник уже занят")
		return
	}
	clients[nick] = conn
	mu.Unlock()

	fmt.Fprintf(conn, "OK:Подключён как %s\n", nick)
	fmt.Printf("[+] Подключился: %s\n", nick)

	defer func() {
		mu.Lock()
		delete(clients, nick)
		mu.Unlock()
		fmt.Printf("[-] Отключился: %s\n", nick)
	}()

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)

		if !strings.HasPrefix(line, "MSG:") {
			fmt.Fprintln(conn, "ERR:Неизвестная команда")
			continue
		}

		rest := line[4:]
		idx := strings.Index(rest, ":")
		if idx == -1 {
			fmt.Fprintln(conn, "ERR:Неверный формат. Используйте: MSG:<ник>:<текст>")
			continue
		}

		targetNick := rest[:idx]
		text := rest[idx+1:]

		mu.Lock()
		targetConn, found := clients[targetNick]
		mu.Unlock()

		if !found {
			fmt.Fprintf(conn, "ERR:Пользователь \"%s\" не найден\n", targetNick)
			continue
		}

		_, sendErr := fmt.Fprintf(targetConn, "FROM:%s:%s\n", nick, text)
		if sendErr != nil {
			fmt.Fprintf(conn, "ERR:Не удалось доставить сообщение пользователю \"%s\"\n", targetNick)
			continue
		}
		fmt.Fprintln(conn, "OK:Сообщение доставлено")
	}
}

func main() {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Ошибка запуска сервера: %v\n", err)
		return
	}
	defer ln.Close()

	fmt.Printf("Сервер запущен на %s\n", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Ошибка подключения: %v\n", err)
			continue
		}
		go handleClient(conn)
	}
}
