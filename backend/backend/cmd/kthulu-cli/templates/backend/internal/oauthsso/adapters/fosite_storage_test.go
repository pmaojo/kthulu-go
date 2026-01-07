package adapters

import (
	"context"
	"errors"
	"testing"

	"github.com/ory/fosite"
)

func TestFositeStorageAuthorizeCodeSessions(t *testing.T) {
	ctx := context.Background()
	request := fosite.NewRequest()
	providedSession := &fosite.DefaultSession{}

	var createdSignature string
	var createdRequester fosite.Requester
	var fetchedSignature string
	var fetchedSession fosite.Session
	var deletedSignature string

	sessions := &fakeSessionRepo{
		create: func(_ context.Context, signature string, requester fosite.Requester) error {
			createdSignature = signature
			createdRequester = requester
			return nil
		},
		get: func(_ context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
			fetchedSignature = signature
			fetchedSession = session
			return fosite.NewRequest(), nil
		},
		delete: func(_ context.Context, signature string) error {
			deletedSignature = signature
			return nil
		},
	}

	storage := &FositeStorage{sessions: sessions, tokens: &fakeTokenRepo{}}

	if err := storage.CreateAuthorizeCodeSession(ctx, "code", request); err != nil {
		t.Fatalf("CreateAuthorizeCodeSession returned error: %v", err)
	}
	if createdSignature != "code" {
		t.Fatalf("expected signature 'code', got %s", createdSignature)
	}
	if createdRequester != request {
		t.Fatalf("expected requester to be forwarded")
	}

	got, err := storage.GetAuthorizeCodeSession(ctx, "code", providedSession)
	if err != nil {
		t.Fatalf("GetAuthorizeCodeSession returned error: %v", err)
	}
	if fetchedSignature != "code" {
		t.Fatalf("expected fetch signature 'code', got %s", fetchedSignature)
	}
	if fetchedSession != providedSession {
		t.Fatalf("session argument not forwarded")
	}
	if got.GetSession() != providedSession {
		t.Fatalf("expected session to be attached to requester")
	}

	if err := storage.InvalidateAuthorizeCodeSession(ctx, "code"); err != nil {
		t.Fatalf("InvalidateAuthorizeCodeSession returned error: %v", err)
	}
	if deletedSignature != "code" {
		t.Fatalf("expected delete signature 'code', got %s", deletedSignature)
	}
}

func TestFositeStorageAccessTokenSessions(t *testing.T) {
	ctx := context.Background()
	providedSession := &fosite.DefaultSession{}

	var createdSignature, deletedSignature string
	var createdRequester fosite.Requester
	tokens := &fakeTokenRepo{
		create: func(_ context.Context, signature string, requester fosite.Requester) error {
			createdSignature = signature
			createdRequester = requester
			return nil
		},
		get: func(_ context.Context, signature string) (fosite.Requester, error) {
			req := fosite.NewRequest()
			req.SetSession(nil)
			if signature != "access" {
				t.Fatalf("unexpected signature %s", signature)
			}
			return req, nil
		},
		delete: func(_ context.Context, signature string) error {
			deletedSignature = signature
			return nil
		},
	}

	storage := &FositeStorage{sessions: &fakeSessionRepo{}, tokens: tokens}

	req := fosite.NewRequest()
	if err := storage.CreateAccessTokenSession(ctx, "access", req); err != nil {
		t.Fatalf("CreateAccessTokenSession returned error: %v", err)
	}
	if createdSignature != "access" || createdRequester != req {
		t.Fatalf("access token create not forwarded correctly")
	}

	got, err := storage.GetAccessTokenSession(ctx, "access", providedSession)
	if err != nil {
		t.Fatalf("GetAccessTokenSession returned error: %v", err)
	}
	if got.GetSession() != providedSession {
		t.Fatalf("expected session to be attached to access token requester")
	}

	if err := storage.DeleteAccessTokenSession(ctx, "access"); err != nil {
		t.Fatalf("DeleteAccessTokenSession returned error: %v", err)
	}
	if deletedSignature != "access" {
		t.Fatalf("expected delete signature 'access', got %s", deletedSignature)
	}
}

func TestFositeStorageRefreshTokenSessions(t *testing.T) {
	ctx := context.Background()
	providedSession := &fosite.DefaultSession{}

	var signatures []string
	tokens := &fakeTokenRepo{
		create: func(_ context.Context, signature string, requester fosite.Requester) error {
			signatures = append(signatures, "create:"+signature)
			return nil
		},
		get: func(_ context.Context, signature string) (fosite.Requester, error) {
			req := fosite.NewRequest()
			signatures = append(signatures, "get:"+signature)
			return req, nil
		},
		delete: func(_ context.Context, signature string) error {
			signatures = append(signatures, "delete:"+signature)
			return nil
		},
	}

	storage := &FositeStorage{sessions: &fakeSessionRepo{}, tokens: tokens}

	if err := storage.CreateRefreshTokenSession(ctx, "refresh", "access", fosite.NewRequest()); err != nil {
		t.Fatalf("CreateRefreshTokenSession returned error: %v", err)
	}
	if signatures[0] != "create:refresh" {
		t.Fatalf("unexpected create call: %v", signatures)
	}

	got, err := storage.GetRefreshTokenSession(ctx, "refresh", providedSession)
	if err != nil {
		t.Fatalf("GetRefreshTokenSession returned error: %v", err)
	}
	if got.GetSession() != providedSession {
		t.Fatalf("expected session to be attached to refresh requester")
	}

	if err := storage.DeleteRefreshTokenSession(ctx, "refresh"); err != nil {
		t.Fatalf("DeleteRefreshTokenSession returned error: %v", err)
	}

	signatures = signatures[:0]
	if err := storage.RotateRefreshToken(ctx, "request-id", "refresh-sig"); err != nil {
		t.Fatalf("RotateRefreshToken returned error: %v", err)
	}
	if len(signatures) != 2 || signatures[0] != "delete:request-id" || signatures[1] != "delete:request-id" {
		t.Fatalf("expected two delete calls for request-id, got %v", signatures)
	}
}

func TestFositeStoragePKCESessions(t *testing.T) {
	ctx := context.Background()
	providedSession := &fosite.DefaultSession{}

	var createSig, deleteSig string
	var getSig string
	var getSession fosite.Session

	sessions := &fakeSessionRepo{
		create: func(_ context.Context, signature string, requester fosite.Requester) error {
			createSig = signature
			return nil
		},
		get: func(_ context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
			getSig = signature
			getSession = session
			return fosite.NewRequest(), nil
		},
		delete: func(_ context.Context, signature string) error {
			deleteSig = signature
			return nil
		},
	}

	storage := &FositeStorage{sessions: sessions, tokens: &fakeTokenRepo{}}

	if err := storage.CreatePKCERequestSession(ctx, "pkce", fosite.NewRequest()); err != nil {
		t.Fatalf("CreatePKCERequestSession returned error: %v", err)
	}
	if createSig != "pkce" {
		t.Fatalf("expected create signature 'pkce', got %s", createSig)
	}

	got, err := storage.GetPKCERequestSession(ctx, "pkce", providedSession)
	if err != nil {
		t.Fatalf("GetPKCERequestSession returned error: %v", err)
	}
	if getSig != "pkce" || getSession != providedSession {
		t.Fatalf("session parameters not forwarded")
	}
	if got.GetSession() != providedSession {
		t.Fatalf("expected session attachment for PKCE request")
	}

	if err := storage.DeletePKCERequestSession(ctx, "pkce"); err != nil {
		t.Fatalf("DeletePKCERequestSession returned error: %v", err)
	}
	if deleteSig != "pkce" {
		t.Fatalf("expected delete signature 'pkce', got %s", deleteSig)
	}
}

func TestFositeStorageRevocationDelegation(t *testing.T) {
	ctx := context.Background()
	var revocations []string

	tokens := &fakeTokenRepo{
		delete: func(_ context.Context, signature string) error {
			revocations = append(revocations, signature)
			return nil
		},
	}

	storage := &FositeStorage{sessions: &fakeSessionRepo{}, tokens: tokens}

	if err := storage.RevokeRefreshToken(ctx, "req-id"); err != nil {
		t.Fatalf("RevokeRefreshToken returned error: %v", err)
	}
	if err := storage.RevokeAccessToken(ctx, "req-id"); err != nil {
		t.Fatalf("RevokeAccessToken returned error: %v", err)
	}

	if len(revocations) != 2 {
		t.Fatalf("expected two revocation calls, got %v", revocations)
	}
	if revocations[0] != "req-id" || revocations[1] != "req-id" {
		t.Fatalf("unexpected revocation arguments: %v", revocations)
	}
}

func TestFositeStorageErrorPropagation(t *testing.T) {
	wantErr := errors.New("boom")
	ctx := context.Background()
	req := fosite.NewRequest()
	sess := &fosite.DefaultSession{}

	tests := []struct {
		name    string
		storage func() *FositeStorage
		call    func(*FositeStorage) error
	}{
		{
			name: "CreateAuthorizeCodeSession",
			storage: func() *FositeStorage {
				return &FositeStorage{
					sessions: &fakeSessionRepo{create: func(context.Context, string, fosite.Requester) error { return wantErr }},
					tokens:   &fakeTokenRepo{},
				}
			},
			call: func(s *FositeStorage) error { return s.CreateAuthorizeCodeSession(ctx, "sig", req) },
		},
		{
			name: "GetAuthorizeCodeSession",
			storage: func() *FositeStorage {
				return &FositeStorage{
					sessions: &fakeSessionRepo{get: func(context.Context, string, fosite.Session) (fosite.Requester, error) { return nil, wantErr }},
					tokens:   &fakeTokenRepo{},
				}
			},
			call: func(s *FositeStorage) error { _, err := s.GetAuthorizeCodeSession(ctx, "sig", sess); return err },
		},
		{
			name: "InvalidateAuthorizeCodeSession",
			storage: func() *FositeStorage {
				return &FositeStorage{
					sessions: &fakeSessionRepo{delete: func(context.Context, string) error { return wantErr }},
					tokens:   &fakeTokenRepo{},
				}
			},
			call: func(s *FositeStorage) error { return s.InvalidateAuthorizeCodeSession(ctx, "sig") },
		},
		{
			name: "CreateAccessTokenSession",
			storage: func() *FositeStorage {
				return &FositeStorage{
					sessions: &fakeSessionRepo{},
					tokens:   &fakeTokenRepo{create: func(context.Context, string, fosite.Requester) error { return wantErr }},
				}
			},
			call: func(s *FositeStorage) error { return s.CreateAccessTokenSession(ctx, "sig", req) },
		},
		{
			name: "GetAccessTokenSession",
			storage: func() *FositeStorage {
				return &FositeStorage{
					sessions: &fakeSessionRepo{},
					tokens:   &fakeTokenRepo{get: func(context.Context, string) (fosite.Requester, error) { return nil, wantErr }},
				}
			},
			call: func(s *FositeStorage) error { _, err := s.GetAccessTokenSession(ctx, "sig", sess); return err },
		},
		{
			name: "DeleteAccessTokenSession",
			storage: func() *FositeStorage {
				return &FositeStorage{sessions: &fakeSessionRepo{}, tokens: &fakeTokenRepo{delete: func(context.Context, string) error { return wantErr }}}
			},
			call: func(s *FositeStorage) error { return s.DeleteAccessTokenSession(ctx, "sig") },
		},
		{
			name: "CreateRefreshTokenSession",
			storage: func() *FositeStorage {
				return &FositeStorage{sessions: &fakeSessionRepo{}, tokens: &fakeTokenRepo{create: func(context.Context, string, fosite.Requester) error { return wantErr }}}
			},
			call: func(s *FositeStorage) error { return s.CreateRefreshTokenSession(ctx, "sig", "access", req) },
		},
		{
			name: "GetRefreshTokenSession",
			storage: func() *FositeStorage {
				return &FositeStorage{sessions: &fakeSessionRepo{}, tokens: &fakeTokenRepo{get: func(context.Context, string) (fosite.Requester, error) { return nil, wantErr }}}
			},
			call: func(s *FositeStorage) error { _, err := s.GetRefreshTokenSession(ctx, "sig", sess); return err },
		},
		{
			name: "DeleteRefreshTokenSession",
			storage: func() *FositeStorage {
				return &FositeStorage{sessions: &fakeSessionRepo{}, tokens: &fakeTokenRepo{delete: func(context.Context, string) error { return wantErr }}}
			},
			call: func(s *FositeStorage) error { return s.DeleteRefreshTokenSession(ctx, "sig") },
		},
		{
			name: "RotateRefreshToken",
			storage: func() *FositeStorage {
				return &FositeStorage{sessions: &fakeSessionRepo{}, tokens: &fakeTokenRepo{delete: func(context.Context, string) error { return wantErr }}}
			},
			call: func(s *FositeStorage) error { return s.RotateRefreshToken(ctx, "req", "sig") },
		},
		{
			name: "CreatePKCERequestSession",
			storage: func() *FositeStorage {
				return &FositeStorage{sessions: &fakeSessionRepo{create: func(context.Context, string, fosite.Requester) error { return wantErr }}, tokens: &fakeTokenRepo{}}
			},
			call: func(s *FositeStorage) error { return s.CreatePKCERequestSession(ctx, "sig", req) },
		},
		{
			name: "GetPKCERequestSession",
			storage: func() *FositeStorage {
				return &FositeStorage{sessions: &fakeSessionRepo{get: func(context.Context, string, fosite.Session) (fosite.Requester, error) { return nil, wantErr }}, tokens: &fakeTokenRepo{}}
			},
			call: func(s *FositeStorage) error { _, err := s.GetPKCERequestSession(ctx, "sig", sess); return err },
		},
		{
			name: "DeletePKCERequestSession",
			storage: func() *FositeStorage {
				return &FositeStorage{sessions: &fakeSessionRepo{delete: func(context.Context, string) error { return wantErr }}, tokens: &fakeTokenRepo{}}
			},
			call: func(s *FositeStorage) error { return s.DeletePKCERequestSession(ctx, "sig") },
		},
		{
			name: "RevokeRefreshToken",
			storage: func() *FositeStorage {
				return &FositeStorage{sessions: &fakeSessionRepo{}, tokens: &fakeTokenRepo{delete: func(context.Context, string) error { return wantErr }}}
			},
			call: func(s *FositeStorage) error { return s.RevokeRefreshToken(ctx, "sig") },
		},
		{
			name: "RevokeAccessToken",
			storage: func() *FositeStorage {
				return &FositeStorage{sessions: &fakeSessionRepo{}, tokens: &fakeTokenRepo{delete: func(context.Context, string) error { return wantErr }}}
			},
			call: func(s *FositeStorage) error { return s.RevokeAccessToken(ctx, "sig") },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(tc.storage()); !errors.Is(err, wantErr) {
				t.Fatalf("expected %v, got %v", wantErr, err)
			}
		})
	}
}

type fakeSessionRepo struct {
	create func(context.Context, string, fosite.Requester) error
	get    func(context.Context, string, fosite.Session) (fosite.Requester, error)
	delete func(context.Context, string) error
}

func (f *fakeSessionRepo) CreateSession(ctx context.Context, signature string, requester fosite.Requester) error {
	if f.create != nil {
		return f.create(ctx, signature, requester)
	}
	return nil
}

func (f *fakeSessionRepo) GetSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	if f.get != nil {
		return f.get(ctx, signature, session)
	}
	return nil, nil
}

func (f *fakeSessionRepo) DeleteSession(ctx context.Context, signature string) error {
	if f.delete != nil {
		return f.delete(ctx, signature)
	}
	return nil
}

type fakeTokenRepo struct {
	create func(context.Context, string, fosite.Requester) error
	get    func(context.Context, string) (fosite.Requester, error)
	delete func(context.Context, string) error
}

func (f *fakeTokenRepo) CreateToken(ctx context.Context, signature string, requester fosite.Requester) error {
	if f.create != nil {
		return f.create(ctx, signature, requester)
	}
	return nil
}

func (f *fakeTokenRepo) GetToken(ctx context.Context, signature string) (fosite.Requester, error) {
	if f.get != nil {
		return f.get(ctx, signature)
	}
	return nil, nil
}

func (f *fakeTokenRepo) DeleteToken(ctx context.Context, signature string) error {
	if f.delete != nil {
		return f.delete(ctx, signature)
	}
	return nil
}
