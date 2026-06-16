package model

import "time"

// TeamRole — роль пользователя в команде (значения совпадают с ENUM в team_members.role).
type TeamRole string

const (
	RoleOwner  TeamRole = "owner"
	RoleAdmin  TeamRole = "admin"
	RoleMember TeamRole = "member"
)

// Team — доменная сущность команды.
type Team struct {
	ID          int64
	Name        string
	Description string
	CreatedBy   int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TeamMember — членство пользователя в команде с ролью
type TeamMember struct {
	TeamID   int64
	UserID   int64
	Role     TeamRole
	JoinedAt time.Time
}

// TeamStats — результат аналитического запроса: по каждой команде её название,
// число участников и число задач, переведённых в done за последние 7 дней.
type TeamStats struct {
	TeamID        int64
	Name          string
	MemberCount   int64
	DoneLast7Days int64
}

// TopCreator — строка топа пользователей по числу созданных задач в команде за период.
type TopCreator struct {
	TeamID       int64
	UserID       int64
	UserName     string
	CreatedCount int64
	Rank         int64
}
