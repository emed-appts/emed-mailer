package job

import (
	"testing"
	"time"

	"github.com/emed-appts/emed-mailer/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMain(m *testing.M) {
	test.PrepareTestEnvironment(m, "../../")
}

func TestChangedApptsJob_Run(t *testing.T) {
	lastRun := time.Now().Add(time.Hour * -24)

	c := &MockCollector{}
	c.
		On("CollectChangedAppts", lastRun).
		Return([]*ApptChange{
			{
				Time:        time.Now(),
				Appointment: time.Now(),
				PatientID:   1,
				PatientName: "Firstname Lastname",
				IsBooking:   true,
			},
			{
				Time:        time.Now(),
				Appointment: time.Now(),
				PatientID:   2,
				PatientName: "Firstname Lastname",
				IsBooking:   false,
			},
		}, nil).
		Once()

	m := &MockMailer{}
	m.
		On("SendMessage", "text/html", mock.AnythingOfType("string")).
		Return(nil).
		Once()

	job := &changedApptsJob{c, m, lastRun}
	job.Run()

	// test that lastRun has been updated
	assert.True(t, job.lastRun.After(lastRun))

	c.AssertExpectations(t)
	m.AssertExpectations(t)
}
