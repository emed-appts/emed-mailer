package job

import (
	"bytes"
	"time"

	"github.com/emed-appts/emed-mailer/internal/template"

	"github.com/rs/zerolog/log"
)

// Mailer interface
type Mailer interface {
	Run(<-chan struct{}) error
	SendMessage(string, string) error
}

// ApptChange struct
type ApptChange struct {
	Time        time.Time
	Appointment time.Time
	PatientID   int
	PatientName string
	IsBooking   bool
}

// Collector interface
type Collector interface {
	// collects latest changed appointments ordered by time of change
	CollectChangedAppts(time.Time) ([]*ApptChange, error)
}

// Job interface
type Job interface {
	Run()
}

type changedApptsJob struct {
	collector Collector
	mailer    Mailer
	lastRun   time.Time
}

// New creates a Job instance
func New(collector Collector, mailer Mailer) Job {
	return &changedApptsJob{
		collector: collector,
		mailer:    mailer,
		lastRun:   time.Now(),
	}
}

// Run executes the job once
func (job *changedApptsJob) Run() {
	// store execution time
	run := time.Now()

	changedAppts, err := job.collector.CollectChangedAppts(job.lastRun)
	if err != nil {
		log.Error().
			Err(err).
			Msg("collect updated appointments failed")

		return
	}
	templateData := struct {
		LastRun      time.Time
		ChangedAppts []*ApptChange
	}{
		LastRun:      job.lastRun,
		ChangedAppts: changedAppts,
	}

	buf := new(bytes.Buffer)
	if err := template.Execute(buf, "changedappts.tmpl", templateData); err != nil {
		log.Error().
			Err(err).
			Msg("could not execute template")

		return
	}

	if err := job.mailer.SendMessage("text/html", buf.String()); err != nil {
		log.Error().
			Err(err).
			Msg("could not send message")

		return
	}

	// set lastRun time
	job.lastRun = run
}
