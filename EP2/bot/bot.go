package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func Reverse(str string) string {
	reversed_str := ""

	for _, v := range str {
		reversed_str = string(v) + reversed_str
	}

	return reversed_str
}

func main() {
	conn, err := net.Dial("tcp", "localhost:3000")
	fmt.Println("Connected!")
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan struct{})

	go func() {
		for {
			buff := make([]byte, 1024)
			bytesRead, err := conn.Read(buff)
			if err != nil {
				log.Fatal(err)
				conn.Close()
				break
			}
			msg_string := string(buff[:bytesRead])
			if strings.HasPrefix(msg_string, "(private)") {
				arr_string := strings.Split(msg_string, ": ")
				content := strings.ReplaceAll(arr_string[1], "\n", "")
				sender := strings.Split(arr_string[0], "(private) ")[1]
				fmt.Println("/msg " + sender + " " + Reverse(content))
				conn.Write([]byte("/msg " + sender + " " + Reverse(content)))
				conn.Write([]byte("\n"))
			}

		}
		done <- struct{}{}
	}()

	io.WriteString(conn, "/changenickname inversor\n")
	<-done
}
