//go:build integration

package integration_test

import (
	"net/http"

	authv1 "github.com/kunduk1/manage-task-service/pkg/auth/v1"
)

func (s *IntegrationSuite) TestRegister_Duplicate() {
	s.register("dup@example.com", "Dup", "secret123")

	rsp := s.do(http.MethodPost, "/api/v1/register", "",
		authv1.RegisterRequest{Email: "dup@example.com", Name: "Dup2", Password: "secret123"})
	s.requireError(rsp, http.StatusConflict, "user already exists")
}

func (s *IntegrationSuite) TestRegister_ValidationError() {
	// слишком короткий пароль (<6) → 400 ErrValidation
	rsp := s.do(http.MethodPost, "/api/v1/register", "",
		authv1.RegisterRequest{Email: "v@example.com", Name: "V", Password: "123"})
	s.requireStatus(rsp, http.StatusBadRequest)
}

func (s *IntegrationSuite) TestRegister_MalformedJSON() {
	rsp := s.do(http.MethodPost, "/api/v1/register", "", []byte("{not valid json"))
	s.requireStatus(rsp, http.StatusBadRequest)
}

func (s *IntegrationSuite) TestLogin_WrongPassword() {
	s.register("u@example.com", "U", "secret123")

	rsp := s.do(http.MethodPost, "/api/v1/login", "",
		authv1.LoginRequest{Email: "u@example.com", Password: "wrong-pass"})
	s.requireError(rsp, http.StatusUnauthorized, "invalid email or password")
}

func (s *IntegrationSuite) TestLogin_RefreshTokenPersistedInRedis() {
	reg := s.register("r@example.com", "R", "secret123")
	tok := s.login("r@example.com", "secret123")

	s.Require().NotEmpty(tok.AccessToken)
	s.Require().NotEmpty(tok.RefreshToken)
	s.Equal("Bearer", tok.TokenType)
	s.Positive(tok.ExpiresIn)

	// refresh-токен должен реально лежать в Redis и указывать на пользователя
	uid, err := s.tokenRepo.GetRefresh(s.ctx, tok.RefreshToken)
	s.Require().NoError(err, "refresh token not found in redis")
	s.Equal(reg.ID, uid)
}
