package handlers

import (
	"github.com/supanova-rp/supanova-server/internal/domain"
)

type Handlers struct {
	System domain.SystemRepository
	Course domain.CourseRepository
	Progress domain.ProgressRespository
}

func NewHandlers(
	system domain.SystemRepository,
	course domain.CourseRepository,
	progress domain.ProgressRespository,
) *Handlers {
	return &Handlers{
		System: system,
		Course: course,
		Progress: progress,
	}
}
