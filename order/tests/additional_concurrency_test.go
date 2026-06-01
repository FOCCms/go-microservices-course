//go:build apitest

package tests

import (
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrent_DoublePay_ExactlyOneSucceeds(t *testing.T) {
	// Количество итераций для поимки гонки
	const numAttempts = 2
	// Количество параллельных запросов внутри одной попытки
	const numConcurrentRequests = 2

	for attempt := 1; attempt <= numAttempts; attempt++ {
		t.Run(fmt.Sprintf("Attempt_%d", attempt), func(t *testing.T) {
			createReq := &CreateOrderRequest{
				UserUUID:   uuid.New().String(),
				HullUUID:   HullAluminumUUID,
				EngineUUID: EngineIonCUUID,
			}
			createResult, createResp := createOrder(t, createReq)
			require.Equal(t, http.StatusCreated, createResp.StatusCode)
			_ = createResp.Body.Close()
			require.NotNil(t, createResult)

			orderUUID := createResult.OrderUUID

			var (
				wg       sync.WaitGroup
				mu       sync.Mutex
				statuses []int

				startSignal = make(chan struct{})
			)

			payReq := &PayOrderRequest{PaymentMethod: "CARD"}

			for range numConcurrentRequests {
				wg.Add(1)
				go func() {
					defer wg.Done()

					<-startSignal

					_, resp := payOrder(t, orderUUID, payReq)
					defer func() { _ = resp.Body.Close() }()

					mu.Lock()
					statuses = append(statuses, resp.StatusCode)
					mu.Unlock()
				}()
			}

			close(startSignal)
			wg.Wait()

			successCount := 0
			conflictCount := 0
			otherCount := 0

			for _, code := range statuses {
				switch code {
				case http.StatusOK:
					successCount++
				case http.StatusConflict:
					conflictCount++
				default:
					otherCount++
				}
			}

			if !assert.Equal(t, 1, successCount, "Попытка %d: Ровно ОДИН запрос на оплату должен вернуть 200 OK", attempt) {
				t.FailNow() // Сразу прекращаем тест, если поймали баг
			}

			assert.Equal(t, numConcurrentRequests-1, conflictCount, "Попытка %d: Остальные должны вернуть 409", attempt)
			assert.Equal(t, 0, otherCount, "Попытка %d: Других статусов быть не должно", attempt)

			finalOrder, getResp := getOrder(t, orderUUID)
			require.Equal(t, http.StatusOK, getResp.StatusCode)
			_ = getResp.Body.Close()

			assert.Equal(t, "PAID", finalOrder.Status, "Попытка %d: Итоговый статус должен быть PAID", attempt)
		})
	}
}

func TestConcurrent_DoubleCancel_ExactlyOneSucceeds(t *testing.T) {
	// Количество итераций для поимки гонки
	const numAttempts = 2
	// Количество параллельных запросов внутри одной попытки
	const numConcurrentRequests = 2

	for attempt := 1; attempt <= numAttempts; attempt++ {
		t.Run(fmt.Sprintf("Attempt_%d", attempt), func(t *testing.T) {
			createReq := &CreateOrderRequest{
				UserUUID:   uuid.New().String(),
				HullUUID:   HullAluminumUUID,
				EngineUUID: EngineIonCUUID,
			}
			createResult, createResp := createOrder(t, createReq)
			require.Equal(t, http.StatusCreated, createResp.StatusCode)
			_ = createResp.Body.Close()
			require.NotNil(t, createResult)

			orderUUID := createResult.OrderUUID

			var (
				wg          sync.WaitGroup
				mu          sync.Mutex
				statuses    []int
				startSignal = make(chan struct{})
			)

			for range numConcurrentRequests {
				wg.Add(1)
				go func() {
					defer wg.Done()

					<-startSignal

					_, resp := cancelOrder(t, orderUUID)
					defer func() { _ = resp.Body.Close() }()

					mu.Lock()
					statuses = append(statuses, resp.StatusCode)
					mu.Unlock()
				}()
			}

			close(startSignal)
			wg.Wait()

			successCount := 0
			conflictCount := 0
			otherCount := 0

			for _, code := range statuses {
				switch code {
				case http.StatusOK:
					successCount++
				case http.StatusConflict:
					conflictCount++
				default:
					otherCount++
				}
			}

			if !assert.Equal(t, 1, successCount, "Попытка %d: Ровно ОДИН запрос на отмену должен вернуть 200 OK", attempt) {
				t.FailNow()
			}

			assert.Equal(t, numConcurrentRequests-1, conflictCount, "Попытка %d: Остальные запросы должны вернуть 409 Conflict", attempt)
			assert.Equal(t, 0, otherCount, "Попытка %d: Других статус-кодов быть не должно", attempt)

			finalOrder, getResp := getOrder(t, orderUUID)
			require.Equal(t, http.StatusOK, getResp.StatusCode)
			_ = getResp.Body.Close()

			assert.Equal(t, "CANCELLED", finalOrder.Status, "Попытка %d: Итоговый статус заказа должен быть CANCELLED", attempt)
		})
	}
}

func TestConcurrent_DoubleCommitParts_ExactlyOneSucceeds(t *testing.T) {
	// Количество итераций для поимки гонки
	const numAttempts = 2
	// Количество параллельных запросов внутри одной попытки
	const numConcurrentRequests = 2

	for attempt := 1; attempt <= numAttempts; attempt++ {
		t.Run(fmt.Sprintf("Attempt_%d", attempt), func(t *testing.T) {
			createReq := &CreateOrderRequest{
				UserUUID:   uuid.New().String(),
				HullUUID:   HullAluminumUUID,
				EngineUUID: EngineIonCUUID,
			}
			createResult, createResp := createOrder(t, createReq)
			require.Equal(t, http.StatusCreated, createResp.StatusCode)
			_ = createResp.Body.Close()
			require.NotNil(t, createResult)

			orderUUID := createResult.OrderUUID

			var (
				wg          sync.WaitGroup
				mu          sync.Mutex
				statuses    []int
				startSignal = make(chan struct{})
			)

			for range numConcurrentRequests {
				wg.Add(1)
				go func() {
					defer wg.Done()

					<-startSignal

					_, resp := cancelOrder(t, orderUUID)
					defer func() { _ = resp.Body.Close() }()

					mu.Lock()
					statuses = append(statuses, resp.StatusCode)
					mu.Unlock()
				}()
			}

			close(startSignal)
			wg.Wait()

			successCount := 0
			conflictCount := 0
			otherCount := 0

			for _, code := range statuses {
				switch code {
				case http.StatusOK:
					successCount++
				case http.StatusConflict:
					conflictCount++
				default:
					otherCount++
				}
			}

			if !assert.Equal(t, 1, successCount, "Попытка %d: Ровно ОДИН запрос на отмену должен вернуть 200 OK", attempt) {
				t.FailNow()
			}

			assert.Equal(t, numConcurrentRequests-1, conflictCount, "Попытка %d: Остальные запросы должны вернуть 409 Conflict", attempt)
			assert.Equal(t, 0, otherCount, "Попытка %d: Других статус-кодов быть не должно", attempt)

			finalOrder, getResp := getOrder(t, orderUUID)
			require.Equal(t, http.StatusOK, getResp.StatusCode)
			_ = getResp.Body.Close()

			assert.Equal(t, "CANCELLED", finalOrder.Status, "Попытка %d: Итоговый статус заказа должен быть CANCELLED", attempt)
		})
	}
}
