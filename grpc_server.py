import grpc
from concurrent import futures
import sys
import os

# Добавляем путь, чтобы Python видел сгенерированные файлы
sys.path.append('./python_service') 

import recommendation_pb2
import recommendation_pb2_grpc
from src.inference_service import ATMInferenceService

# Инициализируем наш сервис с данными
# Укажите правильные пути к вашим CSV
service_logic = ATMInferenceService(
    traffic_path="data/foot_traffic.csv",
    atm_path="data/atm_with_coordinates.csv"
)

class RecommenderHandler(recommendation_pb2_grpc.RecommenderServicer):
    def GetRecommendations(self, request, context):
        # request содержит данные от Go (lat, lng, radius_km)
        # В текущей логике мы просто возвращаем топ лучших по всему городу,
        # но в будущем можно использовать request.lat/lng для фильтрации
        
        print(f"Received request via gRPC. Radius: {request.radius_km}")
        
        try:
            # Вызываем нашу логику
            data = service_logic.get_recommendations(top_n=10)
            
            # Собираем ответ в формате Protobuf
            response_locations = []
            for item in data:
                loc = recommendation_pb2.Location(
                    lat=item['lat'],
                    lng=item['lng'],
                    score=item['score'],
                    predicted_turnover=item['predicted_turnover'],
                    reason=item['reason']
                )
                response_locations.append(loc)
                
            return recommendation_pb2.Response(locations=response_locations)
            
        except Exception as e:
            print(f"Error calculating recommendations: {e}")
            context.set_details(str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            return recommendation_pb2.Response()

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    recommendation_pb2_grpc.add_RecommenderServicer_to_server(RecommenderHandler(), server)
    
    server.add_insecure_port('[::]:50051')
    print("🚀 Python gRPC Server started on port 50051")
    server.start()
    server.wait_for_termination()

if __name__ == '__main__':
    serve()