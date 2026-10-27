package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/tunnel"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
)

func TestResponseError(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		body        string
		wantStatus  int
		wantMessage string
	}{
		{"invalid request", 400, `{"type":"error","error":"Invalid link type"}`, 400, `Failed to activate cluster link on cluster "target" (upstream HTTP status 400): Invalid link type`},
		{"unauthenticated", 401, `{"type":"error","error":"Authentication required"}`, 401, `Failed to activate cluster link on cluster "target" (upstream HTTP status 401): Authentication required`},
		{"forbidden", 403, `{"type":"error","error":"Missing entitlement"}`, 403, `Failed to activate cluster link on cluster "target" (upstream HTTP status 403): Missing entitlement`},
		{"not found", 404, `{"type":"error","error":"Cluster not found"}`, 404, `Failed to activate cluster link on cluster "target" (upstream HTTP status 404): Cluster not found`},
		{"conflict", 409, `{"type":"error","error":"Link already exists","error_code":500}`, 409, `Failed to activate cluster link on cluster "target" (upstream HTTP status 409): Link already exists`},
		{"unprocessable", 422, `{"type":"error","error":"Invalid configuration"}`, 422, `Failed to activate cluster link on cluster "target" (upstream HTTP status 422): Invalid configuration`},
		{"rate limited", 429, `{"type":"error","error":"Too many requests"}`, 429, `Failed to activate cluster link on cluster "target" (upstream HTTP status 429): Too many requests`},
		{"timeout", 504, `{"type":"error","error":"Timed out"}`, 504, `Failed to activate cluster link on cluster "target" (upstream HTTP status 504): Timed out`},
		{"upstream failure", 500, `{"type":"error","error":"Failed creating identity"}`, 502, `Failed to activate cluster link on cluster "target" (upstream HTTP status 500): Failed creating identity`},
		{"unavailable", 503, `{"type":"error","error":"Cluster unavailable"}`, 502, `Failed to activate cluster link on cluster "target" (upstream HTTP status 503): Cluster unavailable`},
		{"redirect", 302, `<html>internal proxy details</html>`, 502, `Failed to activate cluster link on cluster "target" (upstream HTTP status 302)`},
		{"malformed", 500, `{"error":`, 502, `Failed to activate cluster link on cluster "target" (upstream HTTP status 500)`},
		{"non-error envelope", 500, `{"type":"sync","error":"Not a valid error"}`, 502, `Failed to activate cluster link on cluster "target" (upstream HTTP status 500): Not a valid error`},
		{"missing message", 500, `{"type":"error"}`, 502, `Failed to activate cluster link on cluster "target" (upstream HTTP status 500)`},
		{"malformed unauthenticated", 401, `<html>proxy details</html>`, 401, `Failed to activate cluster link on cluster "target" (upstream HTTP status 401)`},
		{"malformed forbidden", 403, `{"error":`, 403, `Failed to activate cluster link on cluster "target" (upstream HTTP status 403)`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tunnel.ResponseError(types.ClusterManagerTunnelResponse{
				Status: tt.status,
				Body:   []byte(tt.body),
			}, "target", "activate cluster link")
			if !api.StatusErrorCheck(err, tt.wantStatus) {
				t.Fatalf("expected status %d, got %v", tt.wantStatus, err)
			}
			wantMessage := tt.wantMessage
			if err.Error() != wantMessage {
				t.Fatalf("expected %q, got %q", wantMessage, err.Error())
			}

			recorder := httptest.NewRecorder()
			if err := response.SmartError(err).Render(recorder, httptest.NewRequest(http.MethodPost, "/", nil)); err != nil {
				t.Fatal(err)
			}
			var envelope api.Response
			if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
			if recorder.Code != tt.wantStatus || envelope.Code != tt.wantStatus || envelope.Type != api.ErrorResponse || envelope.Error != wantMessage {
				t.Fatalf("inconsistent error response: HTTP %d, body %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}
