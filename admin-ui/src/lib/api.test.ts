import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

// We need to test the ApiClient class. Since `api` is an exported singleton using
// import.meta.env, let's create our own instance for testing.
// First, let's mock localStorage and fetch.

describe('ApiClient', () => {
  const mockFetch = vi.fn()
  let mockLocalStorage: Record<string, string>

  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch)
    mockFetch.mockReset()

    // Mock localStorage
    mockLocalStorage = {}
    vi.stubGlobal('localStorage', {
      getItem: (key: string) => mockLocalStorage[key] ?? null,
      setItem: (key: string, value: string) => {
        mockLocalStorage[key] = value
      },
      removeItem: (key: string) => {
        delete mockLocalStorage[key]
      },
    })

    // Mock window.location
    const locationMock = { href: '' }
    vi.stubGlobal('location', locationMock)
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  function mockJsonResponse(data: unknown, status = 200) {
    return {
      ok: status >= 200 && status < 300,
      status,
      statusText: 'OK',
      headers: new Headers({ 'content-type': 'application/json' }),
      json: () => Promise.resolve(data),
    }
  }

  function mockErrorResponse(data: unknown, status: number) {
    return {
      ok: false,
      status,
      statusText: 'Error',
      headers: new Headers({ 'content-type': 'application/json' }),
      json: () => Promise.resolve(data),
    }
  }

  // ── Login ──

  describe('login', () => {
    it('should POST to /admin/auth/login and store tokens', async () => {
      const tokens = {
        access_token: 'at123',
        refresh_token: 'rt456',
        expires_in: 3600,
        token_type: 'Bearer',
      }
      mockFetch.mockResolvedValueOnce(mockJsonResponse(tokens))

      // Import fresh each time to get a new module with our mocked globals
      const { api } = await import('./api')
      const result = await api.login('admin@test.com', 'password')

      expect(mockFetch).toHaveBeenCalledOnce()
      const [url, options] = mockFetch.mock.calls[0]
      expect(url).toContain('/admin/auth/login')
      expect(options.method).toBe('POST')
      expect(JSON.parse(options.body)).toEqual({
        email: 'admin@test.com',
        password: 'password',
      })
      expect(result).toEqual(tokens)
      expect(mockLocalStorage['access_token']).toBe('at123')
      expect(mockLocalStorage['refresh_token']).toBe('rt456')
    })

    it('should throw on login failure', async () => {
      mockFetch.mockResolvedValueOnce(
        mockErrorResponse(
          { error: 'invalid_credentials', message: 'Invalid credentials', status_code: 401 },
          401
        )
      )

      const { api } = await import('./api')
      await expect(api.login('admin@test.com', 'wrong')).rejects.toMatchObject({
        error: 'invalid_credentials',
      })
    })
  })

  // ── JWT header injection ──

  describe('JWT header injection', () => {
    it('should include Authorization header with Bearer token', async () => {
      mockLocalStorage['access_token'] = 'my-jwt-token'
      mockLocalStorage['token_expires_at'] = String(Date.now() + 60000)

      const userData = { id: '1', email: 'admin@test.com', name: 'Admin', role: 'admin' }
      mockFetch.mockResolvedValueOnce(mockJsonResponse(userData))

      const { api } = await import('./api')
      await api.getMe()

      const [, options] = mockFetch.mock.calls[0]
      expect(options.headers['Authorization']).toBe('Bearer my-jwt-token')
    })
  })

  // ── Logout ──

  describe('logout', () => {
    it('should POST to /admin/auth/logout and clear tokens', async () => {
      mockLocalStorage['access_token'] = 'at'
      mockLocalStorage['refresh_token'] = 'rt'
      mockLocalStorage['token_expires_at'] = String(Date.now() + 60000)

      mockFetch.mockResolvedValueOnce(mockJsonResponse(undefined, 204))

      const { api } = await import('./api')
      await api.logout()

      expect(mockLocalStorage['access_token']).toBeUndefined()
      expect(mockLocalStorage['refresh_token']).toBeUndefined()
      expect(mockLocalStorage['token_expires_at']).toBeUndefined()
    })
  })

  // ── Error handling ──

  describe('error handling', () => {
    it('should throw ApiError on non-ok response', async () => {
      mockLocalStorage['access_token'] = 'at'
      mockLocalStorage['token_expires_at'] = String(Date.now() + 60000)

      mockFetch.mockResolvedValueOnce(
        mockErrorResponse(
          { error: 'not_found', message: 'User not found', status_code: 404 },
          404
        )
      )

      const { api } = await import('./api')
      await expect(api.getUser('nonexistent')).rejects.toMatchObject({
        error: 'not_found',
        message: 'User not found',
      })
    })

    it('should handle non-JSON error responses', async () => {
      mockLocalStorage['access_token'] = 'at'
      mockLocalStorage['token_expires_at'] = String(Date.now() + 60000)

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: () => Promise.reject(new Error('not json')),
      })

      const { api } = await import('./api')
      await expect(api.getUsers()).rejects.toMatchObject({
        error: 'request_failed',
      })
    })

    it('should handle 204 No Content responses', async () => {
      mockLocalStorage['access_token'] = 'at'
      mockLocalStorage['token_expires_at'] = String(Date.now() + 60000)

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 204,
        headers: new Headers(),
        json: () => Promise.reject(new Error('no body')),
      })

      const { api } = await import('./api')
      const result = await api.deleteUser('user1')
      expect(result).toBeUndefined()
    })
  })

  // ── Pagination params ──

  describe('pagination params', () => {
    it('should pass cursor, limit, search, status, sort, order to getUsers', async () => {
      mockLocalStorage['access_token'] = 'at'
      mockLocalStorage['token_expires_at'] = String(Date.now() + 60000)

      mockFetch.mockResolvedValueOnce(
        mockJsonResponse({ data: [], has_more: false })
      )

      const { api } = await import('./api')
      await api.getUsers({
        cursor: 'c1',
        limit: 25,
        search: 'john',
        status: 'active',
        sort: 'created_at',
        order: 'desc',
      })

      const url = mockFetch.mock.calls[0][0] as string
      expect(url).toContain('cursor=c1')
      expect(url).toContain('limit=25')
      expect(url).toContain('search=john')
      expect(url).toContain('status=active')
      expect(url).toContain('sort=created_at')
      expect(url).toContain('order=desc')
    })

    it('should not include undefined params', async () => {
      mockLocalStorage['access_token'] = 'at'
      mockLocalStorage['token_expires_at'] = String(Date.now() + 60000)

      mockFetch.mockResolvedValueOnce(
        mockJsonResponse({ data: [], has_more: false })
      )

      const { api } = await import('./api')
      await api.getUsers({ search: 'test' })

      const url = mockFetch.mock.calls[0][0] as string
      expect(url).toContain('search=test')
      expect(url).not.toContain('cursor')
      expect(url).not.toContain('limit')
    })

    it('should call getUsers without params', async () => {
      mockLocalStorage['access_token'] = 'at'
      mockLocalStorage['token_expires_at'] = String(Date.now() + 60000)

      mockFetch.mockResolvedValueOnce(
        mockJsonResponse({ data: [], has_more: false })
      )

      const { api } = await import('./api')
      await api.getUsers()

      const url = mockFetch.mock.calls[0][0] as string
      expect(url).toContain('/admin/users')
      expect(url).not.toContain('?')
    })
  })

  // ── CRUD methods ──

  describe('CRUD methods', () => {
    beforeEach(() => {
      mockLocalStorage['access_token'] = 'at'
      mockLocalStorage['token_expires_at'] = String(Date.now() + 60000)
    })

    it('getUser should GET /admin/users/:id', async () => {
      mockFetch.mockResolvedValueOnce(
        mockJsonResponse({ id: 'u1', email: 'test@test.com', name: 'Test' })
      )

      const { api } = await import('./api')
      const result = await api.getUser('u1')

      expect(mockFetch.mock.calls[0][0]).toContain('/admin/users/u1')
      expect(result.id).toBe('u1')
    })

    it('createUser should POST /admin/users', async () => {
      mockFetch.mockResolvedValueOnce(
        mockJsonResponse({ id: 'u2', email: 'new@test.com', name: 'New' })
      )

      const { api } = await import('./api')
      await api.createUser({ email: 'new@test.com', name: 'New', password: 'pass' })

      const [url, options] = mockFetch.mock.calls[0]
      expect(url).toContain('/admin/users')
      expect(options.method).toBe('POST')
    })

    it('updateUser should PATCH /admin/users/:id', async () => {
      mockFetch.mockResolvedValueOnce(
        mockJsonResponse({ id: 'u1', email: 'test@test.com', name: 'Updated' })
      )

      const { api } = await import('./api')
      await api.updateUser('u1', { name: 'Updated' })

      const [url, options] = mockFetch.mock.calls[0]
      expect(url).toContain('/admin/users/u1')
      expect(options.method).toBe('PATCH')
    })

    it('deleteUser should DELETE /admin/users/:id', async () => {
      mockFetch.mockResolvedValueOnce(mockJsonResponse(undefined, 204))

      const { api } = await import('./api')
      await api.deleteUser('u1')

      const [url, options] = mockFetch.mock.calls[0]
      expect(url).toContain('/admin/users/u1')
      expect(options.method).toBe('DELETE')
    })

    it('blockUser should POST /admin/users/:id/block', async () => {
      mockFetch.mockResolvedValueOnce(
        mockJsonResponse({ id: 'u1', status: 'blocked' })
      )

      const { api } = await import('./api')
      await api.blockUser('u1')

      const [url, options] = mockFetch.mock.calls[0]
      expect(url).toContain('/admin/users/u1/block')
      expect(options.method).toBe('POST')
    })

    it('unblockUser should POST /admin/users/:id/unblock', async () => {
      mockFetch.mockResolvedValueOnce(
        mockJsonResponse({ id: 'u1', status: 'active' })
      )

      const { api } = await import('./api')
      await api.unblockUser('u1')

      const [url, options] = mockFetch.mock.calls[0]
      expect(url).toContain('/admin/users/u1/unblock')
      expect(options.method).toBe('POST')
    })

    it('getApplications should GET /admin/applications', async () => {
      mockFetch.mockResolvedValueOnce(mockJsonResponse([{ id: 'a1', name: 'App' }]))

      const { api } = await import('./api')
      const result = await api.getApplications()

      expect(mockFetch.mock.calls[0][0]).toContain('/admin/applications')
      expect(result).toHaveLength(1)
    })

    it('createApplication should POST /admin/applications', async () => {
      mockFetch.mockResolvedValueOnce(mockJsonResponse({ id: 'a2', name: 'NewApp' }))

      const { api } = await import('./api')
      await api.createApplication({ name: 'NewApp' })

      expect(mockFetch.mock.calls[0][1].method).toBe('POST')
    })

    it('getRoles should GET /admin/roles', async () => {
      mockFetch.mockResolvedValueOnce(mockJsonResponse([{ id: 'r1', name: 'admin' }]))

      const { api } = await import('./api')
      const result = await api.getRoles()

      expect(result).toHaveLength(1)
    })

    it('getWebhooks should GET /admin/webhooks', async () => {
      mockFetch.mockResolvedValueOnce(mockJsonResponse([]))

      const { api } = await import('./api')
      const result = await api.getWebhooks()

      expect(result).toEqual([])
    })

    it('getDashboardMetrics should GET /admin/dashboard/metrics', async () => {
      const metrics = { active_users: 100, login_success_rate: 99.5 }
      mockFetch.mockResolvedValueOnce(mockJsonResponse(metrics))

      const { api } = await import('./api')
      const result = await api.getDashboardMetrics()

      expect(result).toMatchObject(metrics)
    })
  })
})
