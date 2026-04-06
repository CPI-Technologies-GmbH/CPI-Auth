package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// inMemoryTenantRepo is a tiny TenantRepository implementation used by the
// resolver tests so they don't pull in core/db.
type inMemoryTenantRepo struct {
	bySlug   map[string]*models.Tenant
	byDomain map[string]*models.Tenant
}

func newInMemoryTenantRepo() *inMemoryTenantRepo {
	return &inMemoryTenantRepo{
		bySlug:   make(map[string]*models.Tenant),
		byDomain: make(map[string]*models.Tenant),
	}
}

func (r *inMemoryTenantRepo) seed(t *models.Tenant) {
	r.bySlug[t.Slug] = t
	if t.Domain != "" {
		r.byDomain[t.Domain] = t
	}
}

func (r *inMemoryTenantRepo) Create(_ context.Context, _ *models.Tenant) error { return nil }
func (r *inMemoryTenantRepo) GetByID(_ context.Context, id uuid.UUID) (*models.Tenant, error) {
	for _, t := range r.bySlug {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, models.ErrNotFound
}
func (r *inMemoryTenantRepo) GetBySlug(_ context.Context, slug string) (*models.Tenant, error) {
	if t, ok := r.bySlug[slug]; ok {
		return t, nil
	}
	return nil, models.ErrNotFound
}
func (r *inMemoryTenantRepo) GetByDomain(_ context.Context, domain string) (*models.Tenant, error) {
	if t, ok := r.byDomain[domain]; ok {
		return t, nil
	}
	return nil, models.ErrNotFound
}
func (r *inMemoryTenantRepo) Update(_ context.Context, _ *models.Tenant) error { return nil }
func (r *inMemoryTenantRepo) Delete(_ context.Context, _ uuid.UUID) error      { return nil }
func (r *inMemoryTenantRepo) List(_ context.Context, _ models.PaginationParams) (*models.PaginatedResult[models.Tenant], error) {
	return &models.PaginatedResult[models.Tenant]{}, nil
}

// captureHandler records the tenant id seen + final URL path so tests can
// assert on what the downstream handler observed after the resolver ran.
type captureHandler struct {
	tenantID   uuid.UUID
	tenantSlug string
	path       string
	called     bool
}

func (c *captureHandler) ServeHTTP(_ http.ResponseWriter, r *http.Request) {
	c.called = true
	c.tenantID = GetTenantID(r.Context())
	c.tenantSlug = GetTenantSlug(r.Context())
	c.path = r.URL.Path
}

func TestPathBasedTenantResolver_StripsPrefixAndSetsTenant(t *testing.T) {
	repo := newInMemoryTenantRepo()
	tenant := &models.Tenant{ID: uuid.New(), Name: "LastSoftware", Slug: "lastsoftware"}
	repo.seed(tenant)

	cap := &captureHandler{}
	h := PathBasedTenantResolver(repo, cap)

	req := httptest.NewRequest(http.MethodGet, "/t/lastsoftware/oauth/authorize", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if !cap.called {
		t.Fatal("downstream handler not called")
	}
	if cap.tenantID != tenant.ID {
		t.Errorf("tenant id = %v, want %v", cap.tenantID, tenant.ID)
	}
	if cap.tenantSlug != "lastsoftware" {
		t.Errorf("tenant slug = %q, want %q", cap.tenantSlug, "lastsoftware")
	}
	if cap.path != "/oauth/authorize" {
		t.Errorf("downstream path = %q, want %q", cap.path, "/oauth/authorize")
	}
}

func TestPathBasedTenantResolver_PassesThroughLegacyPaths(t *testing.T) {
	repo := newInMemoryTenantRepo()
	cap := &captureHandler{}
	h := PathBasedTenantResolver(repo, cap)

	req := httptest.NewRequest(http.MethodGet, "/oauth/authorize", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if !cap.called {
		t.Fatal("downstream handler not called for legacy path")
	}
	if cap.tenantID != uuid.Nil {
		t.Errorf("tenant id should be nil for legacy path, got %v", cap.tenantID)
	}
	if cap.path != "/oauth/authorize" {
		t.Errorf("downstream path = %q, want unchanged", cap.path)
	}
}

func TestPathBasedTenantResolver_404ForUnknownSlug(t *testing.T) {
	repo := newInMemoryTenantRepo()
	cap := &captureHandler{}
	h := PathBasedTenantResolver(repo, cap)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/t/nonexistent/oauth/authorize", nil)
	h.ServeHTTP(rec, req)

	if cap.called {
		t.Fatal("downstream handler should NOT be called when slug is unknown")
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestPathBasedTenantResolver_PassesThroughForInvalidSlugFormat(t *testing.T) {
	// /t/INVALID/... contains uppercase which is not a valid slug. The resolver
	// must NOT try to look it up and must NOT 404 — it should pass through so
	// the request can hit a normal handler (or 404 there). This protects
	// against the resolver swallowing legitimately-malformed URLs.
	repo := newInMemoryTenantRepo()
	cap := &captureHandler{}
	h := PathBasedTenantResolver(repo, cap)

	req := httptest.NewRequest(http.MethodGet, "/t/Bad_Slug/oauth/authorize", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if !cap.called {
		t.Fatal("downstream handler should be called for invalid slug")
	}
	if cap.path != "/t/Bad_Slug/oauth/authorize" {
		t.Errorf("path should be unchanged for invalid slug, got %q", cap.path)
	}
}

func TestPathBasedTenantResolver_NoSlashAfterSlug(t *testing.T) {
	// /t/lastsoftware (no trailing slash) should still resolve and rewrite to "/".
	repo := newInMemoryTenantRepo()
	tenant := &models.Tenant{ID: uuid.New(), Slug: "lastsoftware"}
	repo.seed(tenant)

	cap := &captureHandler{}
	h := PathBasedTenantResolver(repo, cap)

	req := httptest.NewRequest(http.MethodGet, "/t/lastsoftware", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if cap.tenantID != tenant.ID {
		t.Errorf("tenant id = %v, want %v", cap.tenantID, tenant.ID)
	}
	if cap.path != "/" {
		t.Errorf("path = %q, want %q", cap.path, "/")
	}
}

func TestTenantResolver_RejectsReservedSubdomain(t *testing.T) {
	// Pre-Phase-0 a tenant with slug "auth" could capture auth.cpi.dev. The
	// reserved-subdomain blacklist must prevent that even if such a tenant
	// somehow gets created.
	repo := newInMemoryTenantRepo()
	repo.seed(&models.Tenant{ID: uuid.New(), Slug: "auth"})

	cap := &captureHandler{}
	wrapped := TenantResolver(repo)(cap)

	req := httptest.NewRequest(http.MethodGet, "/oauth/authorize", nil)
	req.Host = "auth.cpi.dev"
	wrapped.ServeHTTP(httptest.NewRecorder(), req)

	if cap.tenantID != uuid.Nil {
		t.Errorf("tenant id should be nil for reserved subdomain, got %v", cap.tenantID)
	}
}

func TestTenantResolver_AllowsValidSubdomain(t *testing.T) {
	repo := newInMemoryTenantRepo()
	tenant := &models.Tenant{ID: uuid.New(), Slug: "lastsoftware"}
	repo.seed(tenant)

	cap := &captureHandler{}
	wrapped := TenantResolver(repo)(cap)

	req := httptest.NewRequest(http.MethodGet, "/oauth/authorize", nil)
	req.Host = "lastsoftware.auth.cpi.dev"
	wrapped.ServeHTTP(httptest.NewRecorder(), req)

	if cap.tenantID != tenant.ID {
		t.Errorf("tenant id = %v, want %v", cap.tenantID, tenant.ID)
	}
}

func TestTenantResolver_CustomDomain(t *testing.T) {
	repo := newInMemoryTenantRepo()
	tenant := &models.Tenant{ID: uuid.New(), Slug: "lastsoftware", Domain: "login.lastsoftware.com"}
	repo.seed(tenant)

	cap := &captureHandler{}
	wrapped := TenantResolver(repo)(cap)

	req := httptest.NewRequest(http.MethodGet, "/oauth/authorize", nil)
	req.Host = "login.lastsoftware.com"
	wrapped.ServeHTTP(httptest.NewRecorder(), req)

	if cap.tenantID != tenant.ID {
		t.Errorf("tenant id = %v, want %v", cap.tenantID, tenant.ID)
	}
}

func TestTenantResolver_HeaderWins(t *testing.T) {
	repo := newInMemoryTenantRepo()
	wantedID := uuid.New()
	repo.seed(&models.Tenant{ID: wantedID, Slug: "explicit"})

	cap := &captureHandler{}
	wrapped := TenantResolver(repo)(cap)

	req := httptest.NewRequest(http.MethodGet, "/oauth/authorize", nil)
	req.Header.Set("X-Tenant-ID", wantedID.String())
	req.Host = "auth.cpi.dev" // would otherwise be reserved
	wrapped.ServeHTTP(httptest.NewRecorder(), req)

	if cap.tenantID != wantedID {
		t.Errorf("tenant id = %v, want %v", cap.tenantID, wantedID)
	}
}

func TestTenantResolver_InvalidHeaderIsIgnored(t *testing.T) {
	repo := newInMemoryTenantRepo()

	cap := &captureHandler{}
	wrapped := TenantResolver(repo)(cap)

	req := httptest.NewRequest(http.MethodGet, "/oauth/authorize", nil)
	req.Header.Set("X-Tenant-ID", "not-a-uuid")
	wrapped.ServeHTTP(httptest.NewRecorder(), req)

	if cap.tenantID != uuid.Nil {
		t.Errorf("tenant id should be nil for invalid header, got %v", cap.tenantID)
	}
}
