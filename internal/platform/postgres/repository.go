package postgres

import (
	"context"
	"database/sql"
	"geocash/internal/domain/terminal"
	"log"
	"strings"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll(ctx context.Context) ([]terminal.ATM, error) {
	// 1. ЗАПРОС (9 колонок)
	query := `
		SELECT id, terminal_id, bank, lat, lng, 
			   COALESCE(total_cash_kzt, 0), 
			   COALESCE(efficiency_status, 'Normal'),
			   COALESCE(address, ''),        
			   COALESCE(complaints, '[]')    
		FROM terminals`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("❌ ОШИБКА SQL ЗАПРОСА: %v", err)
		return nil, err
	}
	defer rows.Close()

	var atms []terminal.ATM
	for rows.Next() {
		var a terminal.ATM

		// 2. СКАНИРОВАНИЕ (СТРОГО 9 ПЕРЕМЕННЫХ)
		if err := rows.Scan(
			&a.ID,
			&a.Name,
			&a.Bank,
			&a.Lat,
			&a.Lng,
			&a.TotalCashKZT,
			&a.EfficiencyStatus,
			&a.Address,    // <-- Адрес
			&a.Complaints, // <-- Жалобы
		); err != nil {
			// 👇 ТЕПЕРЬ МЫ УВИДИМ, ПОЧЕМУ ОНО НЕ РАБОТАЕТ
			log.Printf("❌ ОШИБКА ЧТЕНИЯ СТРОКИ: %v", err)
			continue
		}

		a.IsForte = strings.Contains(strings.ToLower(a.Bank), "forte")
		atms = append(atms, a)
	}

	// Если список пустой, пишем в лог
	if len(atms) == 0 {
		log.Println("⚠️ ВНИМАНИЕ: Запрос вернул 0 банкоматов. Таблица пустая?")
	}

	return atms, nil
}

// Заглушки
func (r *Repository) EnrichATM(atm *terminal.ATM)                        { atm.IsForte = true }
func (r *Repository) EnrichCompetitor(atm *terminal.ATM)                 { atm.IsForte = false }
func (r *Repository) GenerateRandomCompetitors(count int) []terminal.ATM { return []terminal.ATM{} }
