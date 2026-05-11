package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "github.com/FOCCms/go-microservices-course/inventory/internal/api/inventory/v1"
	"github.com/FOCCms/go-microservices-course/inventory/internal/api/inventory/v1/mocks"
	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/entity"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
)

func TestListParts(t *testing.T) {
	t.Parallel()

	var (
		ctx      = context.Background()
		uuid1    = gofakeit.UUID()
		uuid2    = gofakeit.UUID()
		anyTime  = time.Now()
		anyUUID1 = uuid.New()
		anyUUID2 = uuid.New()
	)

	tests := []struct {
		name      string
		req       *inventoryv1.ListPartsRequest
		setupMock func(svc *mocks.PartService)
		wantCode  codes.Code
	}{
		{
			name: "успешный список (OK)",
			req:  &inventoryv1.ListPartsRequest{Uuids: []string{uuid1, uuid2}},
			setupMock: func(svc *mocks.PartService) {
				svc.EXPECT().
					List(ctx, mock.MatchedBy(func(f valueobject.PartFilter) bool {
						return len(f.UUIDs) == 2
					})).
					Return([]model.Part{
						model.RestorePart(anyUUID1.String(), "Деталь 1", "", "", 0, 0, 0, valueobject.PartProperties{}, anyTime),
						model.RestorePart(anyUUID2.String(), "Деталь 2", "", "", 0, 0, 0, valueobject.PartProperties{}, anyTime),
					}, nil)
			},
			wantCode: codes.OK,
		},
		{
			name: "ошибка: неверный формат UUID (InvalidArgument)",
			req:  &inventoryv1.ListPartsRequest{Uuids: []string{"невалидный uuid"}},
			setupMock: func(svc *mocks.PartService) {
				svc.EXPECT().List(ctx, mock.Anything).Return(nil, errs.ErrInvalidUUID)
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "ошибка: внутренняя (Internal)",
			req:  &inventoryv1.ListPartsRequest{Uuids: []string{uuid1}},
			setupMock: func(svc *mocks.PartService) {
				svc.EXPECT().List(ctx, mock.Anything).Return(nil, errors.New("неизвестная ошибка"))
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
			res, err := a.ListParts(ctx, tc.req)

			if tc.wantCode == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Len(t, res.Parts, 2)
			} else {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tc.wantCode, st.Code())
			}
		})
	}
}
