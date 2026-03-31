package telemt

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"multipanel/internal/config"
)

type Client struct {
	httpClient *http.Client
}

func NewClient(timeout time.Duration, allowInsecureTLS bool) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if allowInsecureTLS {
		transport.TLSClientConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true,
		}
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}
}

func (c *Client) Collect(ctx context.Context, endpoint config.Endpoint) SourceData {
	result := SourceData{
		Endpoint:    endpoint,
		BaseURL:     endpoint.BaseURL(),
		CollectedAt: time.Now().UTC(),
	}

	var (
		health          *HealthData
		summary         *SummaryData
		users           []UserInfo
		activeIPs       []UserActiveIPs
		gates           *RuntimeGatesData
		meQuality       *RuntimeMeQualityData
		upstreamQuality *RuntimeUpstreamQualityData
		meWriters       *MeWritersData
		dcStatus        *DCSummaryData
		mePool          *RuntimeMePoolStateData
		minimalAll      *MinimalAllData
		natStun         *RuntimeNatStunData
		healthErr       error
		summaryErr      error
		usersErr        error
		activeErr       error
		gatesErr        error
		meQualityErr    error
		upstreamErr     error
		meWritersErr    error
		dcStatusErr     error
		mePoolErr       error
		minimalAllErr   error
		natStunErr      error
	)

	var wg sync.WaitGroup
	wg.Add(12)

	go func() {
		defer wg.Done()
		var data HealthData
		if err := fetchJSON(ctx, c.httpClient, endpoint, "/v1/health", &data); err != nil {
			healthErr = fmt.Errorf("health: %w", err)
			return
		}
		health = &data
	}()

	go func() {
		defer wg.Done()
		var data SummaryData
		if err := fetchJSON(ctx, c.httpClient, endpoint, "/v1/stats/summary", &data); err != nil {
			summaryErr = fmt.Errorf("summary: %w", err)
			return
		}
		summary = &data
	}()

	go func() {
		defer wg.Done()
		var data []UserInfo
		if err := fetchJSON(ctx, c.httpClient, endpoint, "/v1/users", &data); err != nil {
			usersErr = fmt.Errorf("users: %w", err)
			return
		}
		users = data
	}()

	go func() {
		defer wg.Done()
		var data []UserActiveIPs
		if err := fetchJSON(ctx, c.httpClient, endpoint, "/v1/stats/users/active-ips", &data); err != nil {
			activeErr = fmt.Errorf("active_ips: %w", err)
			return
		}
		activeIPs = data
	}()

	go func() {
		defer wg.Done()
		var data RuntimeGatesData
		if err := fetchJSON(ctx, c.httpClient, endpoint, "/v1/runtime/gates", &data); err != nil {
			gatesErr = fmt.Errorf("runtime_gates: %w", err)
			return
		}
		gates = &data
	}()

	go func() {
		defer wg.Done()
		var data RuntimeMeQualityData
		if err := fetchJSON(ctx, c.httpClient, endpoint, "/v1/runtime/me_quality", &data); err != nil {
			meQualityErr = fmt.Errorf("me_quality: %w", err)
			return
		}
		meQuality = &data
	}()

	go func() {
		defer wg.Done()
		var data RuntimeUpstreamQualityData
		if err := fetchJSON(ctx, c.httpClient, endpoint, "/v1/runtime/upstream_quality", &data); err != nil {
			upstreamErr = fmt.Errorf("upstream_quality: %w", err)
			return
		}
		upstreamQuality = &data
	}()

	go func() {
		defer wg.Done()
		var data MeWritersData
		if err := fetchJSON(ctx, c.httpClient, endpoint, "/v1/stats/me-writers", &data); err != nil {
			meWritersErr = fmt.Errorf("me_writers: %w", err)
			return
		}
		meWriters = &data
	}()

	go func() {
		defer wg.Done()
		var data DCSummaryData
		if err := fetchJSON(ctx, c.httpClient, endpoint, "/v1/stats/dcs", &data); err != nil {
			dcStatusErr = fmt.Errorf("dc_status: %w", err)
			return
		}
		dcStatus = &data
	}()

	go func() {
		defer wg.Done()
		var data RuntimeMePoolStateData
		if err := fetchJSON(ctx, c.httpClient, endpoint, "/v1/runtime/me_pool_state", &data); err != nil {
			mePoolErr = fmt.Errorf("me_pool_state: %w", err)
			return
		}
		mePool = &data
	}()

	go func() {
		defer wg.Done()
		var data MinimalAllData
		if err := fetchJSON(ctx, c.httpClient, endpoint, "/v1/stats/minimal/all", &data); err != nil {
			minimalAllErr = fmt.Errorf("minimal_all: %w", err)
			return
		}
		minimalAll = &data
	}()

	go func() {
		defer wg.Done()
		var data RuntimeNatStunData
		if err := fetchJSON(ctx, c.httpClient, endpoint, "/v1/runtime/nat_stun", &data); err != nil {
			natStunErr = fmt.Errorf("nat_stun: %w", err)
			return
		}
		natStun = &data
	}()

	wg.Wait()

	result.Summary = summary
	result.Users = users
	result.ActiveIPs = activeIPs
	result.Gates = gates
	result.MeQuality = meQuality
	result.UpstreamQuality = upstreamQuality
	result.MeWriters = meWriters
	result.DCStatus = dcStatus
	result.MePool = mePool
	result.MinimalAll = minimalAll
	result.NatStun = natStun

	if health != nil {
		result.ReadOnly = health.ReadOnly
	}

	for _, err := range []error{
		healthErr,
		summaryErr,
		usersErr,
		activeErr,
		gatesErr,
		meQualityErr,
		upstreamErr,
		meWritersErr,
		dcStatusErr,
		mePoolErr,
		minimalAllErr,
		natStunErr,
	} {
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
		}
	}

	successfulRequests := 0
	for _, ok := range []bool{
		health != nil,
		summary != nil,
		usersErr == nil,
		activeErr == nil,
		gates != nil,
		meQuality != nil,
		upstreamQuality != nil,
		meWriters != nil,
		dcStatus != nil,
		mePool != nil,
		minimalAll != nil,
		natStun != nil,
	} {
		if ok {
			successfulRequests++
		}
	}

	result.Reachable = successfulRequests > 0
	result.Partial = result.Reachable && len(result.Errors) > 0

	return result
}

func fetchJSON[T any](ctx context.Context, client *http.Client, endpoint config.Endpoint, path string, out *T) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.BaseURL()+path, nil)
	if err != nil {
		return err
	}
	if endpoint.AuthHeader != "" {
		req.Header.Set("Authorization", endpoint.AuthHeader)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		message := strings.TrimSpace(string(payload))
		if message == "" {
			message = http.StatusText(resp.StatusCode)
		}
		return fmt.Errorf("status %d: %s", resp.StatusCode, message)
	}

	var body envelope[T]
	if err := json.Unmarshal(payload, &body); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if !body.OK {
		message := body.Error.Message
		if message == "" {
			message = "telemt api returned ok=false"
		}
		return fmt.Errorf("%s", message)
	}

	*out = body.Data
	return nil
}
