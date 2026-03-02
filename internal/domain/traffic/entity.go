package traffic

// TrafficSegment описывает одну строку из вашего CSV файла
type TrafficSegment struct {
	EdgeID         int64  // Было ID или ZoneName -> Стало EdgeID (BIGINT)
	WeekdayTraffic int    // Трафик в будни
	WeekendTraffic int    // Трафик в выходные
	Geometry       string // Геометрия в формате WKT (LINESTRING...)
}
