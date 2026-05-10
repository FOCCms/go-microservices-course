package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	model "github.com/FOCCms/go-microservices-course/inventory/internal/model/entity"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
)

func TestCompatibilityChecker_Check(t *testing.T) {
	t.Parallel()

	checker := NewCompatibilityChecker()
	anyTime := time.Now()

	// Вспомогательные функции для создания валидных пропсов
	hullProps, _ := valueobject.NewHullProperties(100)         // Крепкий корпус
	weakHullProps, _ := valueobject.NewHullProperties(30)      // Хлипкий корпус
	engineProps, _ := valueobject.NewEngineProperties("A", 50) // Двигатель, требующий 50 прочности

	shieldPlasma, _ := valueobject.NewShieldProperties(valueobject.ShieldTypePlasma)
	weaponLaser, _ := valueobject.NewWeaponProperties(valueobject.WeaponTypeLaser)
	weaponMissile, _ := valueobject.NewWeaponProperties(valueobject.WeaponTypeMissile)

	tests := []struct {
		name    string
		slots   model.ResolvedShipSlots
		wantErr error
	}{
		{
			name: "успех: одна деталь не проверяется",
			slots: model.ResolvedShipSlots{
				Hull: new(model.RestorePart(uuid.NewString(), "Hull", "", valueobject.PartTypeHull, 100, 10, 0, hullProps, anyTime)),
			},
			wantErr: nil,
		},
		{
			name: "успех: корпус и двигатель совместимы",
			slots: model.ResolvedShipSlots{
				Hull:   new(model.RestorePart(uuid.NewString(), "Hull", "", valueobject.PartTypeHull, 100, 10, 0, hullProps, anyTime)),
				Engine: new(model.RestorePart(uuid.NewString(), "Eng", "", valueobject.PartTypeEngine, 100, 10, 0, engineProps, anyTime)),
			},
			wantErr: nil,
		},
		{
			name: "ошибка: корпус слишком слабый для двигателя",
			slots: model.ResolvedShipSlots{
				Hull:   new(model.RestorePart(uuid.NewString(), "Weak Hull", "", valueobject.PartTypeHull, 100, 10, 0, weakHullProps, anyTime)),
				Engine: new(model.RestorePart(uuid.NewString(), "Heavy Eng", "", valueobject.PartTypeEngine, 100, 10, 0, engineProps, anyTime)),
			},
			wantErr: errs.ErrIncompatibleParts,
		},
		{
			name: "ошибка: плазменный щит конфликтует с лазером",
			slots: model.ResolvedShipSlots{
				Shield: new(model.RestorePart(uuid.NewString(), "Plasm", "", valueobject.PartTypeShield, 100, 10, 0, shieldPlasma, anyTime)),
				Weapon: new(model.RestorePart(uuid.NewString(), "Laser", "", valueobject.PartTypeWeapon, 100, 10, 0, weaponLaser, anyTime)),
			},
			wantErr: errs.ErrIncompatibleParts,
		},
		{
			name: "успех: плазменный щит и ракеты ок",
			slots: model.ResolvedShipSlots{
				Shield: new(model.RestorePart(uuid.NewString(), "Plasm", "", valueobject.PartTypeShield, 100, 10, 0, shieldPlasma, anyTime)),
				Weapon: new(model.RestorePart(uuid.NewString(), "Missile", "", valueobject.PartTypeWeapon, 100, 10, 0, weaponMissile, anyTime)),
			},
			wantErr: nil,
		},
		{
			name: "ошибка: две детали одного типа (два корпуса)",
			slots: func() model.ResolvedShipSlots {
				p := model.RestorePart(uuid.NewString(), "Hull", "", valueobject.PartTypeHull, 100, 10, 0, hullProps, anyTime)
				return model.ResolvedShipSlots{
					Hull:   &p,
					Engine: &p,
				}
			}(),
			wantErr: errs.ErrIncompatibleParts,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := checker.Check(tc.slots)

			if tc.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
