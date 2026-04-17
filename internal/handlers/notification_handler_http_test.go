package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"assignment-2/internal/models"
	"assignment-2/internal/utils"

	"cloud.google.com/go/firestore"
)

// helpers
//

func notificationHandlerWithStore(store *MockStore) *Handler {
	return newMockHandler(store)
}

func makeNotificationRequest(method, target, body, apiKey string) *http.Request {
	var buf *bytes.Buffer
	if body != "" {
		buf = bytes.NewBufferString(body)
	} else {
		buf = &bytes.Buffer{}
	}
	req := httptest.NewRequest(method, target, buf)
	if apiKey != "" {
		req.Header.Set("X-Api-Key", apiKey)
		req.Header.Set("X-API-Key", apiKey) // handler reads both casings
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

// NotificationSpinner – method routing

func TestNotificationSpinner_POST_routed(t *testing.T) {
	store := &MockStore{
		CreateNotificationFn: func(_ context.Context, n models.RegisterWebhook, _ string) (string, error) {
			return "new-id-123", nil
		},
	}
	h := notificationHandlerWithStore(store)

	body := buildRegisterBody(utils.VALIDEVENTS[0], "https://hook.example.com", "NO", nil)
	req := makeNotificationRequest(http.MethodPost, "/notifications", body, "test-api-key")
	rr := httptest.NewRecorder()
	h.NotificationSpinner(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}
}

func TestNotificationSpinner_GET_routed(t *testing.T) {
	store := &MockStore{
		GetAllNotificationsForUserFn: func(_ context.Context, _ string) ([]models.AllRegisteredWebhook, error) {
			return []models.AllRegisteredWebhook{}, nil
		},
	}
	h := notificationHandlerWithStore(store)
	req := makeNotificationRequest(http.MethodGet, "/notifications", "", "key")
	rr := httptest.NewRecorder()
	h.NotificationSpinner(rr, req)

	if rr.Code == http.StatusMethodNotAllowed {
		t.Error("GET should be routed, not rejected")
	}
}

func TestNotificationSpinner_MethodNotAllowed(t *testing.T) {
	h := notificationHandlerWithStore(&MockStore{})
	for _, method := range []string{http.MethodPut, http.MethodPatch, http.MethodDelete} {
		req := makeNotificationRequest(method, "/notifications", "", "")
		rr := httptest.NewRecorder()
		h.NotificationSpinner(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: expected 405, got %d", method, rr.Code)
		}
	}
}

func TestNotificationSpinnerById_GET_routed(t *testing.T) {
	store := &MockStore{
		GetSpecificNotificationFn: func(_ context.Context, id string) (models.AllRegisteredWebhook, *firestore.DocumentRef, error) {
			return models.AllRegisteredWebhook{Id: id}, nil, nil
		},
	}
	h := notificationHandlerWithStore(store)
	req := makeNotificationRequest(http.MethodGet, "/notifications/abc", "", "key")
	req.SetPathValue("id", "abc")
	rr := httptest.NewRecorder()
	h.NotificationSpinnerById(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestNotificationSpinnerById_DELETE_routed(t *testing.T) {
	store := &MockStore{
		DeleteNotificationFn: func(_ context.Context, _ string, _ string) error {
			return nil
		},
	}
	h := notificationHandlerWithStore(store)
	req := makeNotificationRequest(http.MethodDelete, "/notifications/abc", "", "key")
	req.SetPathValue("id", "abc")
	rr := httptest.NewRecorder()
	h.NotificationSpinnerById(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestNotificationSpinnerById_MethodNotAllowed(t *testing.T) {
	h := notificationHandlerWithStore(&MockStore{})
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch} {
		req := makeNotificationRequest(method, "/notifications/abc", "", "")
		rr := httptest.NewRecorder()
		h.NotificationSpinnerById(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: expected 405, got %d", method, rr.Code)
		}
	}
}

//
// registerNewNotification

func TestRegisterNewNotification_Success(t *testing.T) {
	wantID := "generated-id-42"
	store := &MockStore{
		CreateNotificationFn: func(_ context.Context, _ models.RegisterWebhook, _ string) (string, error) {
			return wantID, nil
		},
	}
	h := notificationHandlerWithStore(store)

	body := buildRegisterBody(utils.VALIDEVENTS[0], "https://hook.example.com", "", nil)
	req := makeNotificationRequest(http.MethodPost, "/notifications", body, "my-key")
	rr := httptest.NewRecorder()
	h.registerNewNotification(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d – body: %s", rr.Code, rr.Body.String())
	}

	var resp models.RegisteredWebhookResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("could not unmarshal response: %v", err)
	}
	if resp.Id != wantID {
		t.Errorf("expected id %q, got %q", wantID, resp.Id)
	}
}

func TestRegisterNewNotification_MissingBody(t *testing.T) {
	h := notificationHandlerWithStore(&MockStore{})
	req := httptest.NewRequest(http.MethodPost, "/notifications", nil)
	rr := httptest.NewRecorder()
	h.registerNewNotification(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for nil body, got %d", rr.Code)
	}
}

func TestRegisterNewNotification_InvalidJSON(t *testing.T) {
	h := notificationHandlerWithStore(&MockStore{})
	req := makeNotificationRequest(http.MethodPost, "/notifications", `{not valid json`, "key")
	rr := httptest.NewRecorder()
	h.registerNewNotification(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rr.Code)
	}
}

func TestRegisterNewNotification_MissingURL(t *testing.T) {
	h := notificationHandlerWithStore(&MockStore{})
	body := `{"event":"` + utils.VALIDEVENTS[0] + `"}`
	req := makeNotificationRequest(http.MethodPost, "/notifications", body, "key")
	rr := httptest.NewRecorder()
	h.registerNewNotification(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing URL, got %d", rr.Code)
	}
}

func TestRegisterNewNotification_MissingEvent(t *testing.T) {
	h := notificationHandlerWithStore(&MockStore{})
	body := `{"url":"https://example.com/hook"}`
	req := makeNotificationRequest(http.MethodPost, "/notifications", body, "key")
	rr := httptest.NewRecorder()
	h.registerNewNotification(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing event, got %d", rr.Code)
	}
}

func TestRegisterNewNotification_InvalidEvent(t *testing.T) {
	h := notificationHandlerWithStore(&MockStore{})
	body := buildRegisterBody("UNKNOWN_EVENT", "https://hook.example.com", "", nil)
	req := makeNotificationRequest(http.MethodPost, "/notifications", body, "key")
	rr := httptest.NewRecorder()
	h.registerNewNotification(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unknown event, got %d", rr.Code)
	}
}

func TestRegisterNewNotification_ThresholdEventMissingThreshold(t *testing.T) {
	h := notificationHandlerWithStore(&MockStore{})
	body := buildRegisterBody("THRESHOLD", "https://hook.example.com", "", nil)
	req := makeNotificationRequest(http.MethodPost, "/notifications", body, "key")
	rr := httptest.NewRecorder()
	h.registerNewNotification(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when threshold body is absent for THRESHOLD event, got %d", rr.Code)
	}
}

func TestRegisterNewNotification_NonThresholdEventWithThresholdBody(t *testing.T) {
	h := notificationHandlerWithStore(&MockStore{})
	th := &models.ThresholdNotification{
		Field:    utils.VALIDTHRESHOLDS[0],
		Operator: utils.VALIDOPERATORS[0],
		Value:    5.0,
	}
	body := buildRegisterBody(utils.VALIDEVENTS[0], "https://hook.example.com", "", th)
	req := makeNotificationRequest(http.MethodPost, "/notifications", body, "key")
	rr := httptest.NewRecorder()
	h.registerNewNotification(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when threshold body given for non-THRESHOLD event, got %d", rr.Code)
	}
}

func TestRegisterNewNotification_ThresholdEventInvalidThreshold(t *testing.T) {
	h := notificationHandlerWithStore(&MockStore{})
	th := &models.ThresholdNotification{
		Field:    "INVALID_FIELD",
		Operator: "???",
		Value:    0,
	}
	body := buildRegisterBody("THRESHOLD", "https://hook.example.com", "", th)
	req := makeNotificationRequest(http.MethodPost, "/notifications", body, "key")
	rr := httptest.NewRecorder()
	h.registerNewNotification(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid threshold body, got %d", rr.Code)
	}
}

func TestRegisterNewNotification_ThresholdEventSuccess(t *testing.T) {
	store := &MockStore{
		CreateNotificationFn: func(_ context.Context, _ models.RegisterWebhook, _ string) (string, error) {
			return "thresh-id", nil
		},
	}
	h := notificationHandlerWithStore(store)
	th := &models.ThresholdNotification{
		Field:    utils.VALIDTHRESHOLDS[0],
		Operator: utils.VALIDOPERATORS[0],
		Value:    25.0,
	}
	body := buildRegisterBody("THRESHOLD", "https://hook.example.com", "NO", th)
	req := makeNotificationRequest(http.MethodPost, "/notifications", body, "key")
	rr := httptest.NewRecorder()
	h.registerNewNotification(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 for valid threshold registration, got %d – %s", rr.Code, rr.Body.String())
	}
}

func TestRegisterNewNotification_StoreError(t *testing.T) {
	store := &MockStore{
		CreateNotificationFn: func(_ context.Context, _ models.RegisterWebhook, _ string) (string, error) {
			return "", errors.New("firestore down")
		},
	}
	h := notificationHandlerWithStore(store)
	body := buildRegisterBody(utils.VALIDEVENTS[0], "https://hook.example.com", "", nil)
	req := makeNotificationRequest(http.MethodPost, "/notifications", body, "key")
	rr := httptest.NewRecorder()
	h.registerNewNotification(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on store error, got %d", rr.Code)
	}
}

func TestRegisterNewNotification_CaseInsensitiveEvent(t *testing.T) {
	store := &MockStore{
		CreateNotificationFn: func(_ context.Context, n models.RegisterWebhook, _ string) (string, error) {
			return "id", nil
		},
	}
	h := notificationHandlerWithStore(store)
	// send lowercase event – handler uppercases it before validation
	body := `{"url":"https://hook.example.com","event":"` + strings.ToLower(utils.VALIDEVENTS[0]) + `"}`
	req := makeNotificationRequest(http.MethodPost, "/notifications", body, "key")
	rr := httptest.NewRecorder()
	h.registerNewNotification(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 for lowercase event, got %d – %s", rr.Code, rr.Body.String())
	}
}

//
// allNotifications
//

func TestAllNotifications_ReturnsEmpty(t *testing.T) {
	store := &MockStore{
		GetAllNotificationsForUserFn: func(_ context.Context, _ string) ([]models.AllRegisteredWebhook, error) {
			return nil, nil
		},
	}
	h := notificationHandlerWithStore(store)
	req := makeNotificationRequest(http.MethodGet, "/notifications", "", "key")
	rr := httptest.NewRecorder()
	h.allNotifications(rr, req)

	// nil result → empty list branch; should not crash and not return 5xx
	if rr.Code >= 500 {
		t.Errorf("unexpected 5xx: %d", rr.Code)
	}
}

func TestAllNotifications_ReturnsList(t *testing.T) {
	items := []models.AllRegisteredWebhook{
		{Id: "id-1", RegisterWebhook: models.RegisterWebhook{Url: "https://a.com", Event: utils.VALIDEVENTS[0]}},
		{Id: "id-2", RegisterWebhook: models.RegisterWebhook{Url: "https://b.com", Event: utils.VALIDEVENTS[0]}},
	}
	store := &MockStore{
		GetAllNotificationsForUserFn: func(_ context.Context, _ string) ([]models.AllRegisteredWebhook, error) {
			return items, nil
		},
	}
	h := notificationHandlerWithStore(store)
	req := makeNotificationRequest(http.MethodGet, "/notifications", "", "key")
	rr := httptest.NewRecorder()
	h.allNotifications(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var result []models.AllRegisteredWebhook
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("could not parse response: %v", err)
	}
	if len(result) != len(items) {
		t.Errorf("expected %d items, got %d", len(items), len(result))
	}
}

func TestAllNotifications_StoreError(t *testing.T) {
	store := &MockStore{
		GetAllNotificationsForUserFn: func(_ context.Context, _ string) ([]models.AllRegisteredWebhook, error) {
			return nil, errors.New("db error")
		},
	}
	h := notificationHandlerWithStore(store)
	req := makeNotificationRequest(http.MethodGet, "/notifications", "", "key")
	rr := httptest.NewRecorder()
	h.allNotifications(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on store error, got %d", rr.Code)
	}
}

func TestAllNotifications_ContentTypeJSON(t *testing.T) {
	items := []models.AllRegisteredWebhook{
		{Id: "x", RegisterWebhook: models.RegisterWebhook{Url: "https://x.com", Event: utils.VALIDEVENTS[0]}},
	}
	store := &MockStore{
		GetAllNotificationsForUserFn: func(_ context.Context, _ string) ([]models.AllRegisteredWebhook, error) {
			return items, nil
		},
	}
	h := notificationHandlerWithStore(store)
	req := makeNotificationRequest(http.MethodGet, "/notifications", "", "key")
	rr := httptest.NewRecorder()
	h.allNotifications(rr, req)

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}

//
// specificNotification
//

func TestSpecificNotification_Found(t *testing.T) {
	want := models.AllRegisteredWebhook{
		Id:              "abc",
		RegisterWebhook: models.RegisterWebhook{Url: "https://hook.io", Event: utils.VALIDEVENTS[0]},
	}
	store := &MockStore{
		GetSpecificNotificationFn: func(_ context.Context, id string) (models.AllRegisteredWebhook, *firestore.DocumentRef, error) {
			return want, nil, nil
		},
	}
	h := notificationHandlerWithStore(store)
	req := makeNotificationRequest(http.MethodGet, "/notifications/abc", "", "key")
	req.SetPathValue("id", "abc")
	rr := httptest.NewRecorder()
	h.specificNotification(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var got models.AllRegisteredWebhook
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if got.Id != want.Id {
		t.Errorf("expected id %q, got %q", want.Id, got.Id)
	}
}

func TestSpecificNotification_NotFound(t *testing.T) {
	store := &MockStore{
		GetSpecificNotificationFn: func(_ context.Context, _ string) (models.AllRegisteredWebhook, *firestore.DocumentRef, error) {
			return models.AllRegisteredWebhook{}, nil, errors.New("not found")
		},
	}
	h := notificationHandlerWithStore(store)
	req := makeNotificationRequest(http.MethodGet, "/notifications/missing", "", "key")
	req.SetPathValue("id", "missing")
	rr := httptest.NewRecorder()
	h.specificNotification(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestSpecificNotification_MissingID(t *testing.T) {
	h := notificationHandlerWithStore(&MockStore{})
	req := makeNotificationRequest(http.MethodGet, "/notifications/", "", "key")
	// do NOT call SetPathValue so id remains ""
	rr := httptest.NewRecorder()
	h.specificNotification(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing id, got %d", rr.Code)
	}
}

//
// deleteNotification
//

func TestDeleteNotification_Success(t *testing.T) {
	store := &MockStore{
		DeleteNotificationFn: func(_ context.Context, id string, _ string) error {
			return nil
		},
	}
	h := notificationHandlerWithStore(store)
	req := makeNotificationRequest(http.MethodDelete, "/notifications/abc", "", "key")
	req.SetPathValue("id", "abc")
	rr := httptest.NewRecorder()
	h.deleteNotification(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestDeleteNotification_NotFound(t *testing.T) {
	store := &MockStore{
		DeleteNotificationFn: func(_ context.Context, _ string, _ string) error {
			return errors.New("does not exist")
		},
	}
	h := notificationHandlerWithStore(store)
	req := makeNotificationRequest(http.MethodDelete, "/notifications/ghost", "", "key")
	req.SetPathValue("id", "ghost")
	rr := httptest.NewRecorder()
	h.deleteNotification(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestDeleteNotification_NoAccess(t *testing.T) {
	store := &MockStore{
		DeleteNotificationFn: func(_ context.Context, _ string, _ string) error {
			return errors.New("No access")
		},
	}
	h := notificationHandlerWithStore(store)
	req := makeNotificationRequest(http.MethodDelete, "/notifications/forbidden", "", "other-key")
	req.SetPathValue("id", "forbidden")
	rr := httptest.NewRecorder()
	h.deleteNotification(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestDeleteNotification_StoreError(t *testing.T) {
	store := &MockStore{
		DeleteNotificationFn: func(_ context.Context, _ string, _ string) error {
			return errors.New("unexpected firestore error")
		},
	}
	h := notificationHandlerWithStore(store)
	req := makeNotificationRequest(http.MethodDelete, "/notifications/xyz", "", "key")
	req.SetPathValue("id", "xyz")
	rr := httptest.NewRecorder()
	h.deleteNotification(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on unexpected store error, got %d", rr.Code)
	}
}

func TestDeleteNotification_MissingID(t *testing.T) {
	h := notificationHandlerWithStore(&MockStore{})
	req := makeNotificationRequest(http.MethodDelete, "/notifications/", "", "key")
	rr := httptest.NewRecorder()
	h.deleteNotification(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing id, got %d", rr.Code)
	}
}

// helper: build JSON body for registration
//

func buildRegisterBody(event, url, country string, th *models.ThresholdNotification) string {
	type bodyType struct {
		Url       string                        `json:"url"`
		Country   string                        `json:"country,omitempty"`
		Event     string                        `json:"event"`
		Threshold *models.ThresholdNotification `json:"threshold,omitempty"`
	}
	b := bodyType{Url: url, Country: country, Event: event, Threshold: th}
	data, _ := json.Marshal(b)
	return string(data)
}
