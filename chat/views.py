from django.shortcuts import render, redirect
from django.contrib.auth import login
from django.contrib.auth.decorators import login_required
from django.contrib.auth.forms import UserCreationForm
from django.http import JsonResponse
from django.utils import timezone
from .models import Wiadomosc


def rejestracja(request):
    if request.user.is_authenticated:
        return redirect("czat")
    if request.method == "POST":
        form = UserCreationForm(request.POST)
        if form.is_valid():
            user = form.save()
            login(request, user)
            return redirect("czat")
    else:
        form = UserCreationForm()
    return render(request, "chat/rejestracja.html", {"form": form})


@login_required
def czat(request):
    if request.method == "POST":
        tresc = request.POST.get("tresc", "").strip()
        if tresc:
            Wiadomosc.objects.create(autor=request.user, tresc=tresc[:1000])
        if request.headers.get("x-requested-with") == "XMLHttpRequest":
            return JsonResponse({"ok": True})
        return redirect("czat")
    wiadomosci = Wiadomosc.objects.select_related("autor").order_by("-data")[:50]
    wiadomosci = list(reversed(wiadomosci))
    return render(request, "chat/czat.html", {"wiadomosci": wiadomosci})


@login_required
def nowe_wiadomosci(request):
    try:
        od = int(request.GET.get("od", 0))
    except ValueError:
        od = 0
    lista = Wiadomosc.objects.select_related("autor").filter(id__gt=od)[:100]
    dane = []
    for w in lista:
        dane.append({
            "id": w.id,
            "autor": w.autor.username,
            "tresc": w.tresc,
            "data": timezone.localtime(w.data).strftime("%H:%M"),
            "moja": w.autor == request.user,
        })
    return JsonResponse({"wiadomosci": dane})
