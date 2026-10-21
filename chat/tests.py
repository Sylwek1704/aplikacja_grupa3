from django.test import TestCase
from django.contrib.auth.models import User
from django.urls import reverse
from .models import Wiadomosc


class CzatTest(TestCase):
    def setUp(self):
        self.user = User.objects.create_user("jan", password="haslo12345")

    def test_bez_logowania_przekierowanie(self):
        r = self.client.get(reverse("czat"))
        self.assertEqual(r.status_code, 302)

    def test_wyslanie_wiadomosci(self):
        self.client.login(username="jan", password="haslo12345")
        self.client.post(reverse("czat"), {"tresc": "hej"})
        self.assertEqual(Wiadomosc.objects.count(), 1)

    def test_pobieranie_wiadomosci(self):
        Wiadomosc.objects.create(autor=self.user, tresc="test")
        self.client.login(username="jan", password="haslo12345")
        r = self.client.get(reverse("wiadomosci") + "?od=0")
        self.assertEqual(len(r.json()["wiadomosci"]), 1)
