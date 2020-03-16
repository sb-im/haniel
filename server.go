package main

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"sync"
	"time"
)

type SocketServer struct {
	links  []*net.Conn
	lock   sync.Mutex
	logger *log.Logger
	input  chan []byte
	output chan []byte
}

func (this *SocketServer) Listen(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		this.logger.Println("Listener:", addr, err)
	}
	defer listener.Close()

	go this.send(this.input)

	for {
		conn, err := listener.Accept()
		this.addLink(&conn)
		this.logger.Println("New Connect:", &conn)
		if err != nil {
			this.logger.Println("Connect Err:", &conn, err)
		} else {
			go this.recv(&conn, this.output)
		}
	}
}

func (this *SocketServer) Client(addr string) {
	for {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			time.Sleep(3 * time.Second)
		} else {
			this.logger.Println("New connect", &conn)

			this.addLink(&conn)
			go this.send(this.input)

			this.recv(&conn, this.output)
			this.logger.Println("Connect err try reconnect")
		}
	}
}

func (this *SocketServer) addLink(link *net.Conn) {
	this.delLink(link)
	this.lock.Lock()
	this.links = append(this.links, link)
	this.lock.Unlock()
}

func (this *SocketServer) delLink(link *net.Conn) {
	this.lock.Lock()
	for index, run_link := range this.links {
		if run_link == link {

			// Order is not important
			this.links[index] = this.links[len(this.links)-1]
			this.links = this.links[:len(this.links)-1]
		}
	}
	this.lock.Unlock()
}

func (this *SocketServer) send(input chan []byte) {
	for msg := range input {
		for _, conn := range this.links {
			this.logger.Println("Send:", string(msg))
			n, err := (*conn).Write(msg)
			if err != nil {
				this.logger.Println("Write Err:", conn, n, err)
			}
		}
	}
}

func (this *SocketServer) recv(conn *net.Conn, output chan []byte) {
	defer func() {
		this.delLink(conn)
		(*conn).Close()
	}()
	read := bufio.NewReader(*conn)
	for {
		raw, err := readLine(read)
		if err != nil {
			this.logger.Println("Connect Err:", conn, err)
			break
		}

		this.logger.Println("Recv:", conn, string(raw))
		output <- raw
	}
}

// This function mainly solves the case where the number of bytes in a single line is greater than 4096
func readLine(reader *bufio.Reader) ([]byte, error) {
	var buffer bytes.Buffer
	for {
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			return buffer.Bytes(), err
		}
		buffer.Write(line)
		if !isPrefix {
			break
		}
	}
	return buffer.Bytes(), nil
}
