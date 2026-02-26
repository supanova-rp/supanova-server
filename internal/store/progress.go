package store

import (
	"context"
	"slices"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

type UserDetails struct {
	ID    string
	Name  string
	Email string
}

func (s *Store) GetProgress(ctx context.Context, args domain.GetProgressParams) (*domain.Progress, error) {
	sqlcArgs := sqlc.GetProgressParams{
		UserID:   args.UserID,
		CourseID: utils.PGUUIDFromUUID(args.CourseID),
	}

	progress, err := ExecQuery(
		ctx,
		func() (sqlc.GetProgressRow, error) { return s.Queries.GetProgress(ctx, sqlcArgs) },
	)
	if err != nil {
		return nil, err
	}

	return progressFrom(progress), nil
}

func (s *Store) UpdateProgress(ctx context.Context, args domain.UpdateProgressParams) error {
	sqlcArgs := sqlc.UpdateProgressParams{
		UserID:    args.UserID,
		CourseID:  utils.PGUUIDFromUUID(args.CourseID),
		SectionID: utils.PGUUIDFromUUID(args.SectionID),
	}

	return ExecCommand(ctx, func() error {
		return s.Queries.UpdateProgress(ctx, sqlcArgs)
	})
}

func (s *Store) HasCompletedCourse(ctx context.Context, args domain.HasCompletedCourseParams) (bool, error) {
	sqlcArgs := sqlc.HasCompletedCourseParams{
		UserID:   args.UserID,
		CourseID: utils.PGUUIDFromUUID(args.CourseID),
	}

	completed, err := ExecQuery(ctx, func() (pgtype.Bool, error) {
		return s.Queries.HasCompletedCourse(ctx, sqlcArgs)
	})

	return completed.Bool, err
}

func (s *Store) ResetProgress(ctx context.Context, args domain.ResetProgressParams) error {
	sqlcArgs := sqlc.ResetProgressParams{
		UserID:   args.UserID,
		CourseID: utils.PGUUIDFromUUID(args.CourseID),
	}

	return ExecCommand(ctx, func() error {
		return s.Queries.ResetProgress(ctx, sqlcArgs)
	})
}

func (s *Store) SetCourseCompleted(ctx context.Context, args domain.SetCourseCompletedParams) error {
	sqlcArgs := sqlc.SetCourseCompletedParams{
		UserID:   args.UserID,
		CourseID: utils.PGUUIDFromUUID(args.CourseID),
	}

	return ExecCommand(ctx, func() error {
		return s.Queries.SetCourseCompleted(ctx, sqlcArgs)
	})
}

func progressFrom(row sqlc.GetProgressRow) *domain.Progress {
	var sectionUUIDs []uuid.UUID
	for _, sectionID := range row.CompletedSectionIds {
		sectionUUIDs = append(sectionUUIDs, uuid.UUID(sectionID.Bytes))
	}

	return &domain.Progress{
		CompletedSectionIDs: sectionUUIDs,
		CompletedIntro:      row.CompletedIntro.Bool,
	}
}

func (s *Store) GetAllProgress(ctx context.Context) ([]*domain.FullProgress, error) {
	progressRows, err := ExecQuery(
		ctx,
		func() ([]sqlc.GetAllProgressRow, error) { return s.Queries.GetAllProgress(ctx) },
	)
	if err != nil {
		return nil, err
	}

	courseSectionsMap := map[pgtype.UUID][]domain.CourseSectionProgress{}
	for i := range progressRows {
		if _, exists := courseSectionsMap[progressRows[i].CourseID]; !exists {
			sections, err := s.GetCourseSections(ctx, progressRows[i].CourseID)
			if err != nil {
				return nil, err
			}
			courseSectionsMap[progressRows[i].CourseID] = courseSectionProgressFrom(sections)
		}
	}

	progressByUser := map[UserDetails][]*domain.FullUserProgress{}

	for i := range progressRows {
		row := &progressRows[i]
		userDetails := UserDetails{
			ID:    row.UserID,
			Name:  row.UserName.String,
			Email: row.Email.String,
		}

		progressByUser[userDetails] = append(progressByUser[userDetails], fullUserProgressFrom(row, courseSectionsMap[row.CourseID]))
	}

	result := make([]*domain.FullProgress, 0, len(progressByUser))

	for user, progress := range progressByUser {
		result = append(result, &domain.FullProgress{
			UserID:   user.ID,
			UserName: user.Name,
			Email:    user.Email,
			Progress: progress,
		})
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].UserName < result[j].UserName
	})

	return result, nil
}

func fullUserProgressFrom(row *sqlc.GetAllProgressRow, courseSections []domain.CourseSectionProgress) *domain.FullUserProgress {
	var completedSectionIDs []uuid.UUID
	for _, sectionID := range row.CompletedSectionIds {
		completedSectionIDs = append(completedSectionIDs, uuid.UUID(sectionID.Bytes))
	}

	// Copy so we don't mutate the shared slice from the map
	sections := make([]domain.CourseSectionProgress, len(courseSections))
	copy(sections, courseSections)

	for i, section := range sections {
		if slices.Contains(completedSectionIDs, section.ID) {
			sections[i].Completed = true
		}
	}

	return &domain.FullUserProgress{
		CourseID:              uuid.UUID(row.CourseID.Bytes),
		CourseName:            row.CourseTitle.String,
		CourseSectionProgress: sections,
		CompletedIntro:        row.CompletedIntro.Bool,
		CompletedCourse:       row.CompletedCourse.Bool,
	}
}

func courseSectionProgressFrom(sections []domain.CourseSection) []domain.CourseSectionProgress {
	result := make([]domain.CourseSectionProgress, 0, len(sections))

	for _, s := range sections {
		title := s.GetTitle()
		result = append(result, domain.CourseSectionProgress{
			ID:    s.GetID(),
			Title: &title,
			Type:  string(s.GetType()),
		})
	}
	return result
}
