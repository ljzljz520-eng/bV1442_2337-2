package tickets

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubCleaner struct {
	cleaned string
	err     error
}

func (s stubCleaner) Clean(string) (string, error) {
	return s.cleaned, s.err
}

func requestBody(t *testing.T, description string) *bytes.Reader {
	t.Helper()
	body, err := json.Marshal(cleanRequest{Description: description})
	if err != nil {
		t.Fatal(err)
	}
	return bytes.NewReader(body)
}

func TestCleanTicketReturnsStatisticsForDescription(t *testing.T) {
	handler := NewHandler(NewLocalTextCleaner())
	req := httptest.NewRequest(http.MethodPost, "/api/tickets/clean", requestBody(t, "<b>快递</b> 已 到达"))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected successful response, got %d", recorder.Code)
	}
	var result Result
	if err := json.NewDecoder(recorder.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.CleanedText != "快递 已 到达" {
		t.Fatalf("expected cleaned description, got %q", result.CleanedText)
	}
	if result.Statistics.Characters != 7 {
		t.Fatalf("expected 7 characters, got %d", result.Statistics.Characters)
	}
	if result.Statistics.Words != 5 {
		t.Fatalf("expected 5 words, got %d", result.Statistics.Words)
	}
	if result.Statistics.Category != "物流配送" {
		t.Fatalf("expected logistics category, got %q", result.Statistics.Category)
	}
}

func TestCleanTicketRejectsWhitespaceOnlyDescription(t *testing.T) {
	handler := NewHandler(stubCleaner{cleaned: ""})
	req := httptest.NewRequest(http.MethodPost, "/api/tickets/clean", requestBody(t, " \n\t "))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected empty description to be rejected, got %d", recorder.Code)
	}
	var response errorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.Error != ErrDescriptionEmpty.Error() {
		t.Fatalf("expected empty description error, got %q", response.Error)
	}
}

func TestCleanTicketReportsUnavailableCleaner(t *testing.T) {
	handler := NewHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/tickets/clean", requestBody(t, "需要处理"))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected unavailable cleaner to fail, got %d", recorder.Code)
	}
	var response errorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.Error != ErrCleanerUnavailable.Error() {
		t.Fatalf("expected unavailable cleaner error, got %q", response.Error)
	}
}

func TestCleanTicketReportsCleanerFailure(t *testing.T) {
	handler := NewHandler(stubCleaner{err: errors.New("disk unavailable")})
	req := httptest.NewRequest(http.MethodPost, "/api/tickets/clean", requestBody(t, "物流延迟"))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected cleaner failure to fail, got %d", recorder.Code)
	}
	var response errorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.Error != "文本清理失败" {
		t.Fatalf("expected cleaner failure message, got %q", response.Error)
	}
}
