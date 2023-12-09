package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type client chan<- string // canal de mensagem

type User struct {
	ip       string
	nickname string
	channel  chan<- string
	conn     net.Conn
}

type Message struct {
	sender  string
	destiny string
	text    string
}

var (
	entering = make(chan User)
	leaving  = make(chan User)
	editing  = make(chan User)
	messages = make(chan Message)
)

func broadcaster() {
	clients := make(map[User]bool) // todos os clientes conectados
	users := make(map[string]User)

	for {
		select {
		case msg := <-messages:
			// broadcast de mensagens. Envio para todos
			if msg.destiny == "all" {
				for i := range users {
					user := users[i]
					if strings.ToUpper(user.nickname) != strings.ToUpper(msg.sender) {
						user.channel <- msg.text
					}
				}
			} else { //msg privada
				for i := range users {
					user := users[i]

					if strings.ToUpper(user.nickname) == strings.ToUpper(msg.destiny) {
						user.channel <- msg.text
					}
				}
			}

		case cli := <-entering:
			clients[cli] = true
			users[cli.ip] = cli

		case edited_cli := <-editing:
			users[edited_cli.ip] = edited_cli
			users[edited_cli.ip].channel <- "Seu apelido é: " + edited_cli.nickname

		case user := <-leaving:
			delete(clients, user)
			delete(users, user.ip)
			close(user.channel)
			user.conn.Close()
		}
	}
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func processInput(user User, inputText string) {
	if strings.HasPrefix(inputText, "/mudarapelido") {
		newNick := strings.TrimPrefix(inputText, "/mudarapelido")
		messages <- Message{sender: "server", text: user.nickname + " mudou para: " + newNick, destiny: "all"}
		user.nickname = newNick
		entering <- user
	} else if strings.HasPrefix(inputText, "/mensagem") {
		parts := strings.SplitN(inputText, " ", 3)
		destiny := parts[1]
		message := parts[2]
		messages <- Message{sender: user.nickname, text: "(privado) " + user.nickname + ": " + message, destiny: destiny}
	} else if strings.HasPrefix(inputText, "/sair") || strings.HasPrefix(inputText, "/s") {
		messages <- Message{sender: "server", text: user.nickname + " saiu da sala.", destiny: "all"}
	} else if strings.HasPrefix(inputText, "/checarip") || strings.HasPrefix(inputText, "/ip") {
		messages <- Message{sender: "server", text: user.ip, destiny: user.ip}
	} else {
		messages <- Message{sender: user.nickname, text: user.nickname + ": " + inputText, destiny: "all"}
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	user := User{ip: conn.RemoteAddr().String(), nickname: conn.RemoteAddr().String(), channel: ch, conn: conn}

	go clientWriter(conn, ch)

	user.channel <- "vc é " + user.nickname
	messages <- Message{sender: user.nickname, text: user.nickname + " chegou!", destiny: "all"}
	entering <- user

	input := bufio.NewScanner(conn)
	for input.Scan() {
		processInput(user, input.Text())
	}

	leaving <- user
	messages <- Message{sender: user.nickname, text: user.nickname + " se foi!", destiny: "all"}
	conn.Close()
}

func main() {
	fmt.Println("Iniciando servidor...")
	listener, err := net.Listen("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}
