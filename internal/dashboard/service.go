package dashboard

import (
	"context" // Нужно для передачи контекста в gRPC
	"geocash/internal/analytics"
	"geocash/internal/domain/terminal"
	"geocash/internal/platform/ai" // Импорт вашего пакета с gRPC клиентом
	"math/rand"
)

type Service struct {
	repo     terminal.Repository
	grid     *analytics.GridService
	aiClient *ai.Client // <-- 1. Добавлено поле клиента
}

// Обновляем конструктор, добавляем aiClient
func NewService(repo terminal.Repository, grid *analytics.GridService, aiClient *ai.Client) *Service {
	return &Service{
		repo:     repo,
		grid:     grid,
		aiClient: aiClient, // <-- Сохраняем клиента
	}
}

// Добавляем context.Context в аргументы, так как gRPC требует контекст
func (s *Service) GetDashboardData(ctx context.Context) DashboardResponse {
	// 1. Просим репозиторий создать 60 случайных точек в Астане
	rawAtms := s.repo.GenerateRandomCompetitors(60)

	var forte []terminal.ATM
	var competitors []terminal.ATM

	// 2. Проходим по точкам и распределяем их
	for i := range rawAtms {
		atm := &rawAtms[i]

		// С вероятностью 30% делаем банкомат НАШИМ (Forte)
		if rand.Float32() < 0.3 {
			s.repo.EnrichATM(atm)
			forte = append(forte, *atm)
		} else {
			s.repo.EnrichCompetitor(atm)
			competitors = append(competitors, *atm)
		}
	}

	// 3. Генерируем сетку (Heatmap)
	heatmap := s.grid.GenerateHexGrid()

	// 4. 🔥 ПОЛУЧАЕМ РЕКОМЕНДАЦИИ ЧЕРЕЗ gRPC
	// Координаты центра Астаны
	recommendations, err := s.aiClient.FetchRecommendations(ctx, 51.147, 71.430)
	if err != nil {
		// Логируем ошибку, но не роняем весь запрос
		// Если логгера нет, можно просто проигнорировать или вывести в консоль
		// fmt.Printf("Error fetching recommendations: %v\n", err)
		recommendations = []ai.Recommendation{} // Возвращаем пустой список
	}

	return DashboardResponse{
		Forte:           forte,
		Competitors:     competitors,
		HeatmapGrid:     heatmap,
		Recommendations: recommendations, // <-- Добавляем в ответ
	}
}
