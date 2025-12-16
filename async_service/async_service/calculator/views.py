from rest_framework.decorators import api_view # type: ignore
from rest_framework.response import Response # type: ignore
from rest_framework import status # type: ignore
import time
import random
import requests # type: ignore
from concurrent import futures

executor = futures.ThreadPoolExecutor(max_workers=3)

def calculate_formula(initial_sum, items):
    """
    Формула: Ок = Он + П – Р
    Он = initial_sum
    П  = сумма items с типом 'income'
    Р  = сумма items с типом 'expense'
    """
    try:
        on_value = float(initial_sum) # Он
        
        p_value = 0.0 # П
        r_value = 0.0 # Р

        # Пробегаем по счетам и распределяем суммы
        for item in items:
            amount = float(item.get('amount', 0))
            acc_type = item.get('type', 'expense') # Если типа нет, считаем расходом по дефолту

            if acc_type == 'income':
                p_value += amount
            else:
                # Все остальное считаем как Р (розничный товарооборот/расход)
                r_value += amount

        # Ок = Он + П – Р
        ok_result = on_value + p_value - r_value
        
        print(f"Debug calc: {on_value} + {p_value} - {r_value} = {ok_result}")
        return round(ok_result, 2)

    except (ValueError, TypeError) as e:
        print(f"Error in calculation: {e}")
        return 0.0

def process_application_task(data):
    app_id = data.get('id')
    initial_sum = data.get('initial_sum')
    items = data.get('items', []) # Список счетов

    print(f"Start processing Application ID: {app_id}...")

    # 1. Задержка 5-10 секунд (по заданию)
    delay = random.randint(5, 10)
    time.sleep(delay)

    # 2. Расчет по твоей формуле
    result_value = calculate_formula(initial_sum, items)
    
    # 3. Готовим ответ
    result_data = {
        'id': app_id,
        'result': result_value
    }

    return result_data

def send_result_to_go(result_data):
    # Адрес Go-бэкенда
    go_url = "http://localhost:8081/api/internal/result"
    try:
        requests.post(go_url, json=result_data, timeout=5)
    except Exception as e:
        print(f"Error sending to Go: {e}")

def task_callback(task):
    try:
        result = task.result()
        send_result_to_go(result)
    except Exception as e:
        print(f"Task failed: {e}")

@api_view(['POST'])
def calculate_forecast(request):
    """
    Принимает:
    {
        "id": 1, 
        "initial_sum": 1000, 
        "items": [
            {"amount": 500, "type": "income"},
            {"amount": 200, "type": "expense"}
        ]
    }
    """
    data = request.data
    task = executor.submit(process_application_task, data)
    task.add_done_callback(task_callback)
    return Response({'status': 'processing_started'}, status=status.HTTP_200_OK)