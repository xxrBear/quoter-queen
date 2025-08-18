package db

import (
	"time"
)

// MailState 定义数据库表结构
type MailState struct {
	ID       uint `gorm:"primaryKey"` // 主键
	Subject  string
	Address  string
	SendTime time.Time
}

// WriteMails 接收一个数组并写入 SQLite
