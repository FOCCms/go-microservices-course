package iam

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	errs "github.com/FOCCms/go-microservices-course/iam/internal/errors"
	"github.com/FOCCms/go-microservices-course/iam/internal/model"
)

func (s *Service) Whoami(ctx context.Context, sessionUuid string) (model.Session, model.User, error) {
	ctx, span := otel.Tracer("iam-service").Start(ctx, "iam.Whoami")
	defer span.End()

	span.SetAttributes(attribute.String("iam.session_uuid", sessionUuid))

	if sessionUuid == "" {
		span.RecordError(errs.ErrEmptySessionID)
		span.SetStatus(codes.Error, errs.ErrEmptySessionID.Error())
		return model.Session{}, model.User{}, fmt.Errorf("получить сессию и пользователя: %w", errs.ErrEmptySessionID)
	}
	if err := uuid.Validate(sessionUuid); err != nil {
		span.RecordError(errs.ErrInvalidUUID)
		span.SetStatus(codes.Error, errs.ErrInvalidUUID.Error())
		return model.Session{}, model.User{}, fmt.Errorf("получить сессию и пользователя: %w", errs.ErrInvalidUUID)
	}

	session, err := s.sessionRepository.Get(ctx, sessionUuid)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return model.Session{}, model.User{}, fmt.Errorf("получить сессию и пользователя: %w", err)
	}

	user, err := s.userRepository.GetByUUID(ctx, session.UserUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return model.Session{}, model.User{}, fmt.Errorf("получить сессию и пользователя: %w", err)
	}

	span.SetAttributes(attribute.String("iam.user_uuid", user.UUID.String()))
	span.SetStatus(codes.Ok, "сессия и пользователь успешно получены")

	return session, user, nil
}
