package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

func handleConnection(c net.Conn) {
	fmt.Printf("Serving MME %s\n", c.RemoteAddr().String())
	netData, err := bufio.NewReader(c).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Got ", netData, " from MME")
	result := "success\n"
	c.Write([]byte(string(result)))
	c.Close()
}

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide a port number!")
		return
	}

	PORT := ":" + arguments[1]
	l, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	rand.Seed(time.Now().Unix())

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c)
	}
}
