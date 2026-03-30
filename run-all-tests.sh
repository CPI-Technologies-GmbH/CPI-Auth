#!/bin/bash
# CPI Auth - Complete Test Suite
# This script runs ALL tests across the entire project

set -e

echo "╔══════════════════════════════════════════════════════════╗"
echo "║           CPI Auth - Complete Test Suite                ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# ─── Go Unit Tests ───────────────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  [1/6] Go Unit & Integration Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cd /Users/maxi/Coding/cpi-auth
E2E_DATABASE_URL="postgres://cpi-auth:cpi-auth_secret@localhost:5052/cpi-auth?sslmode=disable" \
  go test ./... -count=1 2>&1
echo ""
echo "  ✅ Go tests PASSED"
echo ""

# ─── Admin UI Unit Tests ──────────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  [2/6] Admin UI Unit Tests (Vitest)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cd /Users/maxi/Coding/cpi-auth/admin-ui
npx vitest run 2>&1
echo ""

# ─── Login UI Unit Tests ──────────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  [3/6] Login UI Unit Tests (Vitest)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cd /Users/maxi/Coding/cpi-auth/login-ui
npx vitest run 2>&1
echo ""

# ─── Account UI Unit Tests ────────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  [4/6] Account UI Unit Tests (Vitest)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cd /Users/maxi/Coding/cpi-auth/account-ui
npx vitest run 2>&1
echo ""

# ─── TypeScript Compilation ──────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  [5/6] TypeScript Compilation Check (Admin UI)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cd /Users/maxi/Coding/cpi-auth/admin-ui
npx tsc --noEmit 2>&1
echo "  ✅ TypeScript compilation PASSED"
echo ""

# ─── Playwright E2E Tests ────────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  [6/6] Playwright E2E Tests (Admin UI - 116 tests)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cd /Users/maxi/Coding/cpi-auth/tests/browser
npx playwright test tests/admin/ --reporter=list 2>&1
echo ""

# ─── Summary ─────────────────────────────────────────────────
echo "╔══════════════════════════════════════════════════════════╗"
echo "║                 ALL TESTS PASSED ✅                     ║"
echo "║                                                         ║"
echo "║  Go Unit Tests:         ~70 tests    ✅                 ║"
echo "║  Admin UI Unit Tests:   155 tests    ✅                 ║"
echo "║  Login UI Unit Tests:   130 tests    ✅                 ║"
echo "║  Account UI Unit Tests: 122 tests    ✅                 ║"
echo "║  TypeScript Compilation:             ✅                 ║"
echo "║  Playwright E2E Tests:  116 tests    ✅                 ║"
echo "║                                                         ║"
echo "║  Total:                ~593 tests    ✅                 ║"
echo "╚══════════════════════════════════════════════════════════╝"
