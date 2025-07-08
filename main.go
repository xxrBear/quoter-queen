package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/joho/godotenv"
)

func loadConfig() (string, string, string) {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("加载.env文件失败:", err)
	}
	username := os.Getenv("EMAIL_USERNAME")
	pwdBase64 := os.Getenv("EMAIL_PASSWORD")
	server := os.Getenv("EMAIL_SERVER")

	pwdBytes, err := base64.StdEncoding.DecodeString(pwdBase64)
	if err != nil {
		log.Fatal("密码Base64解码失败:", err)
	}
	password := string(pwdBytes)
	return username, password, server
}

func connect(server string) *client.Client {
	c, err := client.DialTLS(server, nil)
	if err != nil {
		log.Fatal("连接失败:", err)
	}
	fmt.Println("已连接:", server)
	return c
}

func fetchRecentEmails(c *client.Client, folder string, maxFetch uint32) {
	mbox, err := c.Select(folder, false)
	if err != nil {
		log.Fatal("选择邮箱失败:", err)
	}
	if mbox.Messages == 0 {
		fmt.Println("没有邮件")
		return
	}

	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > maxFetch {
		from = mbox.Messages - maxFetch + 1
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddRange(from, to)

	section := &imap.BodySectionName{}
	messages := make(chan *imap.Message, 10)

	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqSet, []imap.FetchItem{imap.FetchEnvelope, section.FetchItem()}, messages)
	}()

	fmt.Println("最近邮件：")
	for msg := range messages {
		if msg == nil {
			continue
		}
		fmt.Println("--------")
		fmt.Println("主题:", msg.Envelope.Subject)
		if len(msg.Envelope.From) > 0 {
			fmt.Println("发件人:", msg.Envelope.From[0].Address())
		} else {
			fmt.Println("发件人: 无")
		}
		// 邮件正文可在这里解析
	}

	if err := <-done; err != nil {
		log.Fatal("拉取邮件失败:", err)
	}
}

func main() {
	username, password, server := loadConfig()
	fmt.Printf("用户名: [%s]\n", username)
	// 不打印密码，安全考虑

	c := connect(server)
	defer c.Logout()

	if err := c.Login(username, password); err != nil {
		log.Fatal("登录失败:", err)
	}
	fmt.Println("登录成功:", username)

	fetchRecentEmails(c, "银行询价", 5)
}
