from django.contrib import admin
from django.urls import path
from calculator.views import calculate_forecast

urlpatterns = [
    path('admin/', admin.site.urls),
    # Эндпоинт, который будет вызывать Go
    path('calculate/', calculate_forecast), 
]