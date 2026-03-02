package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	mock_auth "github.com/O-Nikitin/Social/cmd/api/mock/auth"
	mock_mailer "github.com/O-Nikitin/Social/cmd/api/mock/mailer"
	mock_storage "github.com/O-Nikitin/Social/cmd/api/mock/store"
	"github.com/O-Nikitin/Social/internal/ratelimiter"
	"github.com/O-Nikitin/Social/internal/store"
	"github.com/O-Nikitin/Social/internal/store/cache"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
)

type AppMocks struct {
	Posts     *mock_storage.MockPosts
	Users     *mock_storage.MockUsers
	Comments  *mock_storage.MockComments
	Followers *mock_storage.MockFollowers
	Roles     *mock_storage.MockRoles
	Cache     *mock_storage.MockUserCache
	Mailer    *mock_mailer.MockClient
	Auth      *mock_auth.MockAuthenticator
}

func newTestApp(t *testing.T, cfg config) (*application, *AppMocks) {
	t.Helper()

	//log := zap.NewNop().Sugar() disable logs
	log := zap.Must(zap.NewProduction()).Sugar()

	ctrl := gomock.NewController(t)
	mockPosts := mock_storage.NewMockPosts(ctrl)
	mockUsers := mock_storage.NewMockUsers(ctrl)
	mockComments := mock_storage.NewMockComments(ctrl)
	mockFollowers := mock_storage.NewMockFollowers(ctrl)
	mockRoles := mock_storage.NewMockRoles(ctrl)

	mockUserCache := mock_storage.NewMockUserCache(ctrl)

	mockMailer := mock_mailer.NewMockClient(ctrl)

	mockAuth := mock_auth.NewMockAuthenticator(ctrl)

	storage := store.Storage{
		Posts:     mockPosts,
		Users:     mockUsers,
		Comments:  mockComments,
		Followers: mockFollowers,
		Roles:     mockRoles,
	}

	cache := cache.Storage{
		Users: mockUserCache,
	}

	// Rate limiter
	rateLimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.rateLimiter.RequestsPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	a := &application{
		logger:        log,
		store:         storage,
		cacheStorage:  cache,
		mailer:        mockMailer,
		authenticator: mockAuth,
		config:        cfg,
		rateLimiter:   rateLimiter, //TODO real rate limiter should be replaced with mock
	}
	m := &AppMocks{
		Posts:     mockPosts,
		Users:     mockUsers,
		Comments:  mockComments,
		Followers: mockFollowers,
		Roles:     mockRoles,
		Cache:     mockUserCache,
		Mailer:    mockMailer,
		Auth:      mockAuth,
	}

	return a, m
}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	return rr
}
func checkResponseCode(got, want int, t *testing.T) {
	if got != want {
		t.Errorf("expected response code to be: %d got: %d",
			want, got)
	}
}

func checkContentType(want string, rr *httptest.ResponseRecorder, t *testing.T) {
	if want != rr.Header().Get("Content-Type") {
		t.Errorf("expected heder to be: %s got: %s",
			want, rr.Header().Get("Content-Type"))
	}
}
