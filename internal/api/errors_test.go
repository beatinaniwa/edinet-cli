package api

import (
	"testing"
)

func TestParseErrorResponse_NormalFormat(t *testing.T) {
	body := []byte(`{
		"metadata": {
			"title": "提出された書類を把握するためのAPI",
			"status": "404",
			"message": "Not Found"
		}
	}`)
	err := ParseErrorResponse(body)
	if err == nil {
		t.Fatal("ParseErrorResponse() = nil, want error")
	}
	if err.Code != ErrNotFound {
		t.Errorf("Code = %q, want %q", err.Code, ErrNotFound)
	}
	if err.Status != 404 {
		t.Errorf("Status = %d, want %d", err.Status, 404)
	}
	if err.Message != "Not Found" {
		t.Errorf("Message = %q, want %q", err.Message, "Not Found")
	}
}

func TestParseErrorResponse_BadRequest(t *testing.T) {
	body := []byte(`{
		"metadata": {
			"title": "提出された書類を把握するためのAPI",
			"status": "400",
			"message": "Bad Request"
		}
	}`)
	err := ParseErrorResponse(body)
	if err == nil {
		t.Fatal("ParseErrorResponse() = nil, want error")
	}
	if err.Code != ErrBadRequest {
		t.Errorf("Code = %q, want %q", err.Code, ErrBadRequest)
	}
	if err.Status != 400 {
		t.Errorf("Status = %d, want %d", err.Status, 400)
	}
}

func TestParseErrorResponse_ServerError(t *testing.T) {
	body := []byte(`{
		"metadata": {
			"title": "提出された書類を把握するためのAPI",
			"status": "500",
			"message": "Internal Server Error"
		}
	}`)
	err := ParseErrorResponse(body)
	if err == nil {
		t.Fatal("ParseErrorResponse() = nil, want error")
	}
	if err.Code != ErrServer {
		t.Errorf("Code = %q, want %q", err.Code, ErrServer)
	}
	if err.Status != 500 {
		t.Errorf("Status = %d, want %d", err.Status, 500)
	}
}

func TestParseErrorResponse_AuthFormat401(t *testing.T) {
	body := []byte(`{
		"StatusCode": 401,
		"message": "Access denied due to invalid subscription key. Make sure to provide a valid key for an active subscription."
	}`)
	err := ParseErrorResponse(body)
	if err == nil {
		t.Fatal("ParseErrorResponse() = nil, want error")
	}
	if err.Code != ErrAuthFailed {
		t.Errorf("Code = %q, want %q", err.Code, ErrAuthFailed)
	}
	if err.Status != 401 {
		t.Errorf("Status = %d, want %d", err.Status, 401)
	}
	if err.Message != "Access denied due to invalid subscription key. Make sure to provide a valid key for an active subscription." {
		t.Errorf("Message = %q", err.Message)
	}
}

func TestParseErrorResponse_SuccessResponse(t *testing.T) {
	body := []byte(`{
		"metadata": {
			"title": "提出された書類を把握するためのAPI",
			"parameter": {"date": "2023-04-03", "type": "2"},
			"resultset": {"count": 1},
			"processDateTime": "2023-04-03 13:01",
			"status": "200",
			"message": "OK"
		},
		"results": []
	}`)
	err := ParseErrorResponse(body)
	if err != nil {
		t.Errorf("ParseErrorResponse() for status 200 should return nil, got %v", err)
	}
}

func TestParseErrorResponse_MalformedJSON(t *testing.T) {
	body := []byte(`{invalid json`)
	err := ParseErrorResponse(body)
	if err == nil {
		t.Fatal("ParseErrorResponse() = nil for malformed JSON, want error")
	}
	if err.Code != ErrInternal {
		t.Errorf("Code = %q, want %q", err.Code, ErrInternal)
	}
	if err.Raw != "{invalid json" {
		t.Errorf("Raw = %q, want original body", err.Raw)
	}
}

func TestParseErrorResponse_EmptyBody(t *testing.T) {
	err := ParseErrorResponse([]byte{})
	if err == nil {
		t.Fatal("ParseErrorResponse() = nil for empty body, want error")
	}
	if err.Code != ErrInternal {
		t.Errorf("Code = %q, want %q", err.Code, ErrInternal)
	}
}
