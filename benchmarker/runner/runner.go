package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gocraft/dbr"

	"github.com/yahoojapan/yisucon/benchmarker/config"
	"github.com/yahoojapan/yisucon/benchmarker/db"
	"github.com/yahoojapan/yisucon/benchmarker/logger"
	"github.com/yahoojapan/yisucon/benchmarker/model"
	"github.com/yahoojapan/yisucon/benchmarker/processor"
)

func Run() error {
	l := logger.GetLogger()

	for {
		q, err := db.QueueChecker(config.QueueCheckDuration)

		if err != nil {
			return err
		}

		l.Printf("BENCH team#%d Started...\n", q.TeamID.Int64)

		score := &model.Score{
			Score:   dbr.NewNullInt64(0),
			QueueID: q.QueueID,
			Message: dbr.NewNullString(""),
		}

		q.Host.String = strings.NewReplacer("http://", "",
			"https://", "").Replace(q.Host.String)

		if err = initialize(q.Host.String, q.TeamID.Int64, time.Second*10); err != nil {
			l.Println("runner : initialize error")
			l.Println(err)
			score.Errors = append(score.Errors, &model.Error{
				Error:   err,
				Message: err.Error(),
			})
		} else {
			if p, err := processor.NewProcessor(q.Host.String); err != nil {
				l.Println(err)
				score.Errors = append(score.Errors, &model.Error{
					Error:   err,
					Message: err.Error(),
				})
			} else {
				score = p.Run(config.BenchTimeLimit)
				score.QueueID = q.QueueID
				l.Printf("Score : %d\n", score.Score.Int64)
			}
		}

		if err = finalize(q.Host.String, q.TeamID.Int64); err != nil {
			l.Println("runner : finalize error")
			l.Println(err)
			score.Errors = append(score.Errors, &model.Error{
				Error:   err,
				Message: err.Error(),
			})
		}

		if err = db.SaveResult(q.TeamID.Int64, score); err != nil {
			return err
		}

		l.Printf("BENCH team#%d Done.\n", q.TeamID.Int64)
	}
}

func initialize(host string, teamID int64, dur time.Duration) error {

	val, err := json.Marshal(&model.ProtalHook{
		TeamID: teamID,
		Status: 2,
	})

	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Post(fmt.Sprintf("http://%s/%s/%d", config.PortalHost, "api/benches", teamID), "application/json", bytes.NewReader(val))

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func finalize(host string, teamID int64) error {

	val, err := json.Marshal(&model.ProtalHook{
		TeamID: teamID,
		Status: 0,
	})

	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Post(fmt.Sprintf("http://%s/%s/%d", config.PortalHost, "api/benches", teamID), "application/json", bytes.NewReader(val))

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}
