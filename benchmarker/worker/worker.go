package worker

import (
	"container/ring"
	"context"
	"errors"
	"sync"

	"github.com/yahoojapan/yisucon/benchmarker/checker"
	"github.com/yahoojapan/yisucon/benchmarker/data"
	"github.com/yahoojapan/yisucon/benchmarker/model"
	"github.com/yahoojapan/yisucon/benchmarker/score"
)

type Worker struct {
	Account *model.Account
	Host    string
}

func NewWorkers(host string) (*ring.Ring, error) {

	accounts, err := data.GetAccounts()

	if err != nil {
		return nil, err
	}

	r := ring.New(len(accounts))

	for _, account := range accounts {
		r.Value = &Worker{
			Account: account,
			Host:    host,
		}
		r = r.Next()
	}

	return r, nil
}

func (w *Worker) Run(parent context.Context, sc chan score.Score) error {

	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	c := checker.NewChecker(ctx, w.Host, w.Account)
	defer c.Close()

	scenario := checker.NewDefaultScenario(c)
	defer scenario.Close()

	res := make(chan score.Score, 1)
	defer close(res)

	res <- score.Score{
		Error: nil,
		Score: 0,
	}

	wg := new(sync.WaitGroup)
	defer wg.Wait()

	for {
		select {
		case <-ctx.Done():
			return errors.New("worker time out")
		case sc <- <-res:
			if scenario.IsEmpty() {
				return nil
			}
			wg.Add(1)
			go func(action *checker.Action) {
				res <- score.CalcScore(action.Method, action.Name, action.Action)
				wg.Done()
			}(scenario.Pop())
		}
	}
}
