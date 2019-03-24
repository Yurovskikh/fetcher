package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/sync/semaphore"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	mu        sync.Mutex
	singleton *fetcher
)

type Fetcher interface {
	Get() (string, error)
	List() ([]string, error)
}

type fetcher struct {
	sem        *semaphore.Weighted
	httpClient *http.Client
	url        string
}

func NewFetcher(url string, timeout time.Duration) Fetcher {
	mu.Lock()
	defer mu.Unlock()

	if singleton == nil {
		transport := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
		httpClient := &http.Client{
			Timeout:   timeout,
			Transport: transport,
		}
		singleton = &fetcher{
			sem:        semaphore.NewWeighted(2),
			httpClient: httpClient,
			url:        url,
		}
	}
	return singleton
}

type getDataResponse struct {
	Data string `json:"data"`
}

func (f *fetcher) Get() (string, error) {
	err := f.sem.Acquire(context.TODO(), 1)
	if err != nil {
		return "", err
	}
	defer f.sem.Release(1)

	body, err := f.httpGet("/get")
	if err != nil {
		return "", err
	}
	var result getDataResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}
	return result.Data, nil
}

type getDataSetResponse struct {
	DataSet []string `json:"data_set"`
}

func (f *fetcher) List() ([]string, error) {
	err := f.sem.Acquire(context.TODO(), 1)
	if err != nil {
		return nil, err
	}
	defer f.sem.Release(1)

	body, err := f.httpGet("/list")
	if err != nil {
		return nil, err
	}
	var result getDataSetResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	return result.DataSet, nil
}

func (f *fetcher) httpGet(path string) ([]byte, error) {
	resp, err := http.Get(f.url + path)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status code %s", http.StatusText(resp.StatusCode))
	}
	return ioutil.ReadAll(resp.Body)
}
