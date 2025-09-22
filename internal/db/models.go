package db

import (
    "time"

    "gorm.io/datatypes"
)

// 角色枚举（与 OpenAPI 模型一致）
type Role string

const (
    RoleStudent    Role = "STUDENT"
    RoleAdmin      Role = "ADMIN"
    RoleSuperAdmin Role = "SUPER_ADMIN"
)

// 工单状态枚举（与 OpenAPI 模型一致）
type TicketStatus string

const (
    TicketStatusNew           TicketStatus = "NEW"
    TicketStatusClaimed       TicketStatus = "CLAIMED"
    TicketStatusInProgress    TicketStatus = "IN_PROGRESS"
    TicketStatusResolved      TicketStatus = "RESOLVED"
    TicketStatusClosed        TicketStatus = "CLOSED"
    TicketStatusSpamPending   TicketStatus = "SPAM_PENDING"
    TicketStatusSpamConfirmed TicketStatus = "SPAM_CONFIRMED"
    TicketStatusSpamRejected  TicketStatus = "SPAM_REJECTED"
)

// User 表：用户基础信息
type User struct {
    ID           uint      `gorm:"primaryKey"`
    Email        string    `gorm:"type:varchar(255);uniqueIndex;not null;comment:邮箱"`
    Name         string    `gorm:"type:varchar(255);not null;comment:姓名"`
    Role         Role      `gorm:"type:varchar(20);index;not null;comment:角色"`
    Phone        *string   `gorm:"type:varchar(50)"`
    Dept         *string   `gorm:"type:varchar(100)"`
    IsActive     bool      `gorm:"not null;default:true"`
    AllowEmail   bool      `gorm:"not null;default:true;comment:允许邮件提醒"`
    PasswordHash string    `gorm:"type:char(60);not null;comment:密码哈希"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

func (User) TableName() string { return "users" }

// Ticket 表：工单
type Ticket struct {
    ID              uint         `gorm:"primaryKey"`
    UserID          uint         `gorm:"index;not null;comment:提单学生ID"`
    Title           string       `gorm:"type:varchar(255);not null"`
    Content         string       `gorm:"type:text;not null"`
    Category        string       `gorm:"type:varchar(100);index"`
    IsUrgent        bool         `gorm:"not null;default:false"`
    IsAnonymous     bool         `gorm:"not null;default:false"`
    Status          TicketStatus `gorm:"type:varchar(20);index;not null;default:'NEW'"`
    AssignedAdminID *uint        `gorm:"index;comment:受理管理员ID"`
    ClaimedAt       *time.Time
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

func (Ticket) TableName() string { return "tickets" }

// TicketMessage 表：工单消息（含内部备注）
type TicketMessage struct {
    ID            uint      `gorm:"primaryKey"`
    TicketID      uint      `gorm:"index;not null"`
    SenderUserID  uint      `gorm:"index;not null"`
    Body          string    `gorm:"type:text;not null"`
    IsInternalNote bool     `gorm:"not null;default:false;comment:是否内部备注"`
    CreatedAt     time.Time
}

func (TicketMessage) TableName() string { return "ticket_messages" }

// Image 表：图片元数据
type Image struct {
    ID        uint      `gorm:"primaryKey"`
    Sha256    string    `gorm:"type:char(64);uniqueIndex;not null"`
    Mime      string    `gorm:"type:varchar(100);not null"`
    Size      int64     `gorm:"not null;default:0"`
    Width     int       `gorm:"not null;default:0"`
    Height    int       `gorm:"not null;default:0"`
    ObjectKey string    `gorm:"type:varchar(255);not null"`
    RefCount  int       `gorm:"not null;default:0"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (Image) TableName() string { return "images" }

// TicketImage 关联表：工单-图片 多对多
type TicketImage struct {
    TicketID  uint      `gorm:"primaryKey"`
    ImageID   uint      `gorm:"primaryKey"`
    CreatedAt time.Time
}

func (TicketImage) TableName() string { return "ticket_images" }

// Rating 表：工单评分（一个工单一条评分）
type Rating struct {
    ID        uint      `gorm:"primaryKey"`
    TicketID  uint      `gorm:"uniqueIndex;not null;comment:一个工单只允许一个评分"`
    UserID    uint      `gorm:"index;not null"`
    Stars     int       `gorm:"not null"`
    Comment   string    `gorm:"type:text"`
    CreatedAt time.Time
}

func (Rating) TableName() string { return "ratings" }

// AuditLog 表：审计日志（Diff 为 JSON）
type AuditLog struct {
    ID          uint           `gorm:"primaryKey"`
    ActorUserID uint           `gorm:"index;not null"`
    Action      string         `gorm:"type:varchar(100);index;not null"`
    Entity      string         `gorm:"type:varchar(100);index;not null"`
    EntityID    uint           `gorm:"index;not null"`
    Diff        datatypes.JSON `gorm:"type:jsonb"`
    CreatedAt   time.Time
}

func (AuditLog) TableName() string { return "audit_logs" }

// SpamFlag 表：垃圾举报与复核
type SpamFlag struct {
    ID                     uint       `gorm:"primaryKey"`
    TicketID               uint       `gorm:"uniqueIndex;not null"`
    FlaggedByAdminID       uint       `gorm:"index;not null"`
    Reason                 string     `gorm:"type:text;not null"`
    Status                 string     `gorm:"type:varchar(50);index;not null;default:'PENDING'"`
    ReviewedBySuperAdminID *uint      `gorm:"index"`
    ReviewedAt             *time.Time
    CreatedAt              time.Time
    UpdatedAt              time.Time
}

func (SpamFlag) TableName() string { return "spam_flags" }

// CannedReply 表：常用回复
type CannedReply struct {
    ID          uint      `gorm:"primaryKey"`
    AdminUserID uint      `gorm:"index;not null"`
    Title       string    `gorm:"type:varchar(255);not null"`
    Body        string    `gorm:"type:text;not null"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

func (CannedReply) TableName() string { return "canned_replies" }