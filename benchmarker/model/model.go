package model

import "github.com/gocraft/dbr"

type (
	Team struct {
		ID    dbr.NullInt64  `db:"id"`
		Name  dbr.NullString `db:"name"`
		Host  dbr.NullString `db:"host"`
		Score dbr.NullInt64  `db:"best_score"`
		Lang  dbr.NullString `db:"lang"`
	}

	Queue struct {
		ID     dbr.NullInt64 `db:"id"`
		TeamID dbr.NullInt64 `db:"team_id"`
		Status dbr.NullInt64 `db:"status"`
		Date   dbr.NullTime  `db:"date"`
	}

	Score struct {
		ID      dbr.NullInt64  `db:"id"`
		QueueID dbr.NullInt64  `db:"queue_id"`
		Score   dbr.NullInt64  `db:"score"`
		Message dbr.NullString `db:"message"`
		Date    dbr.NullTime   `db:"date"`
		Errors  []*Error
	}

	User struct {
		ID      dbr.NullInt64  `db:"id"`
		TeamID  dbr.NullInt64  `db:"team_id"`
		Name    dbr.NullString `db:"name"`
		Account dbr.NullString `db:"account"`
	}

	TeamQueue struct {
		TeamID  dbr.NullInt64  `db:"team_id" json:"team_id"`
		QueueID dbr.NullInt64  `db:"queue_id" json:"queue_id"`
		Host    dbr.NullString `db:"host" json:"host"`
		Status  dbr.NullInt64  `db:"status" json:"status"`
		Date    dbr.NullTime   `db:"date" json:"date"`
	}

	Account struct {
		Name string
		Pass string
	}

	Error struct {
		Error   error  `json:"error"`
		Message string `json:"message"`
	}

	ProtalHook struct {
		TeamID int64 `json:"team_id"`
		Status int   `json:"status"`
	}
)

func (s *Score) CreateErrMessage() {
	enc := map[string]bool{}
	for _, errs := range s.Errors {
		err := errs.Error.Error()
		if len(err) != 0 && err != "\n" && err != " " && !enc[err] {
			enc[err] = true
			s.Message.String += err + "\n"
		}
	}
}
