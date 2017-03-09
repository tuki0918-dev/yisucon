package data

import (
	"errors"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/yahoojapan/yisucon/benchmarker/model"
	"github.com/yahoojapan/yisucon/benchmarker/util"
)

var (
	account []*model.Account
	once    sync.Once
)

func GetAccounts() ([]*model.Account, error) {
	once.Do(func() {
		account = userNameReader()
	})

	if account == nil || len(account) == 0 {
		return nil, errors.New("user data not found")
	}

	return shuffleAccount(account), nil
}

func userNameReader() []*model.Account {
	var accounts []*model.Account

	for _, v := range strings.Split(Names, ",") {
		accounts = append(accounts, &model.Account{
			Name: v,
			Pass: util.Cipher(v),
		})
	}

	return accounts
}

func shuffleAccount(accounts []*model.Account) []*model.Account {
	rand.Seed(time.Now().UnixNano())
	for i := range accounts {
		j := rand.Intn(i + 1)
		accounts[i], accounts[j] = accounts[j], accounts[i]
	}
	return accounts
}
