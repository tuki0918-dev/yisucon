package score

import (
	"net"
	"net/http"

	"github.com/yahoojapan/yisucon/benchmarker/session"
)

type Score struct {
	Score int
	Error error
}

func CalcScore(method, name string, f func() (int, error)) Score {

	s := &Score{
		Score: 0,
	}

	res, err := f()

	if err != nil {
		s.Error = err
		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			s.Error = nerr
			// -リクエスト失敗(exception)数 x 20
			s.Score = -20
		} else if err == session.ErrPostTimeOut {
			// -遅延POSTレスポンス数 x 100
			s.Score = -100
		} else {
			// -サーバエラー(error)レスポンス数 x 10
			s.Score = -10
		}
	} else {
		switch {
		case method == http.MethodGet && res > 0:
			//成功レスポンス数(GET) x 1
			s.Score = res
		case method == http.MethodPost && res > 0:
			//成功レスポンス数(POST) x 2
			s.Score = res * 2
		}
	}

	return *s
}
