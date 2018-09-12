package job

import (
	"bytes"
	"time"

	"github.com/emed-appts/emedappts-mailer/pkg/template"

	"github.com/rs/zerolog/log"
)

// Mailer interface
type Mailer interface {
	Run(<-chan struct{}) error
	SendMessage(string, string) error
}

type ApptChange struct {
	Time        time.Time
	Appointment time.Time
	PatientID   int
	PatientName string
	IsBooking   bool
}

type Collector interface {
	// collects latest changed appointments ordered by time of change
	CollectChangedAppts(time.Time) ([]*ApptChange, error)
}

type Job interface {
	Run()
}

type changedApptsJob struct {
	collector Collector
	mailer    Mailer
	lastRun   time.Time
}

func New(collector Collector, mailer Mailer) Job {
	return &changedApptsJob{
		collector: collector,
		mailer:    mailer,
		lastRun:   time.Now(),
	}
}

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
