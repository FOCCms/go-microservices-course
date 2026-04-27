package payment

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	errs "github.com/FOCCms/go-microservices-course/payment/internal/errors"
	"github.com/FOCCms/go-microservices-course/payment/internal/model"
)

func TestPay(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name    string
		req     model.PayRequest
		wantErr error
	}{
		{
			name: "успешная оплата",
			req: model.PayRequest{
				OrderUUID:     uuid.NewString(),
				PaymentMethod: model.PaymentMethodCard,
			},
			wantErr: nil,
		},
		{
			name: "ошибка: неверный UUID",
			req: model.PayRequest{
				OrderUUID:     "invalid-uuid",
				PaymentMethod: model.PaymentMethodCard,
			},
			wantErr: errs.ErrInvalidOrderUUID,
		},
		{
			name: "ошибка: неверный метод оплаты",
			req: model.PayRequest{
				OrderUUID:     uuid.NewString(),
				PaymentMethod: model.PaymentMethodUnspecified,
			},
			wantErr: errs.ErrInvalidPaymentMethod,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := &service{}
			txID, err := svc.Pay(ctx, tc.req)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Empty(t, txID)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, txID)
				_, parseErr := uuid.Parse(txID)
				assert.NoError(t, parseErr)
			}
		})
	}
}
