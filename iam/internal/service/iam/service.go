package iam

import "time"

type Service struct {
	userRepository    UserRepository
	sessionRepository SessionRepository
	sessionTTL        time.Duration
	bcryptCost        int
}

func NewService(userRepository UserRepository, sessionRepository SessionRepository, ttl time.Duration, bcryptCost int) *Service {
	return &Service{
		userRepository:    userRepository,
		sessionRepository: sessionRepository,
		sessionTTL:        ttl,
		bcryptCost:        bcryptCost,
	}
}
