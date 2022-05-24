package main

import (
	"bufio"
	"log"
	"net"
	"os"
)

func main() {
	log.Println("Client 端連線中～")

	// 建立網擄連線
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Println("網路連線錯誤")
		os.Exit(1)
	}
	log.Println("客戶端成功連嫌到服務端，服務端網址：", conn.RemoteAddr())

	go sendMessage(conn)

	// 接收消息
	buffer := make([]byte, 1024)
	for {
		// 讀取消息
		nums, err := conn.Read(buffer)
		if err != nil {
			log.Println("讀取訊息錯誤，登出中")
			os.Exit(1)
		}
		log.Println("收到訊息：", string(buffer[:nums]))
	}
}

func sendMessage(conn net.Conn) {
	// 循環發送
	for {
		// 讀取鍵盤輸入
		reader := bufio.NewReader(os.Stdin)

		// 讀取一行
		data, _, _ := reader.ReadLine()

		// 發送輸入的字串
		_, err := conn.Write(data)
		if err != nil {
			conn.Close()
			log.Println("Error:", err, "客戶端關閉")
			os.Exit(0)
		}
		log.Println("發送消息：", string(data))

		if string(data) == "exit" {
			log.Println("客戶端關閉")
			os.Exit(0)
		}
	}
}
