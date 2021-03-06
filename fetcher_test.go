package fetcher

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestNewFetcher(t *testing.T) {
	defer func() {
		singleton = nil
		once = sync.Once{}
	}()

	fetchers := make([]Fetcher, 0, 100000)
	wg := sync.WaitGroup{}
	for i := 0; i < 100000; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			fetchers = append(fetchers, NewFetcher("http://localhost:8080", time.Second))
		}(&wg)
	}
	wg.Wait()
	// Удостоверимся что все экземляры ссылаются на один и тот же синглтон
	pointer := reflect.ValueOf(NewFetcher("http://localhost:8080", time.Second)).Pointer()
	for _, fetcher := range fetchers {
		assert.Equal(t, pointer, reflect.ValueOf(fetcher).Pointer())
	}
}
func TestFetcher_Get(t *testing.T) {
	defer func() {
		singleton = nil
		once = sync.Once{}
	}()

	testData := generateRandomString(10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/get", req.URL.String())
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(getDataResponse{
			Data: testData,
		})
	}))
	defer server.Close()

	fetcher := NewFetcher(server.URL, 5*time.Second)
	get, err := fetcher.Get()
	assert.Nil(t, err)
	assert.NotEmpty(t, get)
}
func TestFetcher_List(t *testing.T) {
	defer func() {
		singleton = nil
		once = sync.Once{}
	}()

	testData := make([]string, 10)
	for index := range testData {
		testData[index] = generateRandomString(10)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/list", req.URL.String())
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(getDataSetResponse{
			DataSet: testData,
		})
	}))
	defer server.Close()

	fetcher := NewFetcher(server.URL, 5*time.Second)
	list, err := fetcher.List()
	assert.Nil(t, err)
	assert.NotEmpty(t, list)
	assert.Equal(t, testData, list)
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func generateRandomString(length int) string {
	return stringWithCharset(length, charset)
}
