package collector

import (
	"database/sql"
	"strings"
	"time"

	"github.com/emed-appts/emed-mailer/pkg/job"

	"github.com/pkg/errors"
)

type logEntry struct {
	logTime time.Time
	action  string
	date    time.Time
	time    string
	pid     int
	txt     string
}

type DBCollector struct {
	db *sql.DB
}

func New(db *sql.DB) job.Collector {
	return &DBCollector{db}
}

func (collector *DBCollector) CollectChangedAppts(lastRun time.Time) ([]*job.ApptChange, error) {
	// fetch all changed appointments since `lastRun`
	rows, err := collector.db.Query("SELECT datlog, action, datum, zeit, pid, txt FROM pds6_kallog WHERE usc = 'eT' AND datlog > @p1 ORDER BY datlog ASC", lastRun)
	if err != nil {
		return nil, errors.Wrap(err, "could not query database")
	}
	defer rows.Close()

	var changedAppts []*job.ApptChange
	for rows.Next() {
		entry := &logEntry{}
		err := rows.Scan(&entry.logTime, &entry.action, &entry.date, &entry.time, &entry.pid, &entry.txt)
		if err != nil {
			return nil, errors.Wrap(err, "could not scan database row")
		}

		// txt contains <name>, <anything>
		name := strings.SplitN(entry.txt, ",", 2)[0]

		// string -> time.Duration
		dur, err := parseTimeDuration(entry.time)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		changedAppts = append(changedAppts, &job.ApptChange{
			Time:        entry.logTime,
			Appointment: entry.date.Add(dur),
			PatientID:   entry.pid,
			PatientName: name,
			IsBooking:   entry.action == "eFill",
		})
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrap(err, "row got an error")
	}

	return changedAppts, nil
}

func parseTimeDuration(value string) (time.Duration, error) {
	loc, err := time.LoadLocation("Europe/Vienna")
	if err != nil {
		return 0, errors.Wrap(err, "could not load location \"Europe/Vienna\"")
	}
	t, err := time.ParseInLocation("15:04", value, loc)
	if err != nil {
		return 0, errors.Wrap(err, "could not parse time")
	}

	dur := time.Hour*time.Duration(t.Hour()) + time.Minute*time.Duration(t.Minute())

	return dur, nil
}
