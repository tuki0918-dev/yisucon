package checker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yahoojapan/yisucon/benchmarker/config"
	"github.com/yahoojapan/yisucon/benchmarker/logger"
	"github.com/yahoojapan/yisucon/benchmarker/model"
	"github.com/yahoojapan/yisucon/benchmarker/session"
	"github.com/yahoojapan/yisucon/benchmarker/util"
	"github.com/PuerkitoBio/goquery"
)

type Checker struct {
	Account *model.Account
	Host    string
	Session *session.Session
	Logger  *logger.Logger
}

func NewChecker(ctx context.Context, host string, account *model.Account) *Checker {
	return &Checker{
		Account: account,
		Host:    host,
		Session: session.NewSession(ctx, host),
		Logger:  logger.GetLogger(),
	}
}

func (c *Checker) Close() {
	c.Session.Close()
	c.Account = nil
	c.Session = nil
}

func (c *Checker) InitialCheck() (int, error) {
	//初期チェック
	ctx, cancel := context.WithTimeout(context.Background(), config.InitializeTimeout)
	defer cancel()

	errChan := make(chan error, 1)
	done := make(chan struct{}, 0)

	go func() {
		defer close(done)
		defer close(errChan)

		resp, err := c.Session.SendSimpleRequest(http.MethodGet, fmt.Sprintf("http://%s/%s", c.Host, "initialize"), nil)

		if err != nil {
			errChan <- err
			return
		}

		if resp == nil || resp.Body == nil {
			errChan <- errors.New("初期化Responseがnilです")
			return
		}

		defer resp.Body.Close()

		var result map[string]interface{}

		json.NewDecoder(resp.Body).Decode(&result)

		if val, ok := result["result"]; !ok || !strings.EqualFold(val.(string), "ok") {
			errChan <- errors.New("初期化処理に失敗しました")
			return
		}

		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		//Timeout Failed
		return -1, errors.New("初期化処理が長すぎます")
	case res := <-errChan:
		return -1, res
	case <-done:
		return 1, nil
	}
}

func (c *Checker) JSCheck() (int, error) {
	//js, css を md5でチェック
	resp, err := c.Session.SendSimpleRequest(http.MethodGet, fmt.Sprintf("http://%s/js/script.js", c.Host), nil)

	if err != nil {
		return -1, err
	}

	if resp == nil || resp.Body == nil {
		return -1, http.ErrBodyNotAllowed
	}

	defer resp.Body.Close()

	if util.GetMD5ByIO(resp.Body) != jsMD5 {
		return -1, errors.New("不正なJavaScriptファイルです")
	}

	return 1, nil
}

func (c *Checker) CSSCheck() (int, error) {
	//js, css を md5でチェック
	resp, err := c.Session.SendSimpleRequest(http.MethodGet, fmt.Sprintf("http://%s/css/style.css", c.Host), nil)

	if err != nil {
		return -1, err
	}

	if resp == nil || resp.Body == nil {
		return -1, http.ErrBodyNotAllowed
	}

	defer resp.Body.Close()

	if util.GetMD5ByIO(resp.Body) != cssMD5 {
		return -1, errors.New("不正なCSSファイルです")
	}

	return 1, nil

}

func (c *Checker) FaviconCheck() (int, error) {
	//favicon叩くだけ
	_, err := c.Session.SendSimpleRequest(http.MethodGet, fmt.Sprintf("http://%s/favicon.ico", c.Host), nil)

	if err != nil {
		return 0, nil
	}

	return 1, nil

}

func (c *Checker) PageLoadCheck() (int, error) {
	//各ページが読み込めること
	score := 0
	path := []string{
		"",
		c.Account.Name,
		"search?q=" + url.QueryEscape(randomWord()),
		"hashtag/" + url.QueryEscape(randomWord()),
	}
	for _, p := range path {
		_, err := c.Session.SendSimpleRequest(http.MethodGet, fmt.Sprintf("http://%s/%s", c.Host, p), nil)
		if err != nil {
			score -= 10
		} else {
			score++
		}
	}
	return score, nil
}

func (c *Checker) MyPageCheck() (int, error) {
	//自分のページ
	resp, err := c.Session.SendSimpleRequest(http.MethodGet, fmt.Sprintf("http://%s/%s", c.Host, c.Account.Name), nil)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	err = checkHTML(func(doc *goquery.Document) error {
		text := doc.Find("h3").Text()
		if text != c.Account.Name+" さんのツイート" {
			return errors.New("タイトルが不適切です")
		}
		var err error
		doc.Find(".tweet").EachWithBreak(func(_ int, s *goquery.Selection) bool {
			name := s.Find(".tweet-user-name").Text()
			if name != c.Account.Name {
				err = errors.New("異なるユーザーのツイートが含まれています")
				return false
			}
			return true
		})
		return err
	})(resp.Body)
	if err != nil {
		c.Logger.Println(err)
		return -1, err
	}

	return 1, nil
}

func (c *Checker) LoginPageCheck() (int, error) {
	// 未ログイン状態の/はログインページ（tweetフォームがない）
	resp, err := c.Session.SendSimpleRequest(http.MethodGet, fmt.Sprintf("http://%s/", c.Host), nil)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	err = rootWithoutLogin(resp.Body)
	if err != nil {
		return -1, err
	}
	return 1, nil
}

func (c *Checker) FakeLoginCheck() (int, error) {
	//ログインできないこと
	resp, err := c.Session.SendFormPost(fmt.Sprintf("http://%s/login", c.Host), map[string]string{
		"name":     c.Account.Name,
		"password": randomPass(),
	})

	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	err = checkHTML(func(doc *goquery.Document) error {
		flush := doc.Find(".flush")
		if flush.Length() == 0 {
			return errors.New("ログインエラーが見つかりません")
		}

		return nil
	})(resp.Body)

	if err != nil {
		c.Logger.Println(err)
		return -1, err
	}

	return 1, nil
}

func (c *Checker) LoginCheck() (int, error) {
	//ログインできること

	resp, err := c.Session.SendFormPost(fmt.Sprintf("http://%s/login", c.Host), map[string]string{
		"name":     c.Account.Name,
		"password": c.Account.Pass,
	})

	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	err = checkHTML(func(doc *goquery.Document) error {
		login := doc.Find(".login")
		if login.Length() != 0 {
			return errors.New("ログイン時にログインフォームが存在します")
		}
		logout := doc.Find(".logout")
		if logout.Length() == 0 {
			return errors.New("ログイン時にログアウトボタンが見つかりません")
		}
		name := doc.Find(".name")
		if name.Text() != "こんにちは "+c.Account.Name+"さん" {
			return errors.New("ログイン時にユーザー名が見つかりません")
		}
		post := doc.Find(".post")
		if post.Length() == 0 {
			return errors.New("ログイン時にツイートフォームが見つかりません")
		}

		doc.Find(".tweet").EachWithBreak(func(_ int, s *goquery.Selection) bool {
			name := s.Find(".tweet-user-name").Text()
			if name != c.Account.Name {
				c.Session.Storage["firstuser"] = name
				return false
			}
			return true
		})
		if _, ok := c.Session.Storage["firstuser"]; !ok {
			return errors.New("タイムラインに他ユーザーが見つかりません")
		}
		until, ok := doc.Find(".tweet").Last().Attr("data-time")
		if !ok {
			return errors.New("data-time属性が見つかりません")
		}
		c.Session.Storage["until"] = url.QueryEscape(until)

		return nil
	})(resp.Body)

	if err != nil {
		c.Logger.Println(err)
		return -1, err
	}

	return 1, nil
}

func (c *Checker) PagingCheck() (int, error) {
	//ログイン後に / パラメータ付きで数ページチェック 200
	//http://localhost:8080/?append=1&until=time_string
	until, ok := c.Session.Storage["until"].(string)
	if !ok {
		return -1, errors.New("untilパラメータが見つかりません")
	}

	resp, err := c.Session.SendSimpleRequest(http.MethodGet, fmt.Sprintf("http://%s/?append=1&until=%s", c.Host, until), nil)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	err = checkHTML(func(doc *goquery.Document) error {
		var err error
		tweets := doc.Find(".tweet")
		if tweets.Length() != 50 {
			return errors.New("表示されているツイートが足りません")
		}

		until, _ = url.QueryUnescape(until)
		newer, err := time.Parse("2006-01-02 15:04:05", until)
		if err != nil {
			return errors.New("untilパラメータの形式が不適切です")
		}
		tweets.EachWithBreak(func(_ int, s *goquery.Selection) bool {
			attr, ok := s.Attr("data-time")
			if !ok {
				err = errors.New("data-time属性が見つかりません")
				return false
			}
			older, err := time.Parse("2006-01-02 15:04:05", attr)
			if err != nil {
				err = errors.New("data-time属性の形式が不適切です")
				return false
			}
			if older.After(newer) {
				err = errors.New("ツイートの並びが不適切です")
				return false
			}
			newer = older
			return true
		})
		if err != nil {
			return err
		}

		until, ok := tweets.Last().Attr("data-time")
		if !ok {
			return errors.New("data-time属性が見つかりません")
		}
		c.Session.Storage["until"] = url.QueryEscape(until)
		return nil
	})(resp.Body)
	if err != nil {
		c.Logger.Println(err)
		return -1, err
	}
	return 1, nil
}

func (c *Checker) SelfPageCheck() (int, error) {
	//自分: あなたのページです
	resp, err := c.Session.SendSimpleRequest(http.MethodGet, fmt.Sprintf("http://%s/%s", c.Host, c.Account.Name), nil)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	err = checkHTML(func(doc *goquery.Document) error {
		text := doc.Find(`h4`).Text()
		if text != "あなたのページです" {
			return errors.New("ログインユーザーのページが不適切です")
		}

		return nil
	})(resp.Body)
	if err != nil {
		c.Logger.Println(err)
		return -1, err
	}

	return 1, nil
}

func (c *Checker) UnfollowButtonCheck() (int, error) {
	//一番上のuser: unfollowボタン
	firstuser, ok := c.Session.Storage["firstuser"].(string)
	if !ok {
		return -1, errors.New("該当するユーザがいません")
	}

	resp, err := c.Session.SendSimpleRequest(http.MethodGet, fmt.Sprintf("http://%s/%s", c.Host, firstuser), nil)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	err = checkHTML(func(doc *goquery.Document) error {
		text := doc.Find(`#user-unfollow-button`).Text()
		if text != "アンフォロー" {
			return errors.New("アンフォローボタンがありません")
		}

		return nil
	})(resp.Body)
	if err != nil {
		c.Logger.Println(err)
		return -1, err
	}

	return 1, nil
}

func (c *Checker) UnfollowCheck() (int, error) {
	//unfollow
	firstuser, ok := c.Session.Storage["firstuser"].(string)
	if !ok {
		return -1, errors.New("該当するユーザがいません")
	}

	resp, err := c.Session.SendFormPost(fmt.Sprintf("http://%s/unfollow", c.Host), map[string]string{
		"user": firstuser,
	})

	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()

	return 1, nil
}

func (c *Checker) RemoveFromTopCheck() (int, error) {
	// Unfollowしたらtopから消える
	firstuser, ok := c.Session.Storage["firstuser"].(string)
	if !ok {
		return -1, errors.New("該当するユーザがいません")
	}

	resp, err := c.Session.SendSimpleRequest(http.MethodGet, fmt.Sprintf("http://%s/", c.Host), nil)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	err = checkHTML(func(doc *goquery.Document) error {
		var err error
		doc.Find(".tweet").EachWithBreak(func(_ int, s *goquery.Selection) bool {
			name := s.Find(".tweet-user-name").Text()
			if name == firstuser {
				err = errors.New("unfollowしたユーザが消えていません")
				return false
			}
			return true
		})
		return err
	})(resp.Body)
	if err != nil {
		c.Logger.Println(err)
		return -1, err
	}

	return 1, nil
}

func (c *Checker) FollowButtonCheck() (int, error) {
	//userページにfollowボタンが出る
	firstuser, ok := c.Session.Storage["firstuser"].(string)
	if !ok {
		return -1, errors.New("該当するユーザがいません")
	}

	resp, err := c.Session.SendSimpleRequest(http.MethodGet, fmt.Sprintf("http://%s/%s", c.Host, firstuser), nil)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	err = checkHTML(func(doc *goquery.Document) error {
		text := doc.Find(`#user-follow-button`).Text()
		if text != "フォロー" {
			return errors.New("フォローボタンがありません")
		}

		return nil
	})(resp.Body)
	if err != nil {
		c.Logger.Println(err)
		return -1, err
	}
	return 1, nil
}

func (c *Checker) FollowCheck() (int, error) {
	//followできる
	firstuser, ok := c.Session.Storage["firstuser"].(string)
	if !ok {
		return -1, errors.New("該当するユーザがいません")
	}

	resp, err := c.Session.SendFormPost(fmt.Sprintf("http://%s/follow", c.Host), map[string]string{
		"user": firstuser,
	})

	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	return 1, nil
}

func (c *Checker) FollowerTweetCheck() (int, error) {
	//topにフォローしたユーザーのtweetが出る(いちばん上じゃない可能性がある)
	firstuser, ok := c.Session.Storage["firstuser"].(string)
	if !ok {
		return -1, errors.New("該当するユーザがいません")
	}

	resp, err := c.Session.SendSimpleRequest(http.MethodGet, fmt.Sprintf("http://%s/", c.Host), nil)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	err = checkHTML(func(doc *goquery.Document) error {
		found := false
		doc.Find(".tweet").EachWithBreak(func(_ int, s *goquery.Selection) bool {
			name := s.Find(".tweet-user-name").Text()
			if name == firstuser {
				found = true
				return false
			}
			return true
		})
		if !found {
			return errors.New("followしたユーザが表示されていません")
		}
		return nil
	})(resp.Body)
	if err != nil {
		c.Logger.Println(err)
		return -1, err
	}

	return 1, nil
}

func (c *Checker) HashTagTweetCheck() (int, error) {
	//post（ハッシュタグ付きのデータ）
	tweet := "テストツイート" + randomIntString()
	hashtag := randomWord()
	c.Session.Storage["tweet"] = tweet
	c.Session.Storage["hashtag"] = hashtag

	resp, err := c.Session.SendFormPost(fmt.Sprintf("http://%s/", c.Host), map[string]string{
		"text": fmt.Sprintf("%s #%s", tweet, hashtag),
	})

	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()

	return 1, nil
}
func (c *Checker) TweetCheck() (int, error) {
	//topにtweetが出る(いちばん上じゃない可能性がある)
	tweet, ok := c.Session.Storage["tweet"].(string)
	if !ok {
		return -1, errors.New("投稿したツイートが見つかりません")
	}
	hashtag, ok := c.Session.Storage["hashtag"].(string)
	if !ok {
		return -1, errors.New("投稿したツイートが見つかりません")
	}

	resp, err := c.Session.SendSimpleRequest(http.MethodGet, fmt.Sprintf("http://%s/", c.Host), nil)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	err = checkHTML(func(doc *goquery.Document) error {
		found := false
		doc.Find(".tweet").EachWithBreak(func(_ int, s *goquery.Selection) bool {
			if strings.Contains(s.Text(), tweet) && s.Find(".hashtag").Text() == "#"+hashtag {
				found = true
				return false
			}
			return true
		})
		if !found {
			return errors.New("投稿したツイートが表示されていません")
		}
		return nil
	})(resp.Body)
	if err != nil {
		c.Logger.Println(err)
		return -1, err
	}

	return 1, nil
}

func (c *Checker) HashTagCheck() (int, error) {
	//ハッシュタグがリンクになる
	tweet, ok := c.Session.Storage["tweet"].(string)
	if !ok {
		return -1, errors.New("投稿したツイートが見つかりません")
	}
	hashtag, ok := c.Session.Storage["hashtag"].(string)
	if !ok {
		return -1, errors.New("投稿したツイートが見つかりません")
	}

	resp, err := c.Session.SendSimpleRequest(http.MethodGet, fmt.Sprintf("http://%s/hashtag/%s", c.Host, url.QueryEscape(hashtag)), nil)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	err = checkHTML(func(doc *goquery.Document) error {
		found := false
		doc.Find(".tweet").EachWithBreak(func(_ int, s *goquery.Selection) bool {
			if strings.Contains(s.Text(), tweet) && s.Find(".hashtag").Text() == "#"+hashtag {
				found = true
				return false
			}
			return true
		})
		if !found {
			return errors.New("投稿したツイートが表示されていません")
		}
		return nil
	})(resp.Body)
	if err != nil {
		c.Logger.Println(err)
		return -1, err
	}

	return 1, nil
}

func (c *Checker) TweetSearchCheck() (int, error) {
	//検索できる
	query := randomWord()

	resp, err := c.Session.SendSimpleRequest(http.MethodGet, fmt.Sprintf("http://%s/search?q=%s", c.Host, url.QueryEscape(query)), nil)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	err = checkHTML(func(doc *goquery.Document) error {
		var e error
		doc.Find(".tweet").Each(func(_ int, s *goquery.Selection) {
			if !strings.Contains(s.Text(), query) {
				e = errors.New("検索対象でないツイートが表示されています")
				return
			}
		})
		return e
	})(resp.Body)
	if err != nil {
		c.Logger.Println(err)
		return -1, err
	}

	return 1, nil
}

func (c *Checker) LogoutCheck() (int, error) {
	//ログアウト
	resp, err := c.Session.SendSimpleRequest(http.MethodPost, fmt.Sprintf("http://%s/logout", c.Host), nil)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	err = rootWithoutLogin(resp.Body)
	if err != nil {
		return -1, err
	}
	return 1, nil
}
