package collector

import (
	"database/sql"
	"strings"
	"time"

	"github.com/emed-appts/emed-mailer/internal/collector/tzinfo"
	"github.com/emed-appts/emed-mailer/internal/job"

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

type dbCollector struct {
	db *sql.DB
}

// New creates a collector instance
func New(db *sql.DB) job.Collector {
	return &dbCollector{db}
}

// CollectChangedAppts gathers changed appointments since `lastRun`
func (collector *dbCollector) CollectChangedAppts(lastRun time.Time) ([]*job.ApptChange, error) {
	// fetch all changed appointments since `lastRun`
	rows, err := collector.db.Query("SELECT datlog, action, datum, zeit, pid, txt FROM pds7_kallog WHERE usc = 'eT' AND datlog > @p1 ORDER BY datlog ASC", lastRun)
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

		// string -> time.Time
		t, err := parseTime(entry.time)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		changedAppts = append(changedAppts, &job.ApptChange{
			Time:        entry.logTime,
			Appointment: entry.date.Add(timeDuration(t)),
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

// parseTime parses Time expected to be in Timezone Europe/Vienna
func parseTime(value string) (time.Time, error) {
	loc, err := tzinfo.LoadLocation("Europe/Vienna")
	if err != nil {
		return time.Time{}, errors.Wrap(err, "could not load location \"Europe/Vienna\"")
	}
	t, err := time.ParseInLocation("15:04", value, loc)

	return t, errors.Wrap(err, "could not parse time")
}

func timeDuration(t time.Time) time.Duration {
	return time.Hour*time.Duration(t.Hour()) + time.Minute*time.Duration(t.Minute())
}
