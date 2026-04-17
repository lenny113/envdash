package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"assignment-2/internal/models"
	"assignment-2/internal/utils"
)

//
// helpers

// webhookCapture spins up an httptest.Server that stores the last payload
// received and the total number of calls.
type webhookCapture struct {
	*httptest.Server
	Calls   int
	Payload map[string]interface{}
}

func newWebhookCapture(t *testing.T) *webhookCapture {
	t.Helper()
	wc := &webhookCapture{}
	wc.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wc.Calls++
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &wc.Payload)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(wc.Server.Close)
	return wc
}

// failingWebhookServer returns a server that always responds with 500.
func failingWebhookServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func makeAllRegisteredWebhook(id, country, event, url string, th *models.ThresholdNotification) models.AllRegisteredWebhook {
	return models.AllRegisteredWebhook{
		Id: id,
		RegisterWebhook: models.RegisterWebhook{
			Url:                   url,
			Country:               country,
			Event:                 event,
			ThresholdNotification: th,
		},
	}
}

// postWebhook – unit tests
//

func TestPostWebhook_Success(t *testing.T) {
	wc := newWebhookCapture(t)
	payload := map[string]interface{}{"id": "test", "event": "REGISTER"}
	err := postWebhook(wc.URL, payload)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if wc.Calls != 1 {
		t.Errorf("expected 1 call to webhook, got %d", wc.Calls)
	}
}

func TestPostWebhook_ServerError(t *testing.T) {
	srv := failingWebhookServer(t)
	payload := map[string]interface{}{"id": "x"}
	err := postWebhook(srv.URL, payload)
	if err == nil {
		t.Error("expected error when server returns 500")
	}
}

func TestPostWebhook_InvalidURL(t *testing.T) {
	payload := map[string]interface{}{"id": "x"}
	err := postWebhook("http://127.0.0.1:0", payload) // nothing listening
	if err == nil {
		t.Error("expected error for unreachable URL")
	}
}

func TestPostWebhook_PayloadFields(t *testing.T) {
	wc := newWebhookCapture(t)
	payload := map[string]interface{}{
		"id":      "abc",
		"country": "NO",
		"event":   "REGISTER",
		"time":    "20240101 12:00",
	}
	if err := postWebhook(wc.URL, payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for k, v := range payload {
		if wc.Payload[k] != v {
			t.Errorf("field %q: expected %v, got %v", k, v, wc.Payload[k])
		}
	}
}

// sendingLifeCycleWebhook

func TestSendingLifeCycleWebhook_Success(t *testing.T) {
	wc := newWebhookCapture(t)
	n := models.RegisterWebhook{
		Url:     wc.URL,
		Country: "NO",
		Event:   "REGISTER",
	}
	err := sendingLifeCycleWebhook("some-id", n)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wc.Payload["id"] != "some-id" {
		t.Errorf("expected id 'some-id', got %v", wc.Payload["id"])
	}
	if wc.Payload["country"] != "NO" {
		t.Errorf("expected country 'NO', got %v", wc.Payload["country"])
	}
	if wc.Payload["event"] != "REGISTER" {
		t.Errorf("expected event 'REGISTER', got %v", wc.Payload["event"])
	}
}

func TestSendingLifeCycleWebhook_ServerError(t *testing.T) {
	srv := failingWebhookServer(t)
	n := models.RegisterWebhook{Url: srv.URL, Country: "NO", Event: "REGISTER"}
	err := sendingLifeCycleWebhook("id", n)
	if err == nil {
		t.Error("expected error when webhook server fails")
	}
}

// sendThresholdWebhook

func TestSendThresholdWebhook_Success(t *testing.T) {
	wc := newWebhookCapture(t)
	details := models.ThresholdDetails{
		Field:          utils.VALIDTHRESHOLDS[0],
		Operator:       ">",
		ThresholdValue: 20.0,
		MeasuredValue:  25.0,
	}
	err := sendThresholdWebhook("id-1", "NO", wc.URL, details)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wc.Payload["event"] != "THRESHOLD" {
		t.Errorf("expected event THRESHOLD, got %v", wc.Payload["event"])
	}
	detailsMap, ok := wc.Payload["details"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'details' field in payload")
	}
	if detailsMap["measuredValue"].(float64) != 25.0 {
		t.Errorf("expected measuredValue 25.0, got %v", detailsMap["measuredValue"])
	}
}

func TestSendThresholdWebhook_ContainsAllRequiredFields(t *testing.T) {
	wc := newWebhookCapture(t)
	details := models.ThresholdDetails{
		Field: utils.VALIDTHRESHOLDS[0], Operator: "<", ThresholdValue: 10.0, MeasuredValue: 5.0,
	}
	sendThresholdWebhook("id-x", "SE", wc.URL, details)

	requiredTopLevel := []string{"id", "country", "event", "time", "details"}
	for _, key := range requiredTopLevel {
		if _, ok := wc.Payload[key]; !ok {
			t.Errorf("expected key %q in payload", key)
		}
	}
	dm := wc.Payload["details"].(map[string]interface{})
	for _, key := range []string{"field", "operator", "threshold", "measuredValue"} {
		if _, ok := dm[key]; !ok {
			t.Errorf("expected key %q in details", key)
		}
	}
}

// CheckLifecycleNotifications

func TestCheckLifecycleNotifications_MatchingCountryAndEvent(t *testing.T) {
	wc := newWebhookCapture(t)
	notifications := []models.AllRegisteredWebhook{
		makeAllRegisteredWebhook("id-1", "NO", "REGISTER", wc.URL, nil),
	}
	store := &MockStore{
		GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
			return notifications, nil
		},
	}
	h := newMockHandler(store)
	h.CheckLifecycleNotifications(context.Background(), "NO", "REGISTER")

	if wc.Calls != 1 {
		t.Errorf("expected 1 webhook call, got %d", wc.Calls)
	}
}

func TestCheckLifecycleNotifications_CountryMismatch(t *testing.T) {
	wc := newWebhookCapture(t)
	notifications := []models.AllRegisteredWebhook{
		makeAllRegisteredWebhook("id-1", "SE", "REGISTER", wc.URL, nil),
	}
	store := &MockStore{
		GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
			return notifications, nil
		},
	}
	h := newMockHandler(store)
	h.CheckLifecycleNotifications(context.Background(), "NO", "REGISTER")

	if wc.Calls != 0 {
		t.Errorf("expected 0 calls for country mismatch, got %d", wc.Calls)
	}
}

func TestCheckLifecycleNotifications_EventMismatch(t *testing.T) {
	wc := newWebhookCapture(t)
	notifications := []models.AllRegisteredWebhook{
		makeAllRegisteredWebhook("id-1", "NO", "DELETE", wc.URL, nil),
	}
	store := &MockStore{
		GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
			return notifications, nil
		},
	}
	h := newMockHandler(store)
	h.CheckLifecycleNotifications(context.Background(), "NO", "REGISTER")

	if wc.Calls != 0 {
		t.Errorf("expected 0 calls for event mismatch, got %d", wc.Calls)
	}
}

func TestCheckLifecycleNotifications_EmptyCountryMatchesAll(t *testing.T) {
	wc := newWebhookCapture(t)
	// notification with empty country = wildcard
	notifications := []models.AllRegisteredWebhook{
		makeAllRegisteredWebhook("id-1", "", "REGISTER", wc.URL, nil),
	}
	store := &MockStore{
		GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
			return notifications, nil
		},
	}
	h := newMockHandler(store)
	h.CheckLifecycleNotifications(context.Background(), "NO", "REGISTER")

	if wc.Calls != 1 {
		t.Errorf("expected 1 call for wildcard country, got %d", wc.Calls)
	}
}

func TestCheckLifecycleNotifications_MultipleNotifications(t *testing.T) {
	wc1 := newWebhookCapture(t)
	wc2 := newWebhookCapture(t)
	notifications := []models.AllRegisteredWebhook{
		makeAllRegisteredWebhook("id-1", "NO", "REGISTER", wc1.URL, nil),
		makeAllRegisteredWebhook("id-2", "NO", "REGISTER", wc2.URL, nil),
		makeAllRegisteredWebhook("id-3", "SE", "REGISTER", wc2.URL, nil), // different country
	}
	store := &MockStore{
		GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
			return notifications, nil
		},
	}
	h := newMockHandler(store)
	h.CheckLifecycleNotifications(context.Background(), "NO", "REGISTER")

	if wc1.Calls != 1 {
		t.Errorf("wc1: expected 1 call, got %d", wc1.Calls)
	}
	if wc2.Calls != 1 { // only id-2 matched, not id-3
		t.Errorf("wc2: expected 1 call (id-2 only), got %d", wc2.Calls)
	}
}

func TestCheckLifecycleNotifications_StoreError(t *testing.T) {
	store := &MockStore{
		GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
			return nil, errors.New("db down")
		},
	}
	h := newMockHandler(store)
	// Should not panic even when store errors
	h.CheckLifecycleNotifications(context.Background(), "NO", "REGISTER")
}

func TestCheckLifecycleNotifications_EmptyStore(t *testing.T) {
	store := &MockStore{
		GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
			return []models.AllRegisteredWebhook{}, nil
		},
	}
	h := newMockHandler(store)
	// Should not panic on empty list
	h.CheckLifecycleNotifications(context.Background(), "NO", "REGISTER")
}

func TestCheckLifecycleNotifications_CountryIsCaseInsensitive(t *testing.T) {
	wc := newWebhookCapture(t)
	notifications := []models.AllRegisteredWebhook{
		makeAllRegisteredWebhook("id-1", "no", "REGISTER", wc.URL, nil),
	}
	store := &MockStore{
		GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
			return notifications, nil
		},
	}
	h := newMockHandler(store)
	h.CheckLifecycleNotifications(context.Background(), "NO", "REGISTER")

	if wc.Calls != 1 {
		t.Errorf("country match should be case-insensitive, expected 1 call, got %d", wc.Calls)
	}
}

func TestCheckLifecycleNotifications_WebhookFailureContinues(t *testing.T) {
	failing := failingWebhookServer(t)
	wc := newWebhookCapture(t)
	notifications := []models.AllRegisteredWebhook{
		makeAllRegisteredWebhook("id-1", "NO", "REGISTER", failing.URL, nil), // will fail
		makeAllRegisteredWebhook("id-2", "NO", "REGISTER", wc.URL, nil),      // should still fire
	}
	store := &MockStore{
		GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
			return notifications, nil
		},
	}
	h := newMockHandler(store)
	h.CheckLifecycleNotifications(context.Background(), "NO", "REGISTER")

	if wc.Calls != 1 {
		t.Errorf("second webhook should fire even if first fails, got %d calls", wc.Calls)
	}
}

// CheckThresholdNotifications

func thresholdNotification(field, op string, value float64) *models.ThresholdNotification {
	return &models.ThresholdNotification{Field: field, Operator: op, Value: value}
}

func TestCheckThresholdNotifications_ConditionMet_Greater(t *testing.T) {
	wc := newWebhookCapture(t)
	field := strings.ToUpper(utils.VALIDTHRESHOLDS[0])
	n := makeAllRegisteredWebhook("id-1", "NO", "THRESHOLD", wc.URL, thresholdNotification(field, ">", 20.0))
	store := &MockStore{
		GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
			return []models.AllRegisteredWebhook{n}, nil
		},
	}
	h := newMockHandler(store)
	h.CheckThresholdNotifications(context.Background(), "NO", map[string]float64{field: 25.0})

	if wc.Calls != 1 {
		t.Errorf("expected 1 call when condition met (>), got %d", wc.Calls)
	}
}

func TestCheckThresholdNotifications_ConditionNotMet_Greater(t *testing.T) {
	wc := newWebhookCapture(t)
	field := strings.ToUpper(utils.VALIDTHRESHOLDS[0])
	n := makeAllRegisteredWebhook("id-1", "NO", "THRESHOLD", wc.URL, thresholdNotification(field, ">", 20.0))
	store := &MockStore{
		GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
			return []models.AllRegisteredWebhook{n}, nil
		},
	}
	h := newMockHandler(store)
	h.CheckThresholdNotifications(context.Background(), "NO", map[string]float64{field: 10.0})

	if wc.Calls != 0 {
		t.Errorf("expected 0 calls when condition not met, got %d", wc.Calls)
	}
}

func TestCheckThresholdNotifications_AllOperators(t *testing.T) {
	field := strings.ToUpper(utils.VALIDTHRESHOLDS[0])

	tests := []struct {
		op        string
		threshold float64
		measured  float64
		wantCall  bool
	}{
		{">", 10.0, 15.0, true},
		{">", 10.0, 5.0, false},
		{"<", 10.0, 5.0, true},
		{"<", 10.0, 15.0, false},
		{">=", 10.0, 10.0, true},
		{">=", 10.0, 9.9, false},
		{"<=", 10.0, 10.0, true},
		{"<=", 10.0, 10.1, false},
		{"==", 10.0, 10.0, true},
		{"==", 10.0, 10.1, false},
	}

	for _, tc := range tests {
		wc := newWebhookCapture(t)
		n := makeAllRegisteredWebhook("id", "NO", "THRESHOLD", wc.URL, thresholdNotification(field, tc.op, tc.threshold))
		store := &MockStore{
			GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
				return []models.AllRegisteredWebhook{n}, nil
			},
		}
		h := newMockHandler(store)
		h.CheckThresholdNotifications(context.Background(), "NO", map[string]float64{field: tc.measured})

		called := wc.Calls > 0
		if called != tc.wantCall {
			t.Errorf("op=%q threshold=%.1f measured=%.1f: wantCall=%v got called=%v",
				tc.op, tc.threshold, tc.measured, tc.wantCall, called)
		}
	}
}

func TestCheckThresholdNotifications_CountryMismatch(t *testing.T) {
	wc := newWebhookCapture(t)
	field := strings.ToUpper(utils.VALIDTHRESHOLDS[0])
	n := makeAllRegisteredWebhook("id-1", "SE", "THRESHOLD", wc.URL, thresholdNotification(field, ">", 0.0))
	store := &MockStore{
		GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
			return []models.AllRegisteredWebhook{n}, nil
		},
	}
	h := newMockHandler(store)
	h.CheckThresholdNotifications(context.Background(), "NO", map[string]float64{field: 99.0})

	if wc.Calls != 0 {
		t.Errorf("expected 0 calls for country mismatch, got %d", wc.Calls)
	}
}

func TestCheckThresholdNotifications_WildcardCountry(t *testing.T) {
	wc := newWebhookCapture(t)
	field := strings.ToUpper(utils.VALIDTHRESHOLDS[0])
	n := makeAllRegisteredWebhook("id-1", "", "THRESHOLD", wc.URL, thresholdNotification(field, ">", 0.0))
	store := &MockStore{
		GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
			return []models.AllRegisteredWebhook{n}, nil
		},
	}
	h := newMockHandler(store)
	h.CheckThresholdNotifications(context.Background(), "NO", map[string]float64{field: 5.0})

	if wc.Calls != 1 {
		t.Errorf("expected 1 call for wildcard country, got %d", wc.Calls)
	}
}

func TestCheckThresholdNotifications_NilThreshold(t *testing.T) {
	wc := newWebhookCapture(t)
	n := makeAllRegisteredWebhook("id-1", "NO", "THRESHOLD", wc.URL, nil) // nil threshold
	store := &MockStore{
		GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
			return []models.AllRegisteredWebhook{n}, nil
		},
	}
	h := newMockHandler(store)
	field := strings.ToUpper(utils.VALIDTHRESHOLDS[0])
	// Should not panic
	h.CheckThresholdNotifications(context.Background(), "NO", map[string]float64{field: 5.0})

	if wc.Calls != 0 {
		t.Errorf("nil threshold should never fire, got %d calls", wc.Calls)
	}
}

func TestCheckThresholdNotifications_NonThresholdEventSkipped(t *testing.T) {
	wc := newWebhookCapture(t)
	field := strings.ToUpper(utils.VALIDTHRESHOLDS[0])
	th := thresholdNotification(field, ">", 0.0)
	n := makeAllRegisteredWebhook("id-1", "NO", "REGISTER", wc.URL, th) // wrong event
	store := &MockStore{
		GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
			return []models.AllRegisteredWebhook{n}, nil
		},
	}
	h := newMockHandler(store)
	h.CheckThresholdNotifications(context.Background(), "NO", map[string]float64{field: 99.0})

	if wc.Calls != 0 {
		t.Errorf("non-THRESHOLD event should be skipped, got %d calls", wc.Calls)
	}
}

func TestCheckThresholdNotifications_FieldNotInMeasured(t *testing.T) {
	wc := newWebhookCapture(t)
	field := strings.ToUpper(utils.VALIDTHRESHOLDS[0])
	n := makeAllRegisteredWebhook("id-1", "NO", "THRESHOLD", wc.URL, thresholdNotification(field, ">", 0.0))
	store := &MockStore{
		GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
			return []models.AllRegisteredWebhook{n}, nil
		},
	}
	h := newMockHandler(store)
	// provide a different field in measured map
	h.CheckThresholdNotifications(context.Background(), "NO", map[string]float64{"OTHER_FIELD": 99.0})

	if wc.Calls != 0 {
		t.Errorf("missing measured field should skip notification, got %d calls", wc.Calls)
	}
}

func TestCheckThresholdNotifications_StoreError(t *testing.T) {
	store := &MockStore{
		GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
			return nil, errors.New("db error")
		},
	}
	h := newMockHandler(store)
	field := strings.ToUpper(utils.VALIDTHRESHOLDS[0])
	// Should not panic
	h.CheckThresholdNotifications(context.Background(), "NO", map[string]float64{field: 10.0})
}

func TestCheckThresholdNotifications_WebhookFailureContinues(t *testing.T) {
	failing := failingWebhookServer(t)
	wc := newWebhookCapture(t)
	field := strings.ToUpper(utils.VALIDTHRESHOLDS[0])

	notifications := []models.AllRegisteredWebhook{
		makeAllRegisteredWebhook("id-1", "NO", "THRESHOLD", failing.URL, thresholdNotification(field, ">", 0.0)),
		makeAllRegisteredWebhook("id-2", "NO", "THRESHOLD", wc.URL, thresholdNotification(field, ">", 0.0)),
	}
	store := &MockStore{
		GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
			return notifications, nil
		},
	}
	h := newMockHandler(store)
	h.CheckThresholdNotifications(context.Background(), "NO", map[string]float64{field: 10.0})

	if wc.Calls != 1 {
		t.Errorf("second webhook should still fire after first fails, got %d calls", wc.Calls)
	}
}

func TestCheckThresholdNotifications_BoundaryEqual(t *testing.T) {
	wc1 := newWebhookCapture(t)
	wc2 := newWebhookCapture(t)
	field := strings.ToUpper(utils.VALIDTHRESHOLDS[0])

	strictGT := makeAllRegisteredWebhook("id-1", "NO", "THRESHOLD", wc1.URL, thresholdNotification(field, ">", 10.0))
	gte := makeAllRegisteredWebhook("id-2", "NO", "THRESHOLD", wc2.URL, thresholdNotification(field, ">=", 10.0))

	store := &MockStore{
		GetAllNotificationsFn: func(_ context.Context) ([]models.AllRegisteredWebhook, error) {
			return []models.AllRegisteredWebhook{strictGT, gte}, nil
		},
	}
	h := newMockHandler(store)
	h.CheckThresholdNotifications(context.Background(), "NO", map[string]float64{field: 10.0})

	if wc1.Calls != 0 {
		t.Error("strict > should NOT fire when measured == threshold")
	}
	if wc2.Calls != 1 {
		t.Error(">= SHOULD fire when measured == threshold")
	}
}
