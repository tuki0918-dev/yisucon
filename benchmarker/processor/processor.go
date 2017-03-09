package processor

import (
	"container/ring"
	"context"
	"sync"
	"time"

	"github.com/yahoojapan/yisucon/benchmarker/checker"
	"github.com/yahoojapan/yisucon/benchmarker/config"
	"github.com/yahoojapan/yisucon/benchmarker/logger"
	"github.com/yahoojapan/yisucon/benchmarker/model"
	"github.com/yahoojapan/yisucon/benchmarker/score"
	"github.com/yahoojapan/yisucon/benchmarker/worker"
)

type Processor struct {
	w      *ring.Ring
	wg     *sync.WaitGroup //Process WaitGroup
	cwg    *sync.WaitGroup //Broadcast WaitGroup
	cond   *sync.Cond      //Broadcast Condition
	ctx    context.Context
	cancel context.CancelFunc
	result chan score.Score
	done   chan struct{}
	log    *logger.Logger
}

func NewProcessor(host string) (*Processor, error) {
	w, err := worker.NewWorkers(host)

	if err != nil {
		return nil, err
	}

	return &Processor{
		w:      w,
		wg:     new(sync.WaitGroup),
		cwg:    new(sync.WaitGroup),
		cond:   sync.NewCond(new(sync.Mutex)),
		result: make(chan score.Score, config.MaxWorkerCount*config.MaxCheckers),
		done:   make(chan struct{}, config.MaxWorkerCount),
		log:    logger.GetLogger(),
	}, nil
}

func (p *Processor) work() {
	defer p.wg.Done()
	p.w = p.w.Next()
	err := p.w.Value.(*worker.Worker).Run(p.ctx, p.result)
	if err != nil {
		p.log.Println(err)
		return
	}
	p.done <- struct{}{}
}

func (p *Processor) Run(dur time.Duration) *model.Score {

	defer close(p.done)
	defer close(p.result)
	defer p.wg.Wait()

	s := new(model.Score)

	var err error

	s.Score.Int64, err = p.initialProcess()

	if err != nil {
		p.log.Println("processor : initialProcess error")
		s.Errors = append(s.Errors, &model.Error{
			Error:   err,
			Message: err.Error(),
		})
		return s
	}

	for i := 0; i < cap(p.done); i++ {
		p.wg.Add(1)
		p.cwg.Add(1)
		go func() {
			defer p.work()
			defer p.cond.L.Unlock()
			p.cond.L.Lock()
			p.cwg.Done()
			p.cond.Wait()
		}()
	}

	p.cwg.Wait()

	//Start timer
	p.ctx, p.cancel = context.WithTimeout(context.Background(), dur)

	start := time.Now()

	defer p.cancel()

	p.cond.Broadcast()

	for {
		select {
		case <-p.ctx.Done():
			p.log.Println(time.Since(start))
			if s.Score.Int64 < 0 {
				s.Score.Int64 = 0
			}
			return s
		case <-p.done:
			p.wg.Add(1)
			go p.work()
		case result := <-p.result:
			go func(res *score.Score) {
				s.Score.Int64 += int64(res.Score)
				if res.Error != nil {
					s.Errors = append(s.Errors, &model.Error{
						Error: res.Error,
					})
				}
			}(&result)
		}
	}
}

func (p *Processor) initialProcess() (s int64, err error) {
	w := p.w.Value.(*worker.Worker)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := checker.NewChecker(ctx, w.Host, w.Account)
	defer c.Close()

	scenario := checker.NewInitScenario(c)
	defer scenario.Close()

	res := make(chan score.Score, 2)
	defer close(res)

	wg := new(sync.WaitGroup)
	defer wg.Wait()

	res <- score.Score{
		Error: nil,
		Score: 0,
	}

	for {
		select {
		case <-ctx.Done():
			return s, nil
		case sc := <-res:
			if sc.Error != nil {
				return 0, sc.Error
			}
			s += int64(sc.Score)

			if scenario.IsEmpty() {
				return s, nil
			}

			wg.Add(1)
			go func(action *checker.Action) {
				res <- score.CalcScore(action.Method, action.Name, action.Action)
				wg.Done()
			}(scenario.Pop())
		}
	}
}
