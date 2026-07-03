package email_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/infra/email"
	"github.com/stretchr/testify/require"
)

func TestBrevoNotifier_NotifyStatusChanged_SendsEmail(t *testing.T) {
	t.Parallel()

	var gotAPIKey string
	var gotPayload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("api-key")
		require.Equal(t, http.MethodPost, r.Method)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	notifier, err := email.NewBrevoNotifier(email.BrevoConfig{
		APIKey:      "test-key",
		SenderEmail: "sender@example.com",
		SenderName:  "Oficina Teste",
		APIURL:      server.URL,
	}, server.Client())
	require.NoError(t, err)

	err = notifier.NotifyStatusChanged(context.Background(), repository.StatusChangedNotification{
		ServiceOrderID: "so1",
		Code:           "OS-1",
		ClientName:     "Gustavo",
		ClientEmail:    "gustavo@example.com",
		CurrentStatus:  "IN_PROGRESS",
	})
	require.NoError(t, err)
	require.Equal(t, "test-key", gotAPIKey)
	require.Equal(t, "Oficina Teste", gotPayload["sender"].(map[string]any)["name"])
	require.Equal(t, "sender@example.com", gotPayload["sender"].(map[string]any)["email"])
	require.Equal(t, "Oficina - Pos | OS OS-1 atualizada", gotPayload["subject"])
	require.Contains(t, gotPayload["htmlContent"], "Em execucao")
	to := gotPayload["to"].([]any)[0].(map[string]any)
	require.Equal(t, "gustavo@example.com", to["email"])
	require.Equal(t, "Gustavo", to["name"])
}

func TestBrevoNotifier_NotifyStatusChanged_HTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer server.Close()

	notifier, err := email.NewBrevoNotifier(email.BrevoConfig{
		APIKey:      "test-key",
		SenderEmail: "sender@example.com",
		APIURL:      server.URL,
	}, server.Client())
	require.NoError(t, err)

	err = notifier.NotifyStatusChanged(context.Background(), repository.StatusChangedNotification{
		Code:          "OS-1",
		ClientEmail:   "gustavo@example.com",
		CurrentStatus: "FINISHED",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "status=400")
}

func TestNewBrevoNotifier_Validation(t *testing.T) {
	t.Parallel()

	_, err := email.NewBrevoNotifier(email.BrevoConfig{SenderEmail: "sender@example.com"}, nil)
	require.Error(t, err)

	_, err = email.NewBrevoNotifier(email.BrevoConfig{APIKey: "key"}, nil)
	require.Error(t, err)
}
