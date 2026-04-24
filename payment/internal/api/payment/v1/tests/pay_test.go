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

	v1 "github.com/FOCCms/go-microservices-course/payment/internal/api/payment/v1"
	"github.com/FOCCms/go-microservices-course/payment/internal/api/payment/v1/mocks"
	errs "github.com/FOCCms/go-microservices-course/payment/internal/errors"
	"github.com/FOCCms/go-microservices-course/payment/internal/model"
	paymentv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/payment/v1"
)

func TestPayOrder(t *testing.T) {
	t.Parallel()

	var (
		ctx     = context.Background()
		orderID = gofakeit.UUID()
		txID    = gofakeit.UUID()
	)

	tests := []struct {
		name      string
		req       *paymentv1.PayOrderRequest
		setupMock func(svc *mocks.PaymentService)
		wantCode  codes.Code
	}{
		{
			name: "успешная оплата (OK)",
			req:  &paymentv1.PayOrderRequest{OrderUuid: orderID},
			setupMock: func(svc *mocks.PaymentService) {
				svc.EXPECT().
					Pay(ctx, mock.MatchedBy(func(r model.PayRequest) bool {
						return r.OrderUUID == orderID
					})).
					Return(txID, nil)
			},
			wantCode: codes.OK,
		},
		{
			name: "ошибка: неверный UUID (InvalidArgument)",
			req:  &paymentv1.PayOrderRequest{OrderUuid: "невалидный uuid"},
			setupMock: func(svc *mocks.PaymentService) {
				svc.EXPECT().
					Pay(ctx, mock.Anything).
					Return("", errs.ErrInvalidOrderUUID)
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "ошибка: внутренняя ошибка (Internal)",
			req:  &paymentv1.PayOrderRequest{OrderUuid: orderID},
			setupMock: func(svc *mocks.PaymentService) {
				svc.EXPECT().
					Pay(ctx, mock.Anything).
					Return("", errors.New("неизвестная ошибка"))
			},
			wantCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := mocks.NewPaymentService(t)
			tc.setupMock(mockSvc)

			a := v1.NewAPI(mockSvc)
			res, err := a.PayOrder(ctx, tc.req)

			if tc.wantCode == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, txID, res.TransactionUuid)
			} else {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tc.wantCode, st.Code())
			}
		})
	}
}
