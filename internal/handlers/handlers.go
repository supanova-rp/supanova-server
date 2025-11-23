package handlers

import (
	"github.com/supanova-rp/supanova-server/internal/domain"
)

type Handlers struct {
	System   domain.SystemRepository
	Course   domain.CourseRepository
	Progress domain.ProgressRepository
}

func NewHandlers(
	system domain.SystemRepository,
	course domain.CourseRepository,
	progress domain.ProgressRepository,
) *Handlers {
	return &Handlers{
		System:   system,
		Course:   course,
		Progress: progress,
	}
}
