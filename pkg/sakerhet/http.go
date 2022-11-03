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
	URL     string
	Method  string
	Headers []HttpHeaderValuePair
	Body    []byte
}

type HttpIntegrationTestSituationExpectation struct {
	StatusCode int
	Body       []byte
}

func (s HttpIntegrationTestSituation) SituationChecker() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	r, err := http.NewRequestWithContext(ctx, s.Request.Method, s.Request.URL, bytes.NewBuffer(s.Request.Body))
	if err != nil {
		return err
	}

	for _, v := range s.Request.Headers {
		r.Header.Add(v.Header, v.Value)
	}

	client := &http.Client{}

	res, err := client.Do(r)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	received, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if !bytes.Equal(received, s.Expectation.Body) {
		return fmt.Errorf("Unexpected data received! Expected %v, got %v", string(s.Expectation.Body), string(received))
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status code received! Expected %d, got %d", http.StatusOK, res.StatusCode)
	}

	return nil
}
