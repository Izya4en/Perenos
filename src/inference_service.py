import geopandas as gpd
from src.data_loader import load_foot_traffic_data, load_atm_addresses
from src.feature_engineering import create_features
from src.grid_aggregation import aggregate_to_grid
from src.scoring_model import calculate_atm_score

class ATMInferenceService:
    def __init__(self, traffic_path: str, atm_path: str):
        print("Initializing Service: Loading data...")
        
        # 1. Загружаем данные ОДИН РАЗ при старте
        self.raw_traffic = load_foot_traffic_data(traffic_path)
        self.atm_points = load_atm_addresses(atm_path)
        
        # 2. Предобработка (Features + Grid)
        # Если данные меняются редко, сетку лучше построить сразу
        print("Building Grid...")
        processed_traffic = create_features(self.raw_traffic)
        self.grid = aggregate_to_grid(processed_traffic, grid_size=200)
        
        # Фильтруем слабые зоны сразу, чтобы ускорить поиск
        self.grid = self.grid[self.grid["total_traffic"] > 50].copy()
        print("Service Ready!")

    def get_recommendations(self, top_n: int = 10) -> list:
        """
        Метод вызывается при каждом gRPC запросе.
        Он быстрый, так как сетка уже готова.
        """
        
        # Пересчитываем скоринг (вдруг логика скоринга зависит от параметров)
        # Если банкоматы меняются редко, calculate_atm_score тоже можно вынести в __init__
        recommended_zones = calculate_atm_score(
            grid=self.grid,
            atm_points=self.atm_points,
            min_distance=150,
            top_n=top_n,
            mutual_exclusion_radius=400
        )
        
        # Конвертируем GeoDataFrame в список словарей для gRPC
        results = []
        for _, row in recommended_zones.iterrows():
            geom = row.geometry.centroid
            results.append({
                "lat": geom.y,
                "lng": geom.x,
                "score": int(row["atm_score"] * 100), # 0.94 -> 94
                "predicted_turnover": float(row["total_traffic"] * 1500), # Примерная формула оборота
                "reason": f"High traffic zone (Density: {round(row['traffic_density'], 2)})"
            })
            
        return results