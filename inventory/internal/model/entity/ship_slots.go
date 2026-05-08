package model

type ShipSlots struct {
	Hull   string
	Engine string
	Shield string
	Weapon string
}

func (s *ShipSlots) List() []string {
	l := []string{s.Hull, s.Engine}
	if s.Shield != "" {
		l = append(l, s.Shield)
	}
	if s.Weapon != "" {
		l = append(l, s.Weapon)
	}
	return l
}

type ResolvedShipSlots struct {
	Hull   *Part
	Engine *Part
	Shield *Part
	Weapon *Part
}

func (s *ResolvedShipSlots) List() []*Part {
	l := []*Part{s.Hull, s.Engine}
	if s.Shield != nil {
		l = append(l, s.Shield)
	}
	if s.Weapon != nil {
		l = append(l, s.Weapon)
	}
	return l
}

func (s *ResolvedShipSlots) Count() int {
	c := 0
	if s.Hull != nil {
		c++
	}
	if s.Engine != nil {
		c++
	}
	if s.Shield != nil {
		c++
	}
	if s.Weapon != nil {
		c++
	}
	return c
}
