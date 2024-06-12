package main

import (
	"log"
	_ "nfc-write-client/env" // 设置env，必须第一个导入
	"nfc-write-client/views"
)

func main() {
	log.Println("Hello, World!")
	views.NewViews()
}
