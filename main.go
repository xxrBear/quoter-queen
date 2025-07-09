package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/joho/godotenv"
)

func loadEnvConfig() (string, string, string) {
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

func fetchRecentEmails(c *client.Client, folder string) {
	mbox, err := c.Select(folder, true)
	if err != nil {
		log.Fatal("选择邮箱失败:", err)
	}

	if mbox.Messages == 0 {
		fmt.Println("没有邮件")
		return
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	today := time.Now().In(loc).Truncate(24 * time.Hour)
	criteria := imap.NewSearchCriteria()
	criteria.Since = today.UTC()

	// 搜索邮件
	ids, err := c.Search(criteria)
	if err != nil {
		log.Fatal("搜索邮件失败:", err)
	}
	if len(ids) == 0 {
		fmt.Println("今天没有邮件")
		return
	}

	// 构造 UID 集合
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(ids...)
	fmt.Println("搜索到 UID:", ids)

	// 设置邮件拉取项
	// section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope}
	fmt.Println("items:", len(items))

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.Fetch(seqSet, items, messages)
	}()

	fmt.Println(len(messages))

	// 读取邮件数据
	for msg := range messages {
		if msg == nil {
			continue
		}
		fmt.Println("--------")
		fmt.Println("标题:", msg.Envelope.Subject)
		fmt.Println("发件人:", msg.Envelope.From[0].Address())
		fmt.Println("时间:", msg.Envelope.Date)
	}

	if err := <-done; err != nil {
		log.Fatal("拉取邮件失败:", err)
	}
}
func main() {
	username, password, server := loadEnvConfig()

	c := connect(server)
	defer c.Logout()

	if err := c.Login(username, password); err != nil {
		log.Fatal("登录失败:", err)
	}

	fetchRecentEmails(c, "银行询价")

}
