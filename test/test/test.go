package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	addrs := []string{}
	addrs = append(addrs, ":8081")
	addrs = append(addrs, ":8082")
	addrs = append(addrs, ":8083")
	addrs = append(addrs, ":8084")
	addrs = append(addrs, ":8085")
	addrs = append(addrs, ":8086")
	addrs = append(addrs, ":8087")
	addrs = append(addrs, ":8088")
	addrs = append(addrs, ":8089")
	addrs = append(addrs, ":9999")

	// 将对象编码为 JSON 格式
	jsonData, err := json.MarshalIndent(addrs, "", "    ")
	if err != nil {
		fmt.Println("JSON 编码失败:", err)
		return
	}

	// 将 JSON 数据写入文件
	file, err := os.Create("addrs.json")
	if err != nil {
		fmt.Println("创建文件失败:", err)
		return
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("写入文件失败:", err)
		return
	}

	fmt.Println("JSON 文件生成成功")
}
