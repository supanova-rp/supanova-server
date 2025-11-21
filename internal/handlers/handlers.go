package handlers

import (
	"github.com/supanova-rp/supanova-server/internal/domain"
)

type Handlers struct {
	System domain.SystemRepository
	Course domain.CourseRepository
}

func NewHandlers(
	system domain.SystemRepository,
	course domain.CourseRepository,
) *Handlers {
	return &Handlers{
		System: system,
		Course: course,
	}
}
