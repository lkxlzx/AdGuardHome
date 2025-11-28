package dnsforward

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
	"github.com/AdguardTeam/dnsproxy/upstream"
	"github.com/miekg/dns"
)

// testUpstreamGroupRequest represents the request to test an upstream group
type testUpstreamGroupRequest struct {
	GroupID   string   `json:"group_id"`
	Upstreams []string `json:"upstreams"`
}

// upstreamTestResult represents the test result for a single upstream
type upstreamTestResult struct {
	Upstream     string `json:"upstream"`
	Status       string `json:"status"` // "ok" or "error"
	ResponseTime string `json:"response_time"`
	Error        string `json:"error"`
}

// testUpstreamGroupResponse represents the response of upstream group test
type testUpstreamGroupResponse struct {
	Status  string               `json:"status"`
	Results []upstreamTestResult `json:"results"`
	Summary struct {
		Total           int    `json:"total"`
		Success         int    `json:"success"`
		Failed          int    `json:"failed"`
		AvgResponseTime string `json:"avg_response_time"`
	} `json:"summary"`
}

// handleTestUpstreamGroup handles the POST /control/test_upstream_group request
func (s *Server) handleTestUpstreamGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := s.logger
	
	req := &testUpstreamGroupRequest{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "invalid request body: %s", err)
		return
	}

	if len(req.Upstreams) == 0 {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "no upstreams provided")
		return
	}

	// Test each upstream
	results := make([]upstreamTestResult, 0, len(req.Upstreams))
	var totalTime time.Duration
	successCount := 0

	for _, upstreamAddr := range req.Upstreams {
		result := s.testSingleUpstream(upstreamAddr)
		results = append(results, result)

		if result.Status == "ok" {
			successCount++
			// Parse response time
			if duration, err := time.ParseDuration(result.ResponseTime); err == nil {
				totalTime += duration
			}
		}
	}

	// Calculate average response time
	avgResponseTime := "N/A"
	if successCount > 0 {
		avgDuration := totalTime / time.Duration(successCount)
		avgResponseTime = avgDuration.Round(time.Millisecond).String()
	}

	// Build response
	resp := &testUpstreamGroupResponse{
		Status:  "ok",
		Results: results,
	}
	resp.Summary.Total = len(req.Upstreams)
	resp.Summary.Success = successCount
	resp.Summary.Failed = len(req.Upstreams) - successCount
	resp.Summary.AvgResponseTime = avgResponseTime

	aghhttp.WriteJSONResponseOK(ctx, l, w, r, resp)
}

// testSingleUpstream tests a single upstream server
func (s *Server) testSingleUpstream(upstreamAddr string) upstreamTestResult {
	result := upstreamTestResult{
		Upstream: upstreamAddr,
		Status:   "error",
	}

	// Create upstream
	u, err := upstream.AddressToUpstream(upstreamAddr, &upstream.Options{
		Timeout: 5 * time.Second,
	})
	if err != nil {
		result.Error = fmt.Sprintf("failed to create upstream: %s", err)
		return result
	}

	// Test with a simple DNS query (google.com A record)
	testReq := createTestDNSMessage()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	_, err = u.Exchange(testReq)
	elapsed := time.Since(start)
	
	// Check context cancellation
	if ctx.Err() != nil {
		result.Error = "query timeout"
		return result
	}

	if err != nil {
		result.Error = fmt.Sprintf("query failed: %s", err)
		return result
	}

	result.Status = "ok"
	result.ResponseTime = elapsed.Round(time.Millisecond).String()
	return result
}

// createTestDNSMessage creates a test DNS query message
func createTestDNSMessage() *dns.Msg {
	req := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id:               dns.Id(),
			RecursionDesired: true,
		},
		Question: []dns.Question{{
			Name:   "google.com.",
			Qtype:  dns.TypeA,
			Qclass: dns.ClassINET,
		}},
	}
	return req
}
