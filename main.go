package main

import (
	_ "nfc-write-client/env" // 设置env，必须第一个导入
	"nfc-write-client/views"
)

func main() {
	views.NewViews()
}
