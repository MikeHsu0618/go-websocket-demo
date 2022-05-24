package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

var (
	onlineConns = make(map[string]net.Conn)
	msgChan     = make(chan string, 1024)
	quitChan    = make(chan string, 1024)
)

func main() {
	log.Println("Server 啟動中～")

	listenSocket, err := net.Listen("tcp", "127.0.0.1:8080")
	errorCheck(err)
	defer listenSocket.Close()

	go consumeMessage()

	// 啟動連接
	for {
		conn, err := listenSocket.Accept()
		errorCheck(err)

		// 添加到全局變量
		onlineConns[conn.RemoteAddr().String()] = conn

		// 遍歷每一個連接
		fmt.Println("客戶端列表：\n =============")
		for i := range onlineConns {
			log.Println(i)
		}

		go receiveInfo(conn)
	}
}

func consumeMessage() {
	for {
		select {
		case msg := <-msgChan:
			processMessage(msg)
		case <-quitChan:
			break
		}
	}
}

func processMessage(msg string) {
	contents := strings.Split(msg, "#")
	if len(contents) > 1 {
		content := contents[1]

		if content == "list" {
			var str string
			for i := range onlineConns {
				str += "||||" + i
			}

			// 連接
			conn, ok := onlineConns[contents[0]]
			if !ok {
				log.Fatal("連線失敗")
			}
			_, err := conn.Write([]byte(str))
			if err != nil {
				log.Fatal("發送失敗", err)
			}
			log.Println("發送消息：", str, "目的地：", conn.RemoteAddr())

			return
		}

		// 連接
		for _, conn := range onlineConns {
			_, err := conn.Write([]byte(content))
			if err != nil {
				log.Fatal("發送失敗", err)
			}
			log.Println("發送消息：", content, "目的地：", conn.RemoteAddr())
		}
	}
}

func receiveInfo(conn net.Conn) {
	// 緩衝
	buffer := make([]byte, 1024)

	// 循環讀取
	for {
		// 讀取數據
		nums, err := conn.Read(buffer)
		if err != nil {
			break
		}

		if nums == 0 {
			continue
		}

		content := string(buffer[:nums])
		log.Println("收到消息：", content, "來自：", conn.RemoteAddr())

		if content == "exit" {
			log.Println("客戶端：", conn.RemoteAddr(), "正在退出....")
			clientExit(conn)
		}

		msgChan <- conn.RemoteAddr().String() + "#" + content
	}
}

func clientExit(conn net.Conn) {
	delete(onlineConns, conn.RemoteAddr().String())

	conn.Close()

	// 輸出當前列表
	log.Println("客戶列表： \n ============")

	for i := range onlineConns {
		fmt.Println(i)
	}
}

func errorCheck(err error) {
	if err != nil {
		log.Fatal("Error :", err)
		os.Exit(1)
	}
}
