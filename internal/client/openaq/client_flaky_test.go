//go:build flaky
// +build flaky

package client

import (
	"net/http"
	"os"
	"testing"
	"time"
)

func TestOpenAQClient_GetInfo_PM25AndPM10(t *testing.T) {
	if os.Getenv("OPENAQ_API_KEY") == "" {
		t.Skip("skipping flaky test: OPENAQ_API_KEY is not set")
	}

	httpClient := &http.Client{
		Timeout: 20 * time.Second,
	}

	openAQClient := NewOpenAQClient(httpClient, os.Getenv("OPENAQ_API_KEY"))

	data, err := openAQClient.GetInfo(OpenAQ_InformationRequest{
		ISOCode: "NO",
		PM25:    true,
		PM10:    true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if data.MeanPM25 == nil {
		t.Fatal("expected MeanPM25 in response")
	}
	if *data.MeanPM25 <= 0 {
		t.Fatalf("expected MeanPM25 to be > 0, got %v", *data.MeanPM25)
	}

	if data.MeanPM10 == nil {
		t.Fatal("expected MeanPM10 in response")
	}
	if *data.MeanPM10 <= 0 {
		t.Fatalf("expected MeanPM10 to be > 0, got %v", *data.MeanPM10)
	}

	t.Logf("mean PM2.5: %v", *data.MeanPM25)
	t.Logf("mean PM10: %v", *data.MeanPM10)
}

func TestOpenAQClient_GetInfo_WithTrimmedLowercaseISOCode(t *testing.T) {
	if os.Getenv("OPENAQ_API_KEY") == "" {
		t.Skip("skipping flaky test: OPENAQ_API_KEY is not set")
	}

	httpClient := &http.Client{
		Timeout: 20 * time.Second,
	}

	openAQClient := NewOpenAQClient(httpClient, os.Getenv("OPENAQ_API_KEY"))

	data, err := openAQClient.GetInfo(OpenAQ_InformationRequest{
		ISOCode: " no ",
		PM25:    true,
		PM10:    false,
	})
	if err != nil {
		t.Fatal(err)
	}

	if data.MeanPM25 == nil {
		t.Fatal("expected MeanPM25 in response")
	}
	if *data.MeanPM25 <= 0 {
		t.Fatalf("expected MeanPM25 to be > 0, got %v", *data.MeanPM25)
	}

	if data.MeanPM10 != nil {
		t.Fatal("expected MeanPM10 to be nil when not requested")
	}

	t.Logf("mean PM2.5: %v", *data.MeanPM25)
}
