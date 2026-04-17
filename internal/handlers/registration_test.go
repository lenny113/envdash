package handlers_test

import (
	"assignment-2/internal/handlers"
	model "assignment-2/internal/models"
	"assignment-2/internal/store"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

const regBase = "/envdash/v1/registrations/"

func newHandler() *handlers.Handler {
	return handlers.NewHandler(store.NewMockStore(), nil)
}

func authHeader() map[string]string {
	return map[string]string{"X-API-Key": "valid"}
}

func doRequest(
	t *testing.T,
	h *handlers.Handler,
	method, path, body string,
	headers map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()
	var bodyReader *bytes.Reader
	if body != "" {
		bodyReader = bytes.NewReader([]byte(body))
	} else {
		bodyReader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.RegistrationHandler(rr, req)
	return rr
}

func TestPost_ValidRegistration_Returns201(t *testing.T) {
	h := newHandler()
	body := `{
		"country": "Norway",
		"isoCode": "NO",
		"features": {
			"temperature": true,
			"precipitation": false,
			"capital": true,
			"coordinates": false,
			"population": false,
			"area": false,
			"targetCurrencies": ["NOK","EUR"]
		}
	}`
	rr := doRequest(t, h, http.MethodPost, regBase, body, authHeader())
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if resp["id"] == "" {
		t.Error("expected non-empty id in response")
	}
	if resp["lastChange"] == "" {
		t.Error("expected non-empty lastChange in response")
	}
}

func TestPost_MissingCountryAndIso_Returns400(t *testing.T) {
	h := newHandler()
	body := `{"features": {"temperature": true}}`
	rr := doRequest(t, h, http.MethodPost, regBase, body, authHeader())
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestPost_InvalidIsoCode_Returns400(t *testing.T) {
	h := newHandler()
	body := `{"isoCode": "ZZZ", "features": {"temperature": true}}`
	rr := doRequest(t, h, http.MethodPost, regBase, body, authHeader())
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestPost_CountryIsoCaseMismatch_Returns400(t *testing.T) {
	h := newHandler()
	body := `{"country": "Germany", "isoCode": "NO", "features": {"temperature": true}}`
	rr := doRequest(t, h, http.MethodPost, regBase, body, authHeader())
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestPost_InvalidJSON_Returns400(t *testing.T) {
	h := newHandler()
	rr := doRequest(t, h, http.MethodPost, regBase, `{not json}`, authHeader())
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestPost_Unauthorized_Returns401(t *testing.T) {
	h := newHandler()
	body := `{"country": "Norway", "isoCode": "NO", "features": {"temperature": true}}`
	rr := doRequest(t, h, http.MethodPost, regBase, body, map[string]string{"X-API-Key": "wrong"})
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
}

func TestPost_TooManyCurrencies_Returns400(t *testing.T) {
	h := newHandler()
	currencies := `["NOK","EUR","USD","GBP","JPY","CHF","CAD","AUD","SEK","DKK","NZD"]`
	body := `{"isoCode": "NO", "features": {"temperature": true, "targetCurrencies": ` + currencies + `}}`
	rr := doRequest(t, h, http.MethodPost, regBase, body, authHeader())
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestPost_InvalidCurrencyCode_Returns400(t *testing.T) {
	h := newHandler()
	body := `{"isoCode": "NO", "features": {"targetCurrencies": ["ZZZ"]}}`
	rr := doRequest(t, h, http.MethodPost, regBase, body, authHeader())
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func seedRegistration(t *testing.T, h *handlers.Handler) string {
	t.Helper()
	body := `{"isoCode": "NO", "features": {"temperature": true}}`
	rr := doRequest(t, h, http.MethodPost, regBase, body, authHeader())
	if rr.Code != http.StatusCreated {
		t.Fatalf("seed: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	json.Unmarshal(rr.Body.Bytes(), &resp)
	return resp["id"]
}

func TestGet_ExistingRegistration_Returns200(t *testing.T) {
	h := newHandler()
	id := seedRegistration(t, h)
	rr := doRequest(t, h, http.MethodGet, regBase+id, "", authHeader())
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var reg model.Registration
	if err := json.Unmarshal(rr.Body.Bytes(), &reg); err != nil {
		t.Fatalf("response is not a valid Registration: %v", err)
	}
}

func TestGet_AllRegistrations_ReturnsSlice(t *testing.T) {
	h := newHandler()
	seedRegistration(t, h)
	rr := doRequest(t, h, http.MethodGet, regBase, "", authHeader())
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var regs []model.Registration
	if err := json.Unmarshal(rr.Body.Bytes(), &regs); err != nil {
		t.Fatalf("response is not a valid Registration slice: %v", err)
	}
	if len(regs) == 0 {
		t.Error("expected at least one registration")
	}
}

func TestGet_Unauthorized_Returns401(t *testing.T) {
	h := newHandler()
	rr := doRequest(t, h, http.MethodGet, regBase+"some-id", "", map[string]string{"X-API-Key": "bad"})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestPut_ValidUpdate_Returns200(t *testing.T) {
	h := newHandler()
	id := seedRegistration(t, h)
	body := `{
		"country": "Sweden",
		"isoCode": "SE",
		"features": {"temperature": false, "precipitation": true}
	}`
	rr := doRequest(t, h, http.MethodPut, regBase+id, body, authHeader())
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var reg model.Registration
	json.Unmarshal(rr.Body.Bytes(), &reg)
	if reg.IsoCode != "SE" {
		t.Errorf("expected isoCode SE, got %s", reg.IsoCode)
	}
}

func TestPut_MissingID_Returns400(t *testing.T) {
	h := newHandler()
	body := `{"isoCode": "SE", "features": {"temperature": true}}`
	rr := doRequest(t, h, http.MethodPut, regBase, body, authHeader())
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestPut_NonExistentID_Returns404(t *testing.T) {
	h := newHandler()
	body := `{"isoCode": "SE", "features": {"temperature": true}}`
	rr := doRequest(t, h, http.MethodPut, regBase+"does-not-exist", body, authHeader())
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestPut_Unauthorized_Returns401(t *testing.T) {
	h := newHandler()
	rr := doRequest(t, h, http.MethodPut, regBase+"some-id", `{}`, map[string]string{"X-API-Key": "bad"})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestPut_InvalidBody_Returns400(t *testing.T) {
	h := newHandler()
	id := seedRegistration(t, h)
	rr := doRequest(t, h, http.MethodPut, regBase+id, `{bad}`, authHeader())
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestDelete_ExistingRegistration_Returns204(t *testing.T) {
	h := newHandler()
	id := seedRegistration(t, h)
	rr := doRequest(t, h, http.MethodDelete, regBase+id, "", authHeader())
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDelete_NonExistentID_Returns404(t *testing.T) {
	h := newHandler()
	rr := doRequest(t, h, http.MethodDelete, regBase+"ghost-id", "", authHeader())
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestDelete_MissingID_Returns400(t *testing.T) {
	h := newHandler()
	rr := doRequest(t, h, http.MethodDelete, regBase, "", authHeader())
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestDelete_Unauthorized_Returns401(t *testing.T) {
	h := newHandler()
	rr := doRequest(t, h, http.MethodDelete, regBase+"some-id", "", map[string]string{"X-API-Key": "bad"})
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestPatch_ValidPatch_Returns204(t *testing.T) {
	h := newHandler()
	id := seedRegistration(t, h)
	body := `{"features": {"temperature": false}}`
	rr := doRequest(t, h, http.MethodPatch, regBase+id, body, authHeader())
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestPatch_UpdateCountry_Returns204(t *testing.T) {
	h := newHandler()
	id := seedRegistration(t, h)
	body := `{"country": "Sweden", "isoCode": "SE"}`
	rr := doRequest(t, h, http.MethodPatch, regBase+id, body, authHeader())
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestPatch_MissingID_Returns400(t *testing.T) {
	h := newHandler()
	rr := doRequest(t, h, http.MethodPatch, regBase, `{"features": {}}`, authHeader())
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestPatch_Unauthorized_Returns401(t *testing.T) {
	h := newHandler()
	id := seedRegistration(t, h)
	rr := doRequest(t, h, http.MethodPatch, regBase+id, `{}`, map[string]string{"X-API-Key": "bad"})
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

func TestPatch_InvalidJSON_Returns400(t *testing.T) {
	h := newHandler()
	id := seedRegistration(t, h)
	rr := doRequest(t, h, http.MethodPatch, regBase+id, `{bad}`, authHeader())
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestPatch_InvalidCurrencyInPatch_Returns400(t *testing.T) {
	h := newHandler()
	id := seedRegistration(t, h)
	body := `{"features": {"targetCurrencies": ["ZZZ"]}}`
	rr := doRequest(t, h, http.MethodPatch, regBase+id, body, authHeader())
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHead_ExistingRegistration_Returns200(t *testing.T) {
	h := newHandler()
	id := seedRegistration(t, h)
	rr := doRequest(t, h, http.MethodHead, regBase+id, "", authHeader())
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Error("HEAD response must have no body")
	}
}

func TestHead_AllRegistrations_Returns200(t *testing.T) {
	h := newHandler()
	seedRegistration(t, h)
	rr := doRequest(t, h, http.MethodHead, regBase, "", authHeader())
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHead_NonExistentID_Returns404(t *testing.T) {
	h := newHandler()
	rr := doRequest(t, h, http.MethodHead, regBase+"ghost", "", authHeader())
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHead_Unauthorized_Returns401(t *testing.T) {
	h := newHandler()
	rr := doRequest(t, h, http.MethodHead, regBase, "", map[string]string{"X-API-Key": "bad"})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestOptions_Returns200WithAllowHeader(t *testing.T) {
	h := newHandler()
	rr := doRequest(t, h, http.MethodOptions, regBase, "", authHeader())
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	allow := rr.Header().Get("Allow")
	if allow == "" {
		t.Error("expected Allow header to be set")
	}
}

func TestUnsupportedMethod_Returns405(t *testing.T) {
	h := newHandler()
	rr := doRequest(t, h, http.MethodTrace, regBase, "", authHeader())
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestPost_LowercaseIsoAndCountry_NormalisedAndAccepted(t *testing.T) {
	h := newHandler()
	body := `{"country": "norway", "isoCode": "no", "features": {"temperature": true}}`
	rr := doRequest(t, h, http.MethodPost, regBase, body, authHeader())
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 after normalisation, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRoundTrip_PostThenGet_DataConsistent(t *testing.T) {
	h := newHandler()
	postBody := `{"isoCode": "DE", "features": {"temperature": true, "targetCurrencies": ["EUR"]}}`
	postRR := doRequest(t, h, http.MethodPost, regBase, postBody, authHeader())
	if postRR.Code != http.StatusCreated {
		t.Fatalf("POST failed: %d %s", postRR.Code, postRR.Body.String())
	}
	var created map[string]string
	json.Unmarshal(postRR.Body.Bytes(), &created)

	getRR := doRequest(t, h, http.MethodGet, regBase+created["id"], "", authHeader())
	if getRR.Code != http.StatusOK {
		t.Fatalf("GET failed: %d %s", getRR.Code, getRR.Body.String())
	}
	var reg model.Registration
	if err := json.Unmarshal(getRR.Body.Bytes(), &reg); err != nil {
		t.Fatalf("GET body is not valid JSON: %v", err)
	}
	if reg.IsoCode != "DE" {
		t.Errorf("expected DE, got %s", reg.IsoCode)
	}
}

func TestRoundTrip_PostDeleteGet_NotFound(t *testing.T) {
	h := newHandler()
	id := seedRegistration(t, h)

	delRR := doRequest(t, h, http.MethodDelete, regBase+id, "", authHeader())
	if delRR.Code != http.StatusNoContent {
		t.Fatalf("DELETE failed: %d", delRR.Code)
	}

	getRR := doRequest(t, h, http.MethodGet, regBase+id, "", authHeader())
	var reg model.Registration
	json.Unmarshal(getRR.Body.Bytes(), &reg)
	if reg.IsoCode != "" {
		t.Error("expected empty registration after deletion")
	}
}
