from rest_framework.decorators import api_view # type: ignore
from rest_framework.response import Response # type: ignore
from rest_framework import status  # type: ignore
import time
import random
import requests # type: ignore
from concurrent import futures
executor = futures.ThreadPoolExecutor(max_workers=3)

HARDCODED_TOKEN = "super-secret-key-2025"

def calculate_revenue_formula(initial_sum, quarter):
    """
    Бизнес-логика: Расчет прогнозируемой выручки
    """
    try:
        val = float(initial_sum)
        q = int(quarter)
        bonus = random.random() * 1000
        result = val * (1 + q * 0.05) + bonus
        return round(result, 2)
    except (ValueError, TypeError):
        return 0.0

def process_forecast_task(data):
    app_id = data.get('id')
    initial_sum = data.get('initial_sum')
    quarter = data.get('quarter')
    
    incoming_token = data.get('verification_token')

    print(f"Start processing Forecast ID: {app_id}...")
    delay = random.randint(5, 10)
    time.sleep(delay)
    result_value = calculate_revenue_formula(initial_sum, quarter)
    result_data = {
        'id': app_id,
        'result': result_value,
        'verification_token': incoming_token 
    }

    return result_data

def send_revenue_forecast_to_go(result_data):
    go_url = "http://localhost:8081/api/cash-forecasts/update-result"
    
    try:
        requests.post(go_url, json=result_data, timeout=5)
        print(f"Sent result to Go successfully")
    except Exception as e:
        print(f"Error sending to Go: {e}")

def task_callback(task):
    try:
        result = task.result()
        send_revenue_forecast_to_go(result)
    except Exception as e:
        print(f"Task failed: {e}")

@api_view(['POST'])
def predict_revenue(request):
    """
    Эндпоинт: Предсказание выручки (бывший calculate)
    Ожидает в body: { "id": 1, "initial_sum": 1000, "verification_token": "..." }
    """
    data = request.data
    token = data.get('verification_token')
    if token != HARDCODED_TOKEN:
        return Response({'error': 'Invalid verification token inside body'}, status=status.HTTP_403_FORBIDDEN)

    task = executor.submit(process_forecast_task, data)
    task.add_done_callback(task_callback)

    return Response({'status': 'revenue_prediction_started'}, status=status.HTTP_200_OK)