package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// 这是一个简单的测试示例，演示如何使用 URL 短链接服务
func main() {
	baseURL := "http://localhost:8080"
	
	// 示例：创建一个短链接
	fmt.Println("正在创建短链接...")
	
	client := &http.Client{Timeout: 10 * time.Second}
	
	// 创建短链接请求
	reqBody := map[string]string{
		"url": "https://www.example.com/very/long/url/with/many/parameters?param1=value1&param2=value2&param3=value3",
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	
	resp, err := client.Post(baseURL+"/api/shorten", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return
	}
	
	if resp.StatusCode == http.StatusOK {
		fmt.Printf("短链接创建成功!\n响应: %s\n", string(body))
	} else {
		fmt.Printf("创建失败，状态码: %d, 响应: %s\n", resp.StatusCode, string(body))
	}
}