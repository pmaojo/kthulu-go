package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pmaojo/kthulu-go/backend/core"
	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/domain/repository"
)

// mockOrganizationUserRepository implements OrganizationUserRepository for testing
type mockOrganizationUserRepository struct {
	role domain.OrganizationRole
}

func (m *mockOrganizationUserRepository) Create(ctx context.Context, orgUser *domain.OrganizationUser) error {
	return nil
}
func (m *mockOrganizationUserRepository) FindByID(ctx context.Context, id uint) (*domain.OrganizationUser, error) {
	return nil, nil
}
func (m *mockOrganizationUserRepository) FindByOrganizationAndUser(ctx context.Context, organizationID, userID uint) (*domain.OrganizationUser, error) {
	return nil, nil
}
func (m *mockOrganizationUserRepository) Update(ctx context.Context, orgUser *domain.OrganizationUser) error {
	return nil
}
func (m *mockOrganizationUserRepository) Delete(ctx context.Context, id uint) error { return nil }
func (m *mockOrganizationUserRepository) FindByOrganization(ctx context.Context, organizationID uint) ([]*domain.OrganizationUser, error) {
	return nil, nil
}
func (m *mockOrganizationUserRepository) FindByUser(ctx context.Context, userID uint) ([]*domain.OrganizationUser, error) {
	return nil, nil
}
func (m *mockOrganizationUserRepository) FindByRole(ctx context.Context, organizationID uint, role domain.OrganizationRole) ([]*domain.OrganizationUser, error) {
	return nil, nil
}
func (m *mockOrganizationUserRepository) CountByOrganization(ctx context.Context, organizationID uint) (int64, error) {
	return 0, nil
}
func (m *mockOrganizationUserRepository) IsUserInOrganization(ctx context.Context, organizationID, userID uint) (bool, error) {
	return false, nil
}
func (m *mockOrganizationUserRepository) GetUserRole(ctx context.Context, organizationID, userID uint) (domain.OrganizationRole, error) {
	return m.role, nil
}
func (m *mockOrganizationUserRepository) HasRole(ctx context.Context, organizationID, userID uint, role domain.OrganizationRole) (bool, error) {
	return false, nil
}
func (m *mockOrganizationUserRepository) RemoveUserFromOrganization(ctx context.Context, organizationID, userID uint) error {
	return nil
}
func (m *mockOrganizationUserRepository) UpdateUserRole(ctx context.Context, organizationID, userID uint, role domain.OrganizationRole) error {
	return nil
}

// mockInvitationRepository implements InvitationRepository for testing
type mockInvitationRepository struct {
	invitations []*domain.Invitation
}

func (m *mockInvitationRepository) Create(ctx context.Context, invitation *domain.Invitation) error {
	invitation.ID = uint(len(m.invitations) + 1)
	m.invitations = append(m.invitations, invitation)
	return nil
}
func (m *mockInvitationRepository) FindByID(ctx context.Context, id uint) (*domain.Invitation, error) {
	return nil, nil
}
func (m *mockInvitationRepository) FindByToken(ctx context.Context, token string) (*domain.Invitation, error) {
	return nil, nil
}
func (m *mockInvitationRepository) Update(ctx context.Context, invitation *domain.Invitation) error {
	return nil
}
func (m *mockInvitationRepository) Delete(ctx context.Context, id uint) error { return nil }
func (m *mockInvitationRepository) FindByOrganization(ctx context.Context, organizationID uint) ([]*domain.Invitation, error) {
	return nil, nil
}
func (m *mockInvitationRepository) FindByEmail(ctx context.Context, email string) ([]*domain.Invitation, error) {
	return nil, nil
}
func (m *mockInvitationRepository) FindByInviter(ctx context.Context, inviterID uint) ([]*domain.Invitation, error) {
	return nil, nil
}
func (m *mockInvitationRepository) FindByStatus(ctx context.Context, status domain.InvitationStatus) ([]*domain.Invitation, error) {
	return nil, nil
}
func (m *mockInvitationRepository) FindExpired(ctx context.Context) ([]*domain.Invitation, error) {
	return nil, nil
}
func (m *mockInvitationRepository) ExistsByToken(ctx context.Context, token string) (bool, error) {
	return false, nil
}
func (m *mockInvitationRepository) ExistsPendingByEmail(ctx context.Context, organizationID uint, email string) (bool, error) {
	return false, nil
}
func (m *mockInvitationRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	return nil
}
func (m *mockInvitationRepository) MarkExpired(ctx context.Context, before time.Time) error {
	return nil
}

// mockInvitationNotifier implements NotificationProvider and records the last request
type mockInvitationNotifier struct {
	lastReq repository.NotificationRequest
	err     error
}

func (m *mockInvitationNotifier) SendNotification(ctx context.Context, req repository.NotificationRequest) error {
	m.lastReq = req
	return m.err
}
func (m *mockInvitationNotifier) SendEmailConfirmation(ctx context.Context, email, confirmationCode string) error {
	return nil
}
func (m *mockInvitationNotifier) SendPasswordReset(ctx context.Context, email, resetCode string) error {
	return nil
}
func (m *mockInvitationNotifier) SendWelcomeEmail(ctx context.Context, email, name string) error {
	return nil
}

// recordingLogger records error messages
type recordingLogger struct {
	errors []string
}

func (l *recordingLogger) Debug(msg string, fields ...interface{}) {}
func (l *recordingLogger) Info(msg string, fields ...interface{})  {}
func (l *recordingLogger) Warn(msg string, fields ...interface{})  {}
func (l *recordingLogger) Error(msg string, fields ...interface{}) { l.errors = append(l.errors, msg) }
func (l *recordingLogger) Fatal(msg string, fields ...interface{}) {}
func (l *recordingLogger) With(fields ...interface{}) core.Logger  { return l }
func (l *recordingLogger) Sync() error                             { return nil }

func TestInviteUser_SendsNotification(t *testing.T) {
	ctx := context.Background()
	orgUserRepo := &mockOrganizationUserRepository{role: domain.OrganizationRoleOwner}
	invRepo := &mockInvitationRepository{}
	userRepo := &mockUserRepository{users: make(map[string]*domain.User)}
	notifier := &mockInvitationNotifier{}
	logger := &recordingLogger{}

	uc := NewOrganizationUseCase(nil, orgUserRepo, invRepo, userRepo, notifier, logger)

	req := InviteUserRequest{Email: "invitee@example.com", Role: domain.OrganizationRoleMember}
	invitation, err := uc.InviteUser(ctx, 1, 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if notifier.lastReq.To != req.Email {
		t.Errorf("expected notification to %s, got %s", req.Email, notifier.lastReq.To)
	}
	if token, ok := notifier.lastReq.Data["token"]; !ok || token != invitation.Token {
		t.Errorf("notification token mismatch")
	}
}

func TestInviteUser_NotificationFailureLogged(t *testing.T) {
	ctx := context.Background()
	orgUserRepo := &mockOrganizationUserRepository{role: domain.OrganizationRoleOwner}
	invRepo := &mockInvitationRepository{}
	userRepo := &mockUserRepository{users: make(map[string]*domain.User)}
	notifier := &mockInvitationNotifier{err: errors.New("send failed")}
	logger := &recordingLogger{}

	uc := NewOrganizationUseCase(nil, orgUserRepo, invRepo, userRepo, notifier, logger)

	req := InviteUserRequest{Email: "invitee@example.com", Role: domain.OrganizationRoleMember}
	invitation, err := uc.InviteUser(ctx, 1, 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if invitation == nil {
		t.Fatalf("expected invitation result")
	}
	if len(logger.errors) == 0 {
		t.Errorf("expected error to be logged")
	}
}
