from django.contrib import admin
from django.urls import path
from calculator.views import predict_revenue

urlpatterns = [
    path('admin/', admin.site.urls),
    path('api/predict-revenue', predict_revenue), 
]