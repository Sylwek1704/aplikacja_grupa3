var lista = document.getElementById("wiadomosci");
var formularz = document.getElementById("formularz");
var pole = document.getElementById("tresc");

function ostatnieId() {
    var wszystkie = lista.querySelectorAll(".wiadomosc");
    if (wszystkie.length == 0) {
        return 0;
    }
    return wszystkie[wszystkie.length - 1].dataset.id;
}

function przewin() {
    lista.scrollTop = lista.scrollHeight;
}

function dodajWiadomosc(w) {
    var pusto = document.getElementById("pusto");
    if (pusto) {
        pusto.remove();
    }
    var div = document.createElement("div");
    div.className = "wiadomosc";
    if (w.moja) {
        div.className += " moja";
    }
    div.dataset.id = w.id;

    var autor = document.createElement("span");
    autor.className = "autor";
    autor.textContent = w.autor;

    var godzina = document.createElement("span");
    godzina.className = "godzina";
    godzina.textContent = w.data;

    var tresc = document.createElement("div");
    tresc.className = "tresc";
    tresc.textContent = w.tresc;

    div.appendChild(autor);
    div.appendChild(godzina);
    div.appendChild(tresc);
    lista.appendChild(div);
}

function pobierz() {
    fetch(URL_WIADOMOSCI + "?od=" + ostatnieId())
        .then(function (res) {
            if (res.redirected) {
                window.location.reload();
            }
            return res.json();
        })
        .then(function (dane) {
            if (dane.wiadomosci.length > 0) {
                dane.wiadomosci.forEach(dodajWiadomosc);
                przewin();
            }
        })
        .catch(function (err) {
            console.log(err);
        });
}

formularz.addEventListener("submit", function (e) {
    e.preventDefault();
    var dane = new FormData(formularz);
    pole.value = "";
    fetch(formularz.action || window.location.href, {
        method: "POST",
        body: dane,
        headers: { "X-Requested-With": "XMLHttpRequest" }
    }).then(function () {
        pobierz();
    });
});

przewin();
setInterval(pobierz, 2000);
