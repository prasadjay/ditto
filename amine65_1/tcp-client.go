package main

import "net"
import "fmt"
import "bufio"

//import "strings" // only needed below for sample processing
import "os"

//import "time"

func main() {
	clients = make(map[string]net.Conn)
	go StartServer()
	StartClient()
}

var clients map[string]net.Conn

func StartServer() {
	fmt.Println("Launching server...")

	// listen on all interfaces
	ln, _ := net.Listen("tcp", ":8081")

	// accept connection on port
	conn, _ := ln.Accept()

	// run loop forever (or until ctrl-c)
	for {
		fmt.Println(conn.RemoteAddr())
		// will listen for message to process ending in newline (\n)
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err == nil {
			// output message received
			fmt.Print("Message Received:", string(message))
			// sample process for string received
			//newmessage := strings.ToUpper(message)
			// send new string back to client
			//conn.Write([]byte(newmessage + "\n"))
		} else {
			fmt.Println(err.Error())
		}
	}
}

func StartClient() {
	/*// connect to this socket
	conn, _ := net.Dial("tcp", "127.0.0.1:8081")
	for {
		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Text to send: ")
		text, _ := reader.ReadString('\n')
		// send to socket
		fmt.Fprintf(conn, text+"\n")
		// listen for reply
		//message, _ := bufio.NewReader(conn).ReadString('\n')
		//fmt.Print("Message from server: " + message)
	}*/

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if clients["127.0.0.1"] == nil {
			conn, err := net.Dial("tcp", "127.0.0.1:8082")
			//defer conn.Close()
			if err == nil {
				clients["127.0.0.1"] = conn
				if conn != nil {
					text := scanner.Text() + "\n"
					fmt.Fprintf(conn, text)
				} else {
					fmt.Println("Nil")
				}
			} else {
				fmt.Println(err.Error())
			}
		} else {
			text := scanner.Text() + "\n"
			fmt.Fprintf(clients["127.0.0.1"], text)
		}

	}

	if scanner.Err() != nil {
		// handle error.
		fmt.Println(scanner.Err().Error())
	}
}
