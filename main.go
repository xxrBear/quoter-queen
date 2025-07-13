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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // 纯 Go，无需 CGO
)

// MailState 定义数据库表结构
type MailState struct {
	ID   uint `gorm:"primaryKey"` // 主键
	Name string
	Age  int
}

// 连接数据库，自动创建test.db文件，返回*gorm.DB
func connectDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 自动迁移，创建表
	if err := db.AutoMigrate(&MailState{}); err != nil {
		log.Fatalf("自动迁移失败: %v", err)
	}

	// 插入测试数据
	db.Create(&MailState{Name: "Alice", Age: 30})
	db.Create(&MailState{Name: "Bob", Age: 25})

	// 查询所有用户
	var users []MailState
	if err := db.Find(&users).Error; err != nil {
		log.Fatalf("查询失败: %v", err)
	}

	for _, user := range users {
		fmt.Printf("ID: %d, Name: %s, Age: %d\n", user.ID, user.Name, user.Age)
	}

	return db
}

// 加载环境变量中的邮箱配置
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

// 连接IMAP服务器（TLS）
func connect(server string) *client.Client {
	c, err := client.DialTLS(server, nil)
	if err != nil {
		log.Fatal("连接失败:", err)
	}
	fmt.Println("已连接:", server)
	return c
}

// 抓取指定文件夹当天之后的邮件
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

	ids, err := c.Search(criteria)
	if err != nil {
		log.Fatal("搜索邮件失败:", err)
	}
	if len(ids) == 0 {
		fmt.Println("今天没有邮件")
		return
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(ids...)
	fmt.Println("搜索到 UID:", ids)

	items := []imap.FetchItem{imap.FetchEnvelope}

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.Fetch(seqSet, items, messages)
	}()

	// 遍历邮件
	for msg := range messages {
		if msg == nil || msg.Envelope == nil {
			continue
		}
		fmt.Println("--------")
		fmt.Println("标题:", msg.Envelope.Subject)
		if len(msg.Envelope.From) > 0 {
			fmt.Println("发件人:", msg.Envelope.From[0].Address())
		} else {
			fmt.Println("发件人: 无")
		}
		fmt.Println("时间:", msg.Envelope.Date)
	}

	if err := <-done; err != nil {
		log.Fatal("拉取邮件失败:", err)
	}
}

func main() {
	// 加载配置并连接邮箱示例
	// username, password, server := loadEnvConfig()

	// c := connect(server)
	// defer c.Logout()

	// if err := c.Login(username, password); err != nil {
	// 	log.Fatal("登录失败:", err)
	// }

	// fetchRecentEmails(c, "银行询价")

	// 连接数据库示例
	connectDB()
}
