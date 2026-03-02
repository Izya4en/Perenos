-- =============================================
-- ПОЛНАЯ СХЕМА БАЗЫ ДАННЫХ (ATM ANALYTICS)
-- =============================================

-- Включаем поддержку гео-функций
CREATE EXTENSION IF NOT EXISTS postgis;

-- ---------------------------------------------
-- 1. СУЩНОСТИ
-- ---------------------------------------------

-- 1.1. Наши терминалы (Forte)
CREATE TABLE IF NOT EXISTS terminals (
    id SERIAL PRIMARY KEY,
    terminal_id VARCHAR(50) UNIQUE NOT NULL,
    model VARCHAR(100),
    address TEXT,
    city VARCHAR(50) DEFAULT 'Astana',
    location GEOMETRY(Point, 4326), -- Точка на карте
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 1.2. Конкуренты (Для графика "Доля рынка")
CREATE TABLE IF NOT EXISTS competitors (
    id SERIAL PRIMARY KEY,
    bank_name VARCHAR(100), -- Kaspi, Halyk, Jusan...
    address TEXT,
    location GEOMETRY(Point, 4326),
    created_at TIMESTAMP DEFAULT NOW()
);

-- 1.3. Дорожный граф и трафик (ИЗ ВАШЕГО CSV)
-- Заменили Полигоны на Линии (Edges), так как данные из 2ГИС/Geocash линейные
CREATE TABLE IF NOT EXISTS geo_traffic_edges (
    id SERIAL PRIMARY KEY,
    edge_id BIGINT UNIQUE,                 -- ID из файла (1.914e+16...), используем BIGINT!
    
    weekday_traffic INT DEFAULT 0,         -- Трафик в будни
    weekend_traffic INT DEFAULT 0,         -- Трафик в выходные
    
    hourly_weekday_traffic INT DEFAULT 0,      
    hourly_weekend_traffic INT DEFAULT 0,
    
    -- ВАЖНО: MultiLineString для дорог
    geometry GEOMETRY(MultiLineString, 4326), 
    
    source_data VARCHAR(50) DEFAULT '2GIS',
    created_at TIMESTAMP DEFAULT NOW()
);

-- ---------------------------------------------
-- 2. ОПЕРАЦИОННЫЕ ДАННЫЕ
-- ---------------------------------------------

-- 2.1. Финансовая статистика (ежедневная)
CREATE TABLE IF NOT EXISTS daily_stats (
    id BIGSERIAL PRIMARY KEY,
    terminal_id VARCHAR(50) REFERENCES terminals(terminal_id) ON DELETE CASCADE,
    report_date DATE NOT NULL,
    total_withdrawal_amount NUMERIC(15, 2) DEFAULT 0, -- Снятие
    total_deposit_amount NUMERIC(15, 2) DEFAULT 0,    -- Внесение
    transaction_count INT DEFAULT 0,
    UNIQUE(terminal_id, report_date)
);

-- 2.2. Жалобы и проблемы (Voice of Customer)
CREATE TABLE IF NOT EXISTS client_complaints (
    id BIGSERIAL PRIMARY KEY,
    terminal_id VARCHAR(50) REFERENCES terminals(terminal_id) ON DELETE SET NULL,
    complaint_category VARCHAR(50), -- DIRTY, EATS_CARD, QUEUE, BROKEN
    status VARCHAR(20) DEFAULT 'OPEN',
    user_location GEOMETRY(Point, 4326), -- Где стоял клиент, когда жаловался
    created_at TIMESTAMP DEFAULT NOW()
);

-- 2.3. Инкассация и мониторинг кассет
CREATE TABLE IF NOT EXISTS cash_levels (
    id BIGSERIAL PRIMARY KEY,
    terminal_id VARCHAR(50) REFERENCES terminals(terminal_id) ON DELETE CASCADE,
    check_time TIMESTAMP NOT NULL,
    current_balance NUMERIC(15, 2) NOT NULL,
    max_capacity NUMERIC(15, 2) NOT NULL,
    load_percentage DECIMAL(5, 2) -- % заполненности
);

-- ---------------------------------------------
-- 3. ИНДЕКСЫ (ДЛЯ СКОРОСТИ КАРТЫ)
-- ---------------------------------------------
CREATE INDEX idx_terminals_geom ON terminals USING GIST (location);
CREATE INDEX idx_competitors_geom ON competitors USING GIST (location);
CREATE INDEX idx_traffic_edges_geom ON geo_traffic_edges USING GIST (geometry);


-- ---------------------------------------------
-- 4. АНАЛИТИЧЕСКИЕ ПРЕДСТАВЛЕНИЯ (VIEWS)
-- ---------------------------------------------

-- VIEW 1: Сводка для Dashboard (Карта)
-- Собирает статус терминала: Эффективен/Проблемный
CREATE OR REPLACE VIEW view_dashboard_map AS
SELECT 
    t.terminal_id,
    t.model,
    t.address,
    t.location,
    COALESCE(ds.total_withdrawal_amount, 0) as total_cash_kzt,
    COALESCE(ds.transaction_count, 0) as tx_count,
    
    -- Логика статуса (KPI)
    CASE 
        WHEN (SELECT COUNT(*) FROM client_complaints c WHERE c.terminal_id = t.terminal_id AND c.status = 'OPEN') > 0 THEN 'Ineffective' -- Есть жалобы
        WHEN ds.transaction_count < 10 THEN 'Ineffective' -- Мало транзакций
        ELSE 'Effective'
    END as efficiency_status

FROM terminals t
LEFT JOIN daily_stats ds ON t.terminal_id = ds.terminal_id AND ds.report_date = (CURRENT_DATE - INTERVAL '1 day')::date;


-- VIEW 2: Рекомендации по расширению (Gap Analysis)
-- Ищет дороги с ВЫСОКИМ трафиком (> 5000), где рядом (50м) НЕТ наших терминалов.
CREATE OR REPLACE VIEW view_expansion_recommendations AS
SELECT 
    e.edge_id,
    e.weekday_traffic,
    e.geometry
FROM geo_traffic_edges e
WHERE 
    e.weekday_traffic > 5000 -- Фильтр: только "горячие" улицы
    AND NOT EXISTS (
        SELECT 1 
        FROM terminals t 
        WHERE ST_DWithin(
            e.geometry::geography, 
            t.location::geography, 
            50 -- Ищем в радиусе 50 метров от дороги
        )
    );

-- ---------------------------------------------
-- 5. ТЕСТОВЫЕ ДАННЫЕ (MOCK DATA)
-- Чтобы при запуске карта не была пустой
-- ---------------------------------------------

-- 5.1. Наши терминалы (Астана)
INSERT INTO terminals (terminal_id, model, address, location) VALUES
('FORTE-001', 'NCR SelfServ 6632', 'ТЦ Хан Шатыр', ST_SetSRID(ST_Point(71.4056, 51.1323), 4326)),
('FORTE-002', 'Diebold Opteva', 'Мангилик Ел, 55', ST_SetSRID(ST_Point(71.4138, 51.0885), 4326)),
('FORTE-003', 'NCR SelfServ 22', 'Вокзал Нурлы Жол', ST_SetSRID(ST_Point(71.5464, 51.1243), 4326))
ON CONFLICT (terminal_id) DO NOTHING;

-- 5.2. Конкуренты
INSERT INTO competitors (bank_name, address, location) VALUES
('Halyk Bank', 'ТЦ Хан Шатыр (1 этаж)', ST_SetSRID(ST_Point(71.4058, 51.1325), 4326)),
('Kaspi Bank', 'ТЦ Хан Шатыр (Вход)', ST_SetSRID(ST_Point(71.4055, 51.1320), 4326)),
('Jusan Bank', 'Мангилик Ел, 53', ST_SetSRID(ST_Point(71.4140, 51.0890), 4326));

-- 5.3. Статистика за вчера
INSERT INTO daily_stats (terminal_id, report_date, total_withdrawal_amount, transaction_count) VALUES
('FORTE-001', CURRENT_DATE - INTERVAL '1 day', 2500000.00, 150),
('FORTE-002', CURRENT_DATE - INTERVAL '1 day', 4800000.00, 320), -- Effective
('FORTE-003', CURRENT_DATE - INTERVAL '1 day', 150000.00, 5)     -- Ineffective
ON CONFLICT DO NOTHING;

-- 5.4. Одна жалоба (для теста статуса)
INSERT INTO client_complaints (terminal_id, complaint_category, user_location) VALUES
('FORTE-003', 'DIRTY', ST_SetSRID(ST_Point(71.5465, 51.1244), 4326));