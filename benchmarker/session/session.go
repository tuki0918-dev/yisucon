package session

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/yahoojapan/yisucon/benchmarker/cache"
	"github.com/yahoojapan/yisucon/benchmarker/config"
)

var (
	ErrPostTimeOut       = errors.New("post request timeout")
	ErrBenchmarkerCancel = errors.New("Benchmarker Cancelled")
)

type Session struct {
	Host      string
	Client    *http.Client
	Transport *http.Transport
	Cookies   []*http.Cookie
	Cache     *cache.Cache
	Storage   map[string]interface{}

	cancel context.CancelFunc
	ctx    context.Context
}

func NewSession(ctx context.Context, host string) *Session {

	jar, _ := cookiejar.New(&cookiejar.Options{})

	tran := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 32,
	}

	sess := &Session{
		Host:      host,
		Transport: tran,
		Client: &http.Client{
			Transport: tran,
			Jar:       jar,
			Timeout:   config.RequestTimeout,
		},
		Cache:   cache.NewCache(),
		Storage: make(map[string]interface{}),
	}

	sess.ctx, sess.cancel = context.WithCancel(ctx)

	return sess
}

func (s *Session) Close() {
	defer s.cancel()
	s.Transport.CloseIdleConnections()
	s.Cache.Clear()
	s.Storage = nil
}

func (s *Session) newRequest(method, uri string, body io.Reader) (*http.Request, error) {
	parsedURL, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("不正なURLです")
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "http"
	}

	if parsedURL.Host == "" {
		parsedURL.Host = s.Host
	}

	req, err := http.NewRequest(method, parsedURL.String(), body)

	if err != nil {
		return nil, fmt.Errorf("リクエストに失敗しました")
	}

	req.Header.Del("User-Agent")
	req.Header.Set("User-Agent", config.BenchMarkerUA)
	req.Header.Add("Accept-Encoding", "gzip")

	req.WithContext(s.ctx)

	return req, nil
}

func (s *Session) doRequest(req *http.Request) (*http.Response, error) {

	var res *http.Response
	var err error

	start := time.Now()

	if val, ok := s.Cache.Get(req); ok {
		res = val.Resp
	} else {
		res, err = s.Client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("リクエストに失敗しました")
		}

		if res.StatusCode/100 == 3 {
			rreq := *req
			rreq.URL, err = url.ParseRequestURI(res.Header.Get("Location"))
			if err == nil {
				rres, err := s.Client.Transport.RoundTrip(&rreq)
				if err == nil && rres.StatusCode/100 == 2 {
					res = rres
				}
			}
		}

		if res.StatusCode/100 == 2 {
			go func() {
				data, err := cache.NewHTTPCache(res)
				if data != nil && err == nil && s.Cache.Data != nil {
					s.Cache.Set(req, data)
				}
			}()
		}
	}

	if res.StatusCode/100 != 2 && res.StatusCode/100 != 3 {
		return nil, fmt.Errorf("リクエストに失敗しました。\n%s", res.Status)
	}

	end := time.Since(start)

	if res.Header.Get("Content-Encoding") == "gzip" {
		gres, err := gzip.NewReader(res.Body)
		if err != nil {
			return res, err
		}
		res.Body = gres
	}

	if req.Method == http.MethodPost && end > config.RequestTimeout {
		return res, ErrPostTimeOut
	}

	return res, nil
}

func (s *Session) SendFormPost(uri string, body map[string]string) (*http.Response, error) {

	data := url.Values{}

	for k, v := range body {
		data.Add(k, v)
	}

	req, err := s.newRequest(http.MethodPost, uri, strings.NewReader(data.Encode()))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return s.doRequest(req)
}

func (s *Session) SendSimpleRequest(method, uri string, body io.Reader) (*http.Response, error) {
	req, err := s.newRequest(method, uri, body)

	if err != nil {
		return nil, err
	}

	return s.doRequest(req)
}
