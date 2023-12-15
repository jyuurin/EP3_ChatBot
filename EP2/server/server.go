package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

//definicao de tipos e canais

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

// canais
var (
	entering = make(chan User)
	leaving  = make(chan User)
	editing  = make(chan User)
	messages = make(chan Message)
)

func broadcaster() {
	//dois mapas: clients pega os clientes conectados
	//users: armazena informações do user, e o IP é a chave
	clients := make(map[User]bool) // todos os clientes conectados
	users := make(map[string]User)

	//loop para lidar com as mensagens
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
			} else { //envio para um só usuario/mensagem privada
				for i := range users {
					user := users[i]

					if strings.ToUpper(user.nickname) == strings.ToUpper(msg.destiny) {
						user.channel <- msg.text
					}
				}
			}
		//trata um novo client no chat: ele é add ao mapa de clients e users.
		case cli := <-entering:
			clients[cli] = true
			users[cli.ip] = cli

		//edição do nick do usuario, que também é atualizado no mapa users.
		case edited_cli := <-editing:
			users[edited_cli.ip] = edited_cli
			users[edited_cli.ip].channel <- "Seu apelido é: " + edited_cli.nickname

		//caso o client sai do chat, ele é removido de ambos os mapas e a conexão termina.
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

// função que processa os inputs do usuario:
func processInput(user User, inputText string) {
	switch {
	case strings.HasPrefix(inputText, "/changenickname"):
		newNick := strings.TrimPrefix(inputText, "/changenickname ")
		user.nickname = newNick
		messages <- Message{sender: "server", text: fmt.Sprintf("%s mudou de nome para %s", user.nickname, newNick), destiny: "all"}
		entering <- user

	case strings.HasPrefix(inputText, "/msg"):
		parts := strings.Fields(inputText)
		if len(parts) >= 3 {
			destiny := parts[1]
			message := strings.Join(parts[2:], " ")
			messages <- Message{sender: user.nickname, text: fmt.Sprintf("(private) %s: %s", user.nickname, message), destiny: destiny}
		}

	case strings.HasPrefix(inputText, "/quit"), strings.HasPrefix(inputText, "/q"):
		messages <- Message{sender: "server", text: fmt.Sprintf("%s deixou a sala.", user.nickname), destiny: "all"}

	case strings.HasPrefix(inputText, "/checkip"), strings.HasPrefix(inputText, "/ip"):
		messages <- Message{sender: "server", text: user.ip, destiny: user.ip}

	default:
		messages <- Message{sender: user.nickname, text: fmt.Sprintf("%s: %s", user.nickname, inputText), destiny: "all"}
	}
}

// função lida com a conexão dos clients.
func handleConn(conn net.Conn) {
	ch := make(chan string)
	user := User{ip: conn.RemoteAddr().String(), nickname: conn.RemoteAddr().String(), channel: ch, conn: conn}

	go clientWriter(conn, ch)

	//mensagem quando o client entra com seu apelido.
	user.channel <- "vc é " + user.nickname
	//mensagem para os demais clients de que o usuario x chegou.
	messages <- Message{sender: user.nickname, text: user.nickname + " chegou!", destiny: "all"}
	entering <- user

	//le a entrada do client e processa tudo pela função processInput
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
