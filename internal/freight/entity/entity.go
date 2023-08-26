package entity

import (
	"github.com/ahugofreire/car-simulator-freight/common"
	"time"
)

type CustomTime time.Time

func (ct *CustomTime) UnMarshalJSON(b []byte) error {
	str := string(b)
	t, err := time.Parse(common.DateLayout, str[1:len(str)-1])
	if err != nil {
		return err
	}
	*ct = CustomTime(t)
	return nil
}

type RouteRepository interface {
	Create(route *Route) error
	Update(route *Route) error
	//	FindAll() ([]*Route, error)
	FindByID(ID string) (*Route, error)
}

type Route struct {
	ID           string
	Name         string
	Distance     float64
	Status       string
	FreightPrice float64
	StartedAt    time.Time
	FinishedAt   time.Time
}

func NewRoute(id, name string, distance float64) *Route {
	return &Route{
		ID:       id,
		Name:     name,
		Distance: distance,
		Status:   "pending",
	}
}

func (r *Route) Start(startedAt time.Time) {
	r.Status = "started"
	r.StartedAt = startedAt
}

func (r *Route) Finish(finishedAt time.Time) {
	r.Status = "finished"
	r.FinishedAt = finishedAt
}

type FreightInterface interface {
	Calculate(route *Route)
}

type Freight struct {
	PricePerKm float64
}

func NewFreight(pricePerKm float64) *Freight {
	return &Freight{
		PricePerKm: pricePerKm,
	}
}

func (f *Freight) Calculate(route *Route) {
	route.FreightPrice = route.Distance * f.PricePerKm
}
