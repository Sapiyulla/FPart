package auth

import (
	"context"
	"fpart/internal/domain/user"
	"net/mail"
)

type SigninCommand struct {
	Email, Password string
}

type RegCommand struct {
	Username string
	SigninCommand
}

type AuthService struct {
	userRepo UserRepository
	// need: PasswordHasher
}

func NewAuthService(userRepo UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

func (s *AuthService) Registration(ctx context.Context, cmd RegCommand) (Token string, err error) {
	user, err := user.NewUser(cmd.Username, cmd.Email, cmd.Password)
	if err != nil {
		return "", err
	}

	if err := s.userRepo.AddUser(ctx, user.Name, user.Email, user.Password); err != nil {
		return "", err
	}

	// jwt-token generate and return

	return "token", nil
}

func (s *AuthService) SignIn(ctx context.Context, cmd SigninCommand) (Token string, err error) {
	if _, err := mail.ParseAddress(cmd.Email); err != nil {
		return "", user.IncorrectEmailAddrErr
	}

	User, err := s.userRepo.FindUserByEmail(ctx, cmd.Email)
	if err != nil {
		return "", err
	}

	if User.Password != cmd.Password {
		return "", user.IncorrectPasswordErr
	}

	return User.Name, nil
}
