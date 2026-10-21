from django.db import models
from django.contrib.auth.models import User


class Wiadomosc(models.Model):
    autor = models.ForeignKey(User, on_delete=models.CASCADE)
    tresc = models.TextField(max_length=1000)
    data = models.DateTimeField(auto_now_add=True)

    class Meta:
        ordering = ["data"]

    def __str__(self):
        return f"{self.autor.username}: {self.tresc[:30]}"
