package repository

import (
	"database/sql"
	"time"

	"github.com/ahugofreire/car-simulator-freight/internal/freight/entity"
)

type RouteRepositoryMySQL struct {
	db *sql.DB
}

func NewRouteRepositoryMysql(db *sql.DB) *RouteRepositoryMySQL {
	return &RouteRepositoryMySQL{
		db: db,
	}
}

func (r RouteRepositoryMySQL) Create(route *entity.Route) error {
	sql := "INSERT INTO routes (id, name, distance, status, freight_price) VALUES (?, ?, ?, ?, ?)"
	_, err := r.db.Exec(sql, route.ID, route.Name, route.Distance, route.Status, route.FreightPrice)
	if err != nil {
		return err
	}

	return nil
}

func (r RouteRepositoryMySQL) FindByID(ID string) (*entity.Route, error) {
	query := "SELECT id, name, distance, status, freight_price, started_at, finished_at FROM routes WHERE id = ?"
	row, err := r.db.Query(query, ID)
	if err != nil {
		return nil, err
	}

	var startedAt, finishedAt sql.NullTime
	var route entity.Route

	err = row.Scan(
		&route.ID,
		&route.Name,
		&route.Distance,
		&route.Status,
		&route.FreightPrice,
		&startedAt,
		&finishedAt,
	)
	if err != nil {
		return nil, err
	}

	if startedAt.Valid {
		route.StartedAt = startedAt.Time
	}

	if finishedAt.Valid {
		route.FinishedAt = finishedAt.Time
	}

	return &route, nil
}

func (r *RouteRepositoryMySQL) Update(route *entity.Route) error {
	startedAt := route.StartedAt.Format(time.RFC3339)
	finishedAt := route.FinishedAt.Format(time.RFC3339)
	query := "UPDATE routes SET status = ?, freight_price = ?, started_at = ?, finished_at = ? WHERE id = ?"
	_, err := r.db.Exec(query, route.Status, route.FreightPrice, startedAt, finishedAt, route.ID)
	if err != nil {
		return err
	}

	return nil
}
