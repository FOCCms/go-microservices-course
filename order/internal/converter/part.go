package converter

import "github.com/FOCCms/go-microservices-course/order/internal/model"

func ToShipSlots(r model.CreateOrderRequest) model.ShipSlots {
	s := model.ShipSlots{
		Hull:   r.HullUUID.String(),
		Engine: r.EngineUUID.String(),
	}
	if r.ShieldUUID != nil {
		s.Shield = r.ShieldUUID.String()
	}
	if r.WeaponUUID != nil {
		s.Weapon = r.WeaponUUID.String()
	}
	return s
}
