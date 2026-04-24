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
	"github.com/FOCCms/go-microservices-course/inventory/internal/model"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
)

func TestGetPart(t *testing.T) {
	t.Parallel()

	var (
		ctx    = context.Background()
		partID = gofakeit.UUID()
	)

	tests := []struct {
		name      string
		req       *inventoryv1.GetPartRequest
		setupMock func(svc *mocks.PartService)
		wantCode  codes.Code
	}{
		{
			name: "успешное получение (OK)",
			req:  &inventoryv1.GetPartRequest{Uuid: partID},
			setupMock: func(svc *mocks.PartService) {
				svc.EXPECT().
					Get(ctx, partID).
					Return(model.Part{UUID: partID, Name: "Тестовый компонент"}, nil)
			},
			wantCode: codes.OK,
		},
		{
			name:      "ошибка: UUID не указан (InvalidArgument)",
			req:       &inventoryv1.GetPartRequest{Uuid: ""},
			setupMock: func(svc *mocks.PartService) {},
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "ошибка: деталь не найдена (NotFound)",
			req:  &inventoryv1.GetPartRequest{Uuid: partID},
			setupMock: func(svc *mocks.PartService) {
				svc.EXPECT().
					Get(ctx, partID).
					Return(model.Part{}, errs.ErrPartNotFound)
			},
			wantCode: codes.NotFound,
		},
		{
			name: "ошибка: внутренний сбой (Internal)",
			req:  &inventoryv1.GetPartRequest{Uuid: partID},
			setupMock: func(svc *mocks.PartService) {
				svc.EXPECT().
					Get(ctx, partID).
					Return(model.Part{}, errors.New("неизвестная ошибка"))
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
			res, err := a.GetPart(ctx, tc.req)

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
