package ai

import (
	"context"
	"time"

	// Импорт сгенерированного кода (путь зависит от вашего go.mod)
	pb "geocash/internal/gen/recommendation"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client — структура нашего клиента
type Client struct {
	grpcClient pb.RecommenderClient
	conn       *grpc.ClientConn
}

// NewClient — конструктор
func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:       conn,
		grpcClient: pb.NewRecommenderClient(conn),
	}, nil
}

// Close — закрытие соединения (важно для graceful shutdown)
func (c *Client) Close() error {
	return c.conn.Close()
}

// Структура для возврата данных в Dashboard (чтобы не тянуть pb зависимость)
type Recommendation struct {
	Lat      float64
	Lng      float64
	Score    int
	Turnover float64
	Reason   string
}

// FetchRecommendations — основной метод
func (c *Client) FetchRecommendations(ctx context.Context, lat, lng float64) ([]Recommendation, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &pb.Request{
		Lat:      lat,
		Lng:      lng,
		RadiusKm: 5,
	}

	resp, err := c.grpcClient.GetRecommendations(ctx, req)
	if err != nil {
		return nil, err
	}

	// Маппинг из Proto формата в наш внутренний формат
	var result []Recommendation
	for _, loc := range resp.Locations {
		result = append(result, Recommendation{
			Lat:      loc.Lat,
			Lng:      loc.Lng,
			Score:    int(loc.Score),
			Turnover: loc.PredictedTurnover,
			Reason:   loc.Reason,
		})
	}
	return result, nil
}
