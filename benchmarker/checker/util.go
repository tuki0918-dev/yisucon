package checker

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	rootWithoutLogin = checkHTML(func(doc *goquery.Document) error {
		login := doc.Find(".login")
		if login.Length() == 0 {
			return errors.New("未ログイン時にログインフォームが見つかりません")
		}
		logout := doc.Find(".logout")
		if logout.Length() != 0 {
			return errors.New("未ログイン時にログアウトボタンが存在します")
		}
		name := doc.Find(".name")
		if name.Text() != "こんにちは ゲストさん" {
			return errors.New("非ログイン時にユーザー名が見つかりません")
		}
		post := doc.Find(".post")
		if post.Length() != 0 {
			return errors.New("未ログイン時にツイートフォームが存在します")
		}

		return nil
	})
)

const (
	jsMD5  = "287d922ddd86b00403c00e97e2d79432"
	cssMD5 = "9dca706b1509accdaa68f07155a3c45f"
)

func checkHTML(f func(*goquery.Document) error) func(io.Reader) error {
	return func(r io.Reader) error {
		doc, err := goquery.NewDocumentFromReader(r)
		if err != nil {
			return err
		}
		return f(doc)
	}
}

func randomIntString() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprint(rand.Intn(10000))
}

func randomWord() string {
	rand.Seed(time.Now().UnixNano())
	words := []string{"スポーツ", "sports", "募集", "ダイエット", "travel", "旅行", "海外", "foods", "食事", "美味しい", "おすすめ"}
	return words[rand.Intn(len(words))]
}

func randomPass() string {
	atoz := []rune("abcdefghijklmnopqrstuvwxyz")
	buf := make([]rune, rand.Intn(4)+4)
	for i := range buf {
		buf[i] = atoz[rand.Intn(len(atoz))]
	}
	return string(buf)
}
