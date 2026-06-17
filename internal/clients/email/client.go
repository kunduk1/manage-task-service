package email

//go:generate go tool mockgen -source=client.go -destination=mocks/client_mock.go -package=mocks

import "context"

// Invite — данные для письма-приглашения в команду.
type Invite struct {
	ToEmail   string // e-mail приглашаемого
	ToName    string // имя приглашаемого
	TeamID    int64
	InviterID int64  // кто пригласил
	Role      string // роль, которую получит приглашаемый
}

// Client — абстракция почтового сервиса. Close() есть ради единообразной
// регистрации в closer (как у db/cache клиентов).
type Client interface {
	SendInvite(ctx context.Context, in Invite) error
	Close() error
}
