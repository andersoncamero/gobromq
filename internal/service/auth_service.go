package service

import (
	"fmt"
	"sync"
	"time"

	"github/go/gobromq/internal/auth"
	"github/go/gobromq/internal/entities"
)

type AuthService struct {
	authenticator auth.Autenticator
	sessions      map[string]*entities.AuthenticateSession
	mu            sync.RWMutex
}

func NewAuthService(authenticator auth.Autenticator) *AuthService {
	return &AuthService{
		authenticator: authenticator,
		sessions:      make(map[string]*entities.AuthenticateSession),
	}
}

func (as *AuthService) AuthenticateClient(clientID, username, password string) (*entities.AuthenticateSession, error) {
	user, err := as.authenticator.Authenticate(username, password)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	session := &entities.AuthenticateSession{
		ClientID:    clientID,
		Username:    user.Username,
		Roles:       user.Roles,
		Permissions: as.convertPermissions(user.Permissions),
		AuthTime:    time.Now(),
		Active:      true,
	}

	as.mu.Lock()
	as.sessions[clientID] = session
	as.mu.Unlock()

	return session, nil
}

func (as *AuthService) CanPublish(clientID string, topic string) bool {
	as.mu.RLock()
	session, exists := as.sessions[clientID]
	as.mu.RUnlock()

	if !exists || !session.Active {
		return false
	}

	user, err := as.authenticator.GetUser(session.Username)
	if err != nil {
		return false
	}

	return as.authenticator.Authorize(user, auth.ActionWrite, topic)
}

func (as *AuthService) GetSession(clientID string) (*entities.AuthenticateSession, bool) {
	as.mu.RLock()
	defer as.mu.RUnlock()

	session, exists := as.sessions[clientID]
	return session, exists
}

func (as *AuthService) RemoveSession(clientID string) {
	as.mu.Lock()
	defer as.mu.Unlock()

	delete(as.sessions, clientID)
}

func (as *AuthService) IsAuthenticated(clientID string) bool {
	as.mu.RLock()
	defer as.mu.RUnlock()

	session, exists := as.sessions[clientID]
	return exists && session.Active
}

func (as *AuthService) convertPermissions(userPerms map[string][]auth.Action) map[string][]string {
	result := make(map[string][]string)

	for topi, actions := range userPerms {
		var actionStrs []string
		for _, action := range actions {
			actionStrs = append(actionStrs, action.String())
		}
		result[topi] = actionStrs
	}
	return result
}

func (as *AuthService) GetStats() map[string]interface{} {
	as.mu.RLock()
	defer as.mu.RUnlock()

	activeSessions := 0
	roleCount := make(map[string]int)

	for _, session := range as.sessions {
		if session.Active {
			activeSessions++
		}
		for _, role := range session.Roles {
			roleCount[role]++
		}
	}

	return map[string]interface{}{
		"active_sessions": activeSessions,
		"total_sessions":  len(as.sessions),
		"roles":           roleCount,
	}
}

func (as *AuthService) CanSubscribe(clientID string, topic string) bool {
	as.mu.RLock()
	session, exists := as.sessions[clientID]
	as.mu.RUnlock()

	if !exists || !session.Active {
		return false
	}

	user, err := as.authenticator.GetUser(session.Username)
	if err != nil {
		return false
	}

	return as.authenticator.Authorize(user, auth.ActionRead, topic)
}
