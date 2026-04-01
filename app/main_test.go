package main

import (
	"bytes"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeChain struct {
	value  *big.Int
	setErr error
	getErr error
}

func (f *fakeChain) SetValue(v *big.Int) error {
	if f.setErr != nil {
		return f.setErr
	}
	f.value = new(big.Int).Set(v)
	return nil
}

func (f *fakeChain) GetValue() (*big.Int, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if f.value == nil {
		return big.NewInt(0), nil
	}
	return new(big.Int).Set(f.value), nil
}

type fakeStore struct {
	value   string
	saveErr error
	getErr  error
}

func (f *fakeStore) SaveValue(val string) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.value = val
	return nil
}

func (f *fakeStore) GetSavedValue() (string, error) {
	if f.getErr != nil {
		return "", f.getErr
	}
	return f.value, nil
}

func TestSetRouteStatusCodes(t *testing.T) {
	t.Run("returns 200 when write succeeds", func(t *testing.T) {
		router := setupRouter(&fakeChain{}, &fakeStore{})
		req := httptest.NewRequest(http.MethodPost, "/set", bytes.NewBufferString(`{"value": 42}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("returns 500 when blockchain write fails", func(t *testing.T) {
		router := setupRouter(&fakeChain{setErr: errors.New("boom")}, &fakeStore{})
		req := httptest.NewRequest(http.MethodPost, "/set", bytes.NewBufferString(`{"value": 42}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}

func TestGetRouteStatusCodes(t *testing.T) {
	t.Run("returns 200 when read succeeds", func(t *testing.T) {
		router := setupRouter(&fakeChain{value: big.NewInt(99)}, &fakeStore{})
		req := httptest.NewRequest(http.MethodGet, "/get", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("returns 500 when blockchain read fails", func(t *testing.T) {
		router := setupRouter(&fakeChain{getErr: errors.New("boom")}, &fakeStore{})
		req := httptest.NewRequest(http.MethodGet, "/get", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}

func TestSyncRouteStatusCodes(t *testing.T) {
	t.Run("returns 200 when sync succeeds", func(t *testing.T) {
		router := setupRouter(&fakeChain{value: big.NewInt(7)}, &fakeStore{})
		req := httptest.NewRequest(http.MethodPost, "/sync", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("returns 500 when DB save fails", func(t *testing.T) {
		router := setupRouter(&fakeChain{value: big.NewInt(7)}, &fakeStore{saveErr: errors.New("boom")})
		req := httptest.NewRequest(http.MethodPost, "/sync", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}

func TestCheckRouteStatusCodes(t *testing.T) {
	t.Run("returns 200 when comparison succeeds", func(t *testing.T) {
		router := setupRouter(&fakeChain{value: big.NewInt(7)}, &fakeStore{value: "7"})
		req := httptest.NewRequest(http.MethodGet, "/check", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("returns 500 when DB read fails", func(t *testing.T) {
		router := setupRouter(&fakeChain{value: big.NewInt(7)}, &fakeStore{getErr: errors.New("boom")})
		req := httptest.NewRequest(http.MethodGet, "/check", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}
