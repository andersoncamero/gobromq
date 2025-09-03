package auth

import (
	"fmt"
	"sync"
)

type SimpleAuth struct {
	users map[string]*User
	mu    sync.RWMutex
}

func NewSimpleAuth() *SimpleAuth {
	auth := &SimpleAuth{
		users: make(map[string]*User),
	}

	auth.seedDefaultUsers()
	return auth
}

func (s *SimpleAuth) seedDefaultUsers() {
	admin := &User{
		Username: "admin",
		Password: "password123",
		Roles:    []string{"admin"},
		Permissions: map[string][]Action{
			"*": {ActionRead, ActionWrite},
		},
		Active: true,
	}

	ezloUser := &User{
		Username: "ezlo_device",
		Password: "ezlo_secure_2024",
		Roles:    []string{"device"},
		Permissions: map[string][]Action{
			"ezlo/*":    {ActionRead, ActionWrite},
			"sensor/*":  {ActionWrite},
			"status/*":  {ActionWrite},
			"command/*": {ActionRead},
		},
		Active: true,
	}

	monitor := &User{
		Username: "monitor",
		Password: "monitor_pass",
		Roles:    []string{"observer"},
		Permissions: map[string][]Action{
			"*": {ActionRead}, // Solo lectura
		},
		Active: true,
	}

	testUser := &User{
		Username: "testuser",
		Password: "testpass",
		Roles:    []string{"user"},
		Permissions: map[string][]Action{
			"test/*":     {ActionRead, ActionWrite},
			"personal/*": {ActionRead, ActionWrite},
		},
		Active: true,
	}

	s.users["admin"] = admin
	s.users["ezlo_device"] = ezloUser
	s.users["monitor"] = monitor
	s.users["testuser"] = testUser
}

func (s *SimpleAuth) Authenticate(username, password string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, ErrUserNotFound
	}

	if !user.Active {
		return nil, ErrUserInactive
	}

	if user.Password != password {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (s *SimpleAuth) Authorize(user *User, action Action, topic string) bool {
	if user == nil {
		return false
	}

	return user.CanPerformAction(action, topic)
}

func (s *SimpleAuth) GetUser(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (s *SimpleAuth) AddUser(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if user.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if _, exists := s.users[user.Username]; exists {
		return fmt.Errorf("user %s already exists", user.Username)
	}

	s.users[user.Username] = user
	return nil
}

func (s *SimpleAuth) UpdateUser(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.Username]; !exists {
		return fmt.Errorf("user %s does not exist", user.Username)
	}

	s.users[user.Username] = user
	return nil
}

func (s *SimpleAuth) ListUsers() map[string]*User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*User)

	for username, user := range s.users {
		userCopy := &User{
			Username:    user.Username,
			Password:    "*****",
			Roles:       user.Roles,
			Permissions: user.Permissions,
			Active:      user.Active,
		}

		result[username] = userCopy
	}
	return result
}

func (s *SimpleAuth) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.Unlock()

	activeUsers := 0
	roleCount := make(map[string]int)

	for _, user := range s.users {
		if user.Active {
			activeUsers++
		}

		for _, role := range user.Roles {
			roleCount[role]++
		}
	}

	return map[string]interface{}{
		"active_users": activeUsers,
		"total_users":  len(s.users),
		"roles":        roleCount,
	}
}
