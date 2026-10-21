from django.urls import path
from django.contrib.auth import views as auth_views
from . import views

urlpatterns = [
    path("", views.czat, name="czat"),
    path("wiadomosci/", views.nowe_wiadomosci, name="wiadomosci"),
    path("login/", auth_views.LoginView.as_view(template_name="chat/login.html", redirect_authenticated_user=True), name="login"),
    path("logout/", auth_views.LogoutView.as_view(), name="logout"),
    path("rejestracja/", views.rejestracja, name="rejestracja"),
]
