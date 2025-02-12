package postgres_test

import (
	"context"
	"github.com/aviseu/jobs/internal/app/domain/job"
	"github.com/aviseu/jobs/internal/app/storage/postgres"
	"github.com/aviseu/jobs/internal/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

func TestJobRepository(t *testing.T) {
	suite.Run(t, new(JobRepositorySuite))
}

type JobRepositorySuite struct {
	testutils.IntegrationSuite
}

func (suite *JobRepositorySuite) Test_Save_New_Success() {
	// Prepare
	id := uuid.New()
	chID := uuid.New()
	pAt := time.Date(2025, 1, 1, 0, 1, 0, 0, time.UTC)
	j := job.New(
		id,
		chID,
		"https://example.com/job/id",
		"Software Engineer",
		"Job Description",
		"Indeed",
		"Amsterdam",
		true,
		pAt,
	)
	r := postgres.NewJobRepository(suite.DB)

	// Execute
	err := r.Save(context.Background(), j)

	// Assert return
	suite.NoError(err)

	// Assert state change
	var dbJob postgres.Job
	err = suite.DB.Get(&dbJob, "SELECT * FROM jobs WHERE id = $1", id)
	suite.NoError(err)
	suite.Equal(id, dbJob.ID)
	suite.Equal(chID, dbJob.ChannelID)
	suite.Equal("https://example.com/job/id", dbJob.URL)
	suite.Equal("Software Engineer", dbJob.Title)
	suite.Equal("Job Description", dbJob.Description)
	suite.Equal("Indeed", dbJob.Source)
	suite.Equal("Amsterdam", dbJob.Location)
	suite.True(dbJob.Remote)
	suite.True(dbJob.PostedAt.Equal(pAt))
	suite.True(dbJob.CreatedAt.After(time.Now().Add(-2 * time.Second)))
	suite.True(dbJob.UpdatedAt.After(time.Now().Add(-2 * time.Second)))
}

func (suite *JobRepositorySuite) Test_Save_Existing_Success() {
	// Prepare
	id := uuid.New()
	chID := uuid.New()
	cAt := time.Date(2025, 1, 1, 0, 2, 0, 0, time.UTC)
	_, err := suite.DB.Exec("INSERT INTO jobs (id, channel_id, url, title, description, source, location, remote, posted_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		id,
		chID,
		"https://example.com/job/id",
		"Software Engineer",
		"Job Description",
		"Indeed",
		"Amsterdam",
		true,
		time.Date(2025, 1, 1, 0, 1, 0, 0, time.UTC),
		cAt,
		time.Date(2025, 1, 1, 0, 3, 0, 0, time.UTC),
	)
	suite.NoError(err)

	pAt := time.Date(2025, 1, 1, 0, 4, 0, 0, time.UTC)
	chID2 := uuid.New()
	j := job.New(
		id,
		chID2,
		"https://example.com/job/id/new",
		"Software Engineer new",
		"Job Description new",
		"Indeed new",
		"Amsterdam new",
		false,
		pAt,
	)

	r := postgres.NewJobRepository(suite.DB)

	// Execute
	err = r.Save(context.Background(), j)

	// Assert return
	suite.NoError(err)

	// Assert state change
	var count int
	err = suite.DB.Get(&count, "SELECT COUNT(*) FROM jobs WHERE id = $1", id)
	suite.NoError(err)
	suite.Equal(1, count)

	var dbJob postgres.Job
	err = suite.DB.Get(&dbJob, "SELECT * FROM jobs WHERE id = $1", id)
	suite.NoError(err)
	suite.Equal(id, dbJob.ID)
	suite.Equal(chID2, dbJob.ChannelID)
	suite.Equal("https://example.com/job/id/new", dbJob.URL)
	suite.Equal("Software Engineer new", dbJob.Title)
	suite.Equal("Job Description new", dbJob.Description)
	suite.Equal("Indeed new", dbJob.Source)
	suite.Equal("Amsterdam new", dbJob.Location)
	suite.False(dbJob.Remote)
	suite.True(dbJob.PostedAt.Equal(pAt))
	suite.True(dbJob.CreatedAt.Equal(cAt))
	suite.True(dbJob.UpdatedAt.After(time.Now().Add(-2 * time.Second)))
}

func (suite *JobRepositorySuite) Test_Save_Error() {
	// Prepare
	id := uuid.New()
	chID := uuid.New()
	j := job.New(
		id,
		chID,
		"https://example.com/job/id",
		"Software Engineer",
		"Job Description",
		"Indeed",
		"Amsterdam",
		true,
		time.Date(2025, 1, 1, 0, 1, 0, 0, time.UTC),
	)
	r := postgres.NewJobRepository(suite.BadDB)

	// Execute
	err := r.Save(context.Background(), j)

	// Assert return
	suite.Error(err)
	suite.ErrorContains(err, id.String())
	suite.ErrorContains(err, "sql: database is closed")
}
