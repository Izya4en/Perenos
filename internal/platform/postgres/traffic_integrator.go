package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"geocash/internal/domain/traffic"
	"log"
)

type TrafficIntegrator struct {
	db *sql.DB
}

func NewTrafficIntegrator(db *sql.DB) *TrafficIntegrator {
	return &TrafficIntegrator{db: db}
}

func (t *TrafficIntegrator) EnrichZonesWithTraffic(ctx context.Context, segments []traffic.TrafficSegment) error {
	if len(segments) == 0 {
		return nil
	}

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query := `
		INSERT INTO geo_traffic_edges (edge_id, weekday_traffic, weekend_traffic, geometry)
		VALUES ($1, $2, $3, ST_Multi(ST_GeomFromText($4, 4326)))
		ON CONFLICT (edge_id) DO UPDATE SET
			weekday_traffic = EXCLUDED.weekday_traffic,
			weekend_traffic = EXCLUDED.weekend_traffic;
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("ошибка подготовки SQL запроса: %w", err)
	}
	defer stmt.Close()

	count := 0
	for _, segment := range segments {
		_, err := stmt.ExecContext(ctx,
			segment.EdgeID,
			segment.WeekdayTraffic,
			segment.WeekendTraffic,
			segment.Geometry,
		)

		if err != nil {
			log.Printf("⚠️ Ошибка вставки строки (EdgeID: %d): %v", segment.EdgeID, err)
			continue
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ошибка сохранения транзакции: %w", err)
	}

	log.Printf("✅ Успешно сохранено %d дорожных сегментов в базу", count)
	return nil
}
