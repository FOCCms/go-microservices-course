package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "github.com/FOCCms/go-microservices-course/inventory/internal/api/inventory/v1"
	"github.com/FOCCms/go-microservices-course/inventory/internal/api/inventory/v1/mocks"
	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
)

func TestListParts(t *testing.T) {
	t.Parallel()

	var (
		ctx   = context.Background()
		uuid1 = gofakeit.UUID()
		uuid2 = gofakeit.UUID()
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
					List(ctx, mock.MatchedBy(func(f model.PartFilter) bool {
						return len(f.UUIDs) == 2
					})).
					Return([]model.Part{{Name: "Деталь 1"}, {Name: "Деталь 2"}}, nil)
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
