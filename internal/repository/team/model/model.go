package model

import "time"

type Team struct {
	ID          int64     `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	CreatedBy   int64     `db:"created_by"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type TeamMember struct {
	TeamID   int64     `db:"team_id"`
	UserID   int64     `db:"user_id"`
	Role     string    `db:"role"`
	JoinedAt time.Time `db:"joined_at"`
}

// TeamStats — статистика по команде.
type TeamStats struct {
	TeamID        int64  `db:"team_id"`
	Name          string `db:"name"`
	MemberCount   int64  `db:"member_count"`
	DoneLast7Days int64  `db:"done_last_7_days"`
}

// TopCreator — топ авторов задач в команде.
type TopCreator struct {
	TeamID       int64  `db:"team_id"`
	UserID       int64  `db:"user_id"`
	UserName     string `db:"user_name"`
	CreatedCount int64  `db:"created_count"`
	Rank         int64  `db:"rnk"`
}
