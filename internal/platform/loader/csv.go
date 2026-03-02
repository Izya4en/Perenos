package loader

import (
	"encoding/csv"
	"fmt"
	"geocash/internal/domain/traffic"
	"io"
	"os"
	"strconv"
)

// LoadTrafficCSV читает файл и возвращает массив структур
func LoadTrafficCSV(path string) ([]traffic.TrafficSegment, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)

	// Пропускаем заголовок
	if _, err := r.Read(); err != nil {
		return nil, err
	}

	var segments []traffic.TrafficSegment

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// 1. Парсим ID из научной нотации (1.91E+16)
		edgeIDFloat, err := strconv.ParseFloat(record[0], 64)
		if err != nil {
			continue // Пропускаем битые строки
		}

		// 2. Парсим трафик будни (Колонка 1)
		wd, _ := strconv.Atoi(record[1])

		// 3. Парсим трафик выходные (Колонка 3 - судя по вашему CSV)
		// Добавили это чтение:
		we, _ := strconv.Atoi(record[3])

		segments = append(segments, traffic.TrafficSegment{
			EdgeID:         int64(edgeIDFloat),
			WeekdayTraffic: wd,
			WeekendTraffic: we,        // <--- Не забудьте добавить это поле
			Geometry:       record[5], // Колонка 5 - геометрия
		})
	}

	return segments, nil
}
