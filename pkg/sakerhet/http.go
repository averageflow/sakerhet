package sakerhet

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HttpIntegrationTestSituation struct {
	Request     *HttpIntegrationTestSituationRequest
	Expectation *HttpIntegrationTestSituationExpectation
	Timeout     time.Duration
}

type HttpHeaderValuePair struct {
	Header string
	Value  string
}

type HttpIntegrationTestSituationRequest struct {
	RequestURL     string
	RequestMethod  string
	RequestHeaders []HttpHeaderValuePair
	RequestBody    []byte
}

type HttpIntegrationTestSituationExpectation struct {
	ResponseStatusCode int
	ResponseBody       []byte
}

type HttpIntegrationTestSituationResult struct {
	ResponseStatusCode int
	ResponseBody       []byte
}

func (s HttpIntegrationTestSituation) SituationChecker() (*HttpIntegrationTestSituationResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	r, err := http.NewRequestWithContext(ctx, s.Request.RequestMethod, s.Request.RequestURL, bytes.NewBuffer(s.Request.RequestBody))
	if err != nil {
		return nil, err
	}

	for _, v := range s.Request.RequestHeaders {
		r.Header.Add(v.Header, v.Value)
	}

	client := &http.Client{}

	res, err := client.Do(r)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	received, err := io.ReadAll(res.Body)

	if s.Expectation.ResponseBody != nil {
		if err != nil {
			return nil, err
		}

		if !bytes.Equal(received, s.Expectation.ResponseBody) {
			return nil, fmt.Errorf(
				"Unexpected data received! Expected %v, got %v",
				string(s.Expectation.ResponseBody),
				string(received),
			)
		}
	}

	if res.StatusCode != s.Expectation.ResponseStatusCode {
		return nil, fmt.Errorf(
			"Unexpected status code received! Expected %d, got %d",
			s.Expectation.ResponseStatusCode,
			res.StatusCode,
		)
	}

	return &HttpIntegrationTestSituationResult{
		ResponseStatusCode: res.StatusCode,
		ResponseBody:       received,
	}, nil
}
