package auth

import "strings"

var (
	ErrInvalidCredentials = &AuthError{"INVALID_CREDENTIALS", "Invalid username or password"}
	ErrUserNotFound       = &AuthError{"USER_NOT_FOUND", "User not found"}
	ErrUserInactive       = &AuthError{"USER_INACTIVE", "User is inactive"}
	ErrUnauthorized       = &AuthError{"UNAUTHORIZED", "Unauthorized access"}
)

type Autenticator interface {
	Authenticate(username, password string) (*User, error)
	Authorize(user *User, action Action, topic string) bool
	GetUser(username string) (*User, error)
}

type Action int

const (
	ActionRead Action = iota
	ActionWrite
)

func (a Action) String() string {
	switch a {
	case ActionRead:
		return "READ"
	case ActionWrite:
		return "WRITE"
	default:
		return "UNKNOWN"
	}
}

type User struct {
	Username    string
	Password    string
	Roles       []string
	Permissions map[string][]Action
	Active      bool
}

func (u *User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func (u *User) CanPerformAction(action Action, topic string) bool {
	if !u.Active {
		return false
	}

	if u.HasRole("admin") {
		return true
	}

	for pattern, actions := range u.Permissions {
		if u.matchTopic(pattern, topic) {
			for _, allowedAction := range actions {
				if allowedAction == action {
					return true
				}
			}
		}
	}
	return false
}

func (u *User) matchTopic(pattern, topic string) bool {
	if pattern == "*" {
		return true
	}
	if pattern == topic {
		return true
	}
	return u.topicMatches(pattern, topic)
}

func (u *User) topicMatches(pattern, topic string) bool {
	return u.matchTopicLevels(strings.Split(pattern, "/"), strings.Split(topic, "/"))
}

func (u *User) matchTopicLevels(patternLevels, topicLevels []string) bool {
	patternIndex, topicIndex := 0, 0

	for patternIndex < len(patternLevels) && topicIndex < len(topicLevels) {
		patternLevel := patternLevels[patternIndex]
		topicLevel := topicLevels[topicIndex]

		if patternLevel == "*" {
			return true
		}

		if patternLevel == "+" {
			patternIndex++
			topicIndex++
			continue
		}

		if patternLevel == "#" {
			return true
		}

		if patternLevel != topicLevel {
			return false
		}

		patternIndex++
		topicIndex++
	}

	if patternIndex < len(patternLevels) {
		remaining := patternLevels[patternIndex]
		return remaining == "#" || remaining == "*"
	}

	return patternIndex == len(patternLevels) && topicIndex == len(topicLevels)
}

type AuthError struct {
	Code    string
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}
