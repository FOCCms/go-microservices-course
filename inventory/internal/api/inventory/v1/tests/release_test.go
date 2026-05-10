package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "github.com/FOCCms/go-microservices-course/inventory/internal/api/inventory/v1"
	"github.com/FOCCms/go-microservices-course/inventory/internal/api/inventory/v1/mocks"
	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
)

func TestReleaseParts(t *testing.T) {
	t.Parallel()

	var (
		ctx   = context.Background()
		uuids = []string{gofakeit.UUID(), gofakeit.UUID()}
	)

	tests := []struct {
		name      string
		req       *inventoryv1.ReleasePartsRequest
		setupMock func(svc *mocks.PartService)
		wantCode  codes.Code
	}{
		{
			name: "успешное освобождение (OK)",
			req:  &inventoryv1.ReleasePartsRequest{Uuids: uuids},
			setupMock: func(svc *mocks.PartService) {
				svc.EXPECT().
					Release(ctx, uuids).
					Return(nil)
			},
			wantCode: codes.OK,
		},
		{
			name: "ошибка: нечего освобождать (FailedPrecondition)",
			req:  &inventoryv1.ReleasePartsRequest{Uuids: uuids},
			setupMock: func(svc *mocks.PartService) {
				svc.EXPECT().
					Release(ctx, uuids).
					Return(errs.ErrNothingToRelease)
			},
			wantCode: codes.FailedPrecondition,
		},
		{
			name: "ошибка: деталь не найдена (NotFound)",
			req:  &inventoryv1.ReleasePartsRequest{Uuids: uuids},
			setupMock: func(svc *mocks.PartService) {
				svc.EXPECT().
					Release(ctx, uuids).
					Return(errs.ErrPartNotFound)
			},
			wantCode: codes.NotFound,
		},
		{
			name: "ошибка: внутренняя ошибка (Internal)",
			req:  &inventoryv1.ReleasePartsRequest{Uuids: uuids},
			setupMock: func(svc *mocks.PartService) {
				svc.EXPECT().
					Release(ctx, uuids).
					Return(errors.New("неизвестная ошибка"))
			},
			wantCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := mocks.NewPartService(t)
			tc.setupMock(mockSvc)

			a := v1.NewAPI(mockSvc)
			res, err := a.ReleaseParts(ctx, tc.req)

			if tc.wantCode == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			} else {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tc.wantCode, st.Code())
			}
		})
	}
}
