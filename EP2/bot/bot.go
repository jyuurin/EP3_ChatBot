package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

// funcao de inverter o input do usuario.
func reverse(input string) string {
	runes := []rune(input)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func main() {
	conn, err := net.Dial("tcp", "localhost:3000")
	fmt.Println("Connected!")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	go func() {
		reader := bufio.NewReader(conn)
		for {
			msg_string, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
				break
			}
			//aqui trata o input do usuario e chama a função de inverter.
			msg_string = strings.TrimSpace(msg_string)
			if strings.HasPrefix(msg_string, "/inverteMsg") {
				reversedText := reverse(strings.TrimPrefix(msg_string, "/inverteMsg"))
				fmt.Fprint(conn, "Inverti! "+reversedText)
			}
		}
	}()
	select {}
}
