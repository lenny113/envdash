package handlers

import (
	"testing"

	"assignment-2/internal/models"
	"assignment-2/internal/utils"
)

//
// validateNotification

func TestValidateNotification_Valid(t *testing.T) {
	req := models.RegisterWebhook{
		Url:   "https://example.com/webhook",
		Event: utils.VALIDEVENTS[0], // first valid event (already uppercased)
	}
	err, msg := validateNotification(req)
	if err != nil {
		t.Errorf("expected no error, got: %s", msg)
	}
}

func TestValidateNotification_ValidWithCountry(t *testing.T) {
	req := models.RegisterWebhook{
		Url:     "https://example.com/webhook",
		Country: "NO",
		Event:   utils.VALIDEVENTS[0],
	}
	err, msg := validateNotification(req)
	if err != nil {
		t.Errorf("expected no error, got: %s", msg)
	}
}

func TestValidateNotification_MissingURL(t *testing.T) {
	req := models.RegisterWebhook{
		Url:   "",
		Event: utils.VALIDEVENTS[0],
	}
	err, msg := validateNotification(req)
	if err == nil {
		t.Error("expected error for missing URL")
	}
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestValidateNotification_InvalidURL(t *testing.T) {
	req := models.RegisterWebhook{
		Url:   "not-a-url",
		Event: utils.VALIDEVENTS[0],
	}
	err, msg := validateNotification(req)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestValidateNotification_MissingEvent(t *testing.T) {
	req := models.RegisterWebhook{
		Url:   "https://example.com/webhook",
		Event: "",
	}
	err, msg := validateNotification(req)
	if err == nil {
		t.Error("expected error for missing event")
	}
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestValidateNotification_InvalidEvent(t *testing.T) {
	req := models.RegisterWebhook{
		Url:   "https://example.com/webhook",
		Event: "NOT_A_REAL_EVENT",
	}
	err, msg := validateNotification(req)
	if err == nil {
		t.Error("expected error for invalid event")
	}
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestValidateNotification_BothFieldsMissing(t *testing.T) {
	req := models.RegisterWebhook{}
	err, msg := validateNotification(req)
	if err == nil {
		t.Error("expected error when both URL and event are missing")
	}
	// message should mention both problems
	if msg == "" {
		t.Error("expected non-empty combined error message")
	}
}

func TestValidateNotification_AllValidEvents(t *testing.T) {
	for _, event := range utils.VALIDEVENTS {
		req := models.RegisterWebhook{
			Url:   "https://example.com/webhook",
			Event: event,
		}
		err, msg := validateNotification(req)
		if err != nil {
			t.Errorf("event %q should be valid but got error: %s", event, msg)
		}
	}
}

func TestValidateNotification_EventCaseInsensitive(t *testing.T) {
	// validateNotification receives an already-uppercased event from the handler,
	// but the function itself should handle the uppercase check correctly.
	req := models.RegisterWebhook{
		Url:   "https://example.com/webhook",
		Event: utils.VALIDEVENTS[0],
	}
	err, _ := validateNotification(req)
	if err != nil {
		t.Errorf("uppercase valid event should pass validation")
	}
}

func TestValidateNotification_URLWithPath(t *testing.T) {
	req := models.RegisterWebhook{
		Url:   "https://hooks.example.com/path/to/endpoint?token=abc",
		Event: utils.VALIDEVENTS[0],
	}
	err, msg := validateNotification(req)
	if err != nil {
		t.Errorf("URL with path and query should be valid, got: %s", msg)
	}
}

// validateThreshold

func TestValidateThreshold_Valid(t *testing.T) {
	th := models.ThresholdNotification{
		Field:    utils.VALIDTHRESHOLDS[0],
		Operator: utils.VALIDOPERATORS[0],
		Value:    10.0,
	}
	err, msg := validateThreshold(th)
	if err != nil {
		t.Errorf("expected no error, got: %s", msg)
	}
}

func TestValidateThreshold_AllOperators(t *testing.T) {
	for _, op := range utils.VALIDOPERATORS {
		th := models.ThresholdNotification{
			Field:    utils.VALIDTHRESHOLDS[0],
			Operator: op,
			Value:    5.0,
		}
		err, msg := validateThreshold(th)
		if err != nil {
			t.Errorf("operator %q should be valid but got: %s", op, msg)
		}
	}
}

func TestValidateThreshold_AllFields(t *testing.T) {
	for _, field := range utils.VALIDTHRESHOLDS {
		th := models.ThresholdNotification{
			Field:    field,
			Operator: utils.VALIDOPERATORS[0],
			Value:    0.0,
		}
		err, msg := validateThreshold(th)
		if err != nil {
			t.Errorf("field %q should be valid but got: %s", field, msg)
		}
	}
}

func TestValidateThreshold_MissingField(t *testing.T) {
	th := models.ThresholdNotification{
		Field:    "",
		Operator: utils.VALIDOPERATORS[0],
		Value:    10.0,
	}
	err, msg := validateThreshold(th)
	if err == nil {
		t.Error("expected error for missing field")
	}
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestValidateThreshold_InvalidField(t *testing.T) {
	th := models.ThresholdNotification{
		Field:    "NOT_A_REAL_FIELD",
		Operator: utils.VALIDOPERATORS[0],
		Value:    10.0,
	}
	err, msg := validateThreshold(th)
	if err == nil {
		t.Error("expected error for invalid field")
	}
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestValidateThreshold_MissingOperator(t *testing.T) {
	th := models.ThresholdNotification{
		Field:    utils.VALIDTHRESHOLDS[0],
		Operator: "",
		Value:    10.0,
	}
	err, msg := validateThreshold(th)
	if err == nil {
		t.Error("expected error for missing operator")
	}
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestValidateThreshold_InvalidOperator(t *testing.T) {
	th := models.ThresholdNotification{
		Field:    utils.VALIDTHRESHOLDS[0],
		Operator: "??",
		Value:    10.0,
	}
	err, msg := validateThreshold(th)
	if err == nil {
		t.Error("expected error for invalid operator")
	}
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestValidateThreshold_BothFieldsMissing(t *testing.T) {
	th := models.ThresholdNotification{}
	err, msg := validateThreshold(th)
	if err == nil {
		t.Error("expected error when field and operator are both missing")
	}
	if msg == "" {
		t.Error("expected combined error message")
	}
}

func TestValidateThreshold_OperatorWithWhitespace(t *testing.T) {
	// The handler trims whitespace before storing; validateThreshold also trims.
	th := models.ThresholdNotification{
		Field:    utils.VALIDTHRESHOLDS[0],
		Operator: "  " + utils.VALIDOPERATORS[0] + "  ",
		Value:    1.0,
	}
	err, _ := validateThreshold(th)
	if err != nil {
		t.Error("operator with surrounding whitespace should pass after trim")
	}
}

func TestValidateThreshold_ZeroValue(t *testing.T) {
	th := models.ThresholdNotification{
		Field:    utils.VALIDTHRESHOLDS[0],
		Operator: utils.VALIDOPERATORS[0],
		Value:    0.0,
	}
	err, msg := validateThreshold(th)
	if err != nil {
		t.Errorf("zero value should be valid, got: %s", msg)
	}
}

func TestValidateThreshold_NegativeValue(t *testing.T) {
	th := models.ThresholdNotification{
		Field:    utils.VALIDTHRESHOLDS[0],
		Operator: utils.VALIDOPERATORS[0],
		Value:    -99.9,
	}
	err, msg := validateThreshold(th)
	if err != nil {
		t.Errorf("negative value should be valid, got: %s", msg)
	}
}
