package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"geocash/internal/domain/terminal"
	"geocash/internal/platform/loader"
	"geocash/internal/platform/postgres"

	// 👇 Импорт нашего AI клиента
	"geocash/internal/platform/ai"

	_ "github.com/lib/pq"
)

type TrafficEdgeResponse struct {
	ID             int64  `json:"id"`
	WeekdayTraffic int    `json:"weekday_traffic"`
	GeometryJSON   string `json:"geometry"`
}

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	connStr := fmt.Sprintf("postgres://postgres:secret@%s:5432/atm_db?sslmode=disable", dbHost)

	fmt.Println("🔌 Подключаемся к Postgres:", connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("❌ Ошибка драйвера: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("❌ БД недоступна: %v", err)
	}
	fmt.Println("✅ Успешное подключение к БД!")

	// 1. ИНИЦИАЛИЗАЦИЯ AI КЛИЕНТА (gRPC)
	aiAddr := getEnv("AI_SERVICE_ADDR", "localhost:50051")
	fmt.Println("🤖 Подключаемся к AI Service:", aiAddr)

	aiClient, err := ai.NewClient(aiAddr)
	if err != nil {
		log.Printf("⚠️ WARNING: AI Service недоступен (%v). Работаем без рекомендаций.", err)
	} else {
		defer aiClient.Close()
		fmt.Println("✅ AI Client подключен!")
	}

	// 2. ПЕРЕСОЗДАЕМ ТАБЛИЦУ И ЗАЛИВАЕМ ВСЕ ДАННЫЕ
	ensureSchema(db)

	csvPath := "./traffic_data.csv"
	if _, err := os.Stat(csvPath); err == nil {
		data, err := loader.LoadTrafficCSV(csvPath)
		if err == nil && len(data) > 0 {
			go func() {
				integrator := postgres.NewTrafficIntegrator(db)
				_ = integrator.EnrichZonesWithTraffic(context.Background(), data)
				fmt.Printf("📊 Загружено %d сегментов трафика.\n", len(data))
			}()
		}
	}

	repo := postgres.NewRepository(db)

	http.HandleFunc("/api/dashboard", func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		if r.Method == "OPTIONS" {
			return
		}

		ctx := r.Context()
		atms, err := repo.GetAll(ctx)
		if err != nil {
			log.Printf("⚠️ Ошибка получения банкоматов: %v", err)
			atms = []terminal.ATM{}
		}

		forte := []terminal.ATM{}
		competitors := []terminal.ATM{}

		for _, a := range atms {
			if a.IsForte {
				forte = append(forte, a)
			} else {
				competitors = append(competitors, a)
			}
		}

		// Трафик (GeoJSON)
		rows, err := db.QueryContext(ctx, `
            SELECT edge_id, weekday_traffic, ST_AsGeoJSON(geometry) 
            FROM geo_traffic_edges 
            WHERE geometry IS NOT NULL
            LIMIT 5000;
        `)

		trafficData := []TrafficEdgeResponse{}
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var t TrafficEdgeResponse
				if err := rows.Scan(&t.ID, &t.WeekdayTraffic, &t.GeometryJSON); err == nil {
					trafficData = append(trafficData, t)
				}
			}
		}

		// 🔥 ПОЛУЧЕНИЕ РЕКОМЕНДАЦИЙ ОТ PYTHON (gRPC)
		var recommendations []ai.Recommendation
		if aiClient != nil {
			// Запрашиваем рекомендации для центра Астаны
			recs, err := aiClient.FetchRecommendations(ctx, 51.147, 71.430)
			if err == nil {
				recommendations = recs
				// 👇 ЛОГ УСПЕХА
				fmt.Printf("✅ [gRPC УСПЕХ] Получено %d точек от Python-сервиса!\n", len(recs))
			} else {
				// 👇 ЛОГ ОШИБКИ
				fmt.Printf("❌ [gRPC ОШИБКА] Не удалось получить данные: %v\n", err)
				recommendations = []ai.Recommendation{}
			}
		} else {
			// 👇 ЛОГ ОТСУТСТВИЯ КЛИЕНТА
			fmt.Println("⚠️ [gRPC ПРЕДУПРЕЖДЕНИЕ] Клиент AI не был инициализирован при старте.")
			recommendations = []ai.Recommendation{}
		}

		response := map[string]interface{}{
			"forte":           forte,
			"competitors":     competitors,
			"traffic":         trafficData,
			"recommendations": recommendations,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		w.Write([]byte("✅ GeoCash Backend is Running!"))
	})

	fmt.Println("🚀 Сервер запущен: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func ensureSchema(db *sql.DB) {
	fmt.Println("🧹 Полная пересборка базы данных...")
	db.Exec("DROP TABLE IF EXISTS terminals CASCADE;")

	query := `
        CREATE EXTENSION IF NOT EXISTS postgis;
        CREATE TABLE terminals (
            id SERIAL PRIMARY KEY,
            terminal_id TEXT UNIQUE,
            bank TEXT,
            lat DOUBLE PRECISION,
            lng DOUBLE PRECISION,
            total_cash_kzt DOUBLE PRECISION DEFAULT 0,
            efficiency_status TEXT DEFAULT 'Effective',
            address TEXT,
            complaints TEXT
        );
    `
	db.Exec(query)

	dataQuery := `
        TRUNCATE TABLE terminals RESTART IDENTITY;

        -- 1. 🟢 ВАШИ РЕАЛЬНЫЕ ТОЧКИ (FORTE - Якоря)
        INSERT INTO terminals (terminal_id, bank, lat, lng, total_cash_kzt, efficiency_status, address, complaints) VALUES
        ('FRT-REAL-01', 'ForteBank', 51.1284, 71.4305, 12500000, 'Effective', 'ул. Достык, 8/1 (БЦ Москва)', '[]'),
        ('FRT-REAL-02', 'ForteBank', 51.0888, 71.4138, 18000000, 'Effective', 'MEGA Silk Way (Вход А)', '[]'),
        ('FRT-REAL-03', 'ForteBank', 51.1325, 71.4056, 9400000,  'Effective', 'ТРЦ Хан Шатыр (1 этаж)', '[]'),
        ('FRT-REAL-04', 'ForteBank', 51.1442, 71.4183, 7200000,  'Effective', 'KeruenCity (у кинотеатра)', '[]'),
        ('FRT-REAL-05', 'ForteBank', 51.1557, 71.4682, 11000000, 'Effective', 'Forte Kulanshi Art', '[]'),
        ('FRT-REAL-06', 'ForteBank', 51.1890, 71.4132, 9200000,  'Effective', 'ЖД Вокзал (Старый)', '[]');

        -- 2. 🔴 ВАШИ РЕАЛЬНЫЕ ПРОБЛЕМНЫЕ (Для демо жалоб)
        INSERT INTO terminals (terminal_id, bank, lat, lng, total_cash_kzt, efficiency_status, address, complaints) VALUES
        ('FRT-BAD-01', 'ForteBank', 51.1150, 71.4550, 150000, 'Ineffective', 'Магазин у дома (Алматы 13)', 
         '[{"text": "Грязный экран", "date": "2023-10-25", "category": "DIRTY"}]'),
        ('FRT-BAD-02', 'ForteBank', 51.1750, 71.3950, 0, 'Ineffective', 'ул. А. Молдагуловой, 22', 
         '[{"text": "Не работает 2 дня", "date": "2023-10-26", "category": "BROKEN"}]');

        -- 3. 🤖 ГЕНЕРАЦИЯ МАССОВКИ: 300 КОНКУРЕНТОВ (Kaspi, Halyk, Jusan, BCC)
        INSERT INTO terminals (terminal_id, bank, lat, lng, total_cash_kzt, efficiency_status, address, complaints)
        SELECT 
            'COMP-GEN-' || generate_series,
            (ARRAY['Kaspi', 'Kaspi', 'Kaspi', 'Halyk Bank', 'Halyk Bank', 'Jusan', 'BCC'])[floor(random() * 7 + 1)],
            51.08 + random() * (51.18 - 51.08),
            71.38 + random() * (71.52 - 71.38),
            0,
            'Normal',
            'Сгенерированный адрес #' || generate_series,
            '[]'
        FROM generate_series(1, 300);

        -- 4. 🤖 ГЕНЕРАЦИЯ МАССОВКИ: 50 ДОПОЛНИТЕЛЬНЫХ FORTE (Эффективных)
        INSERT INTO terminals (terminal_id, bank, lat, lng, total_cash_kzt, efficiency_status, address, complaints)
        SELECT 
            'FRT-GEN-' || generate_series,
            'ForteBank',
            51.09 + random() * (51.17 - 51.09),
            71.39 + random() * (71.50 - 71.39),
            (random() * 10000000 + 2000000)::int,
            'Effective',
            'Отделение Forte #' || generate_series,
            '[]'
        FROM generate_series(1, 50);

        -- 5. 🤖 ГЕНЕРАЦИЯ МАССОВКИ: 20 СЛОМАННЫХ FORTE (Разбросаны по окраинам)
        INSERT INTO terminals (terminal_id, bank, lat, lng, total_cash_kzt, efficiency_status, address, complaints)
        SELECT 
            'FRT-BAD-GEN-' || generate_series,
            'ForteBank',
            51.05 + random() * (51.22 - 51.05),
            71.35 + random() * (71.55 - 71.35),
            (random() * 500000)::int,
            'Ineffective',
            'Удаленная точка #' || generate_series,
            '[{"text": "Сгенерированная жалоба на обслуживание", "date": "2023-11-01", "category": "SERVICE"}]'
        FROM generate_series(1, 20);
    `

	_, err := db.Exec(dataQuery)
	if err != nil {
		fmt.Println("❌ Ошибка при наполнении базы:", err)
	} else {
		fmt.Println("✅ УСПЕШНО: База наполнена! (Всего ~380 терминалов)")
	}
}
