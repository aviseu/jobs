package domain

import (
	"github.com/google/uuid"
	"time"
)

type Job struct {
	id          uuid.UUID
	url         string
	title       string
	description string
	source      string
	location    string
	remote      bool
	postedAt    time.Time
	createdAt   time.Time
	updatedAt   time.Time
}

type JobOptional func(*Job)

func WithTimestamps(c, u time.Time) JobOptional {
	return func(j *Job) {
		j.createdAt = c
		j.updatedAt = u
	}
}

func NewJob(id uuid.UUID, url, title, description, source, location string, remote bool, postedAt time.Time, opts ...JobOptional) *Job {
	j := &Job{
		id:          id,
		url:         url,
		title:       title,
		description: description,
		source:      source,
		location:    location,
		remote:      remote,
		postedAt:    postedAt,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}

	for _, opt := range opts {
		opt(j)
	}

	return j
}

func (j *Job) ID() uuid.UUID {
	return j.id
}

func (j *Job) URL() string {
	return j.url
}

func (j *Job) Title() string {
	return j.title
}

func (j *Job) Description() string {
	return j.description
}

func (j *Job) Source() string {
	return j.source
}

func (j *Job) Location() string {
	return j.location
}

func (j *Job) Remote() bool {
	return j.remote
}

func (j *Job) PostedAt() time.Time {
	return j.postedAt
}

func (j *Job) CreatedAt() time.Time {
	return j.createdAt
}

func (j *Job) UpdatedAt() time.Time {
	return j.updatedAt
}
