var lista = document.getElementById("wiadomosci");
var formularz = document.getElementById("formularz");
var pole = document.getElementById("tresc");
var plik = document.getElementById("zdjecie");
var wybrane = document.getElementById("wybrane");
var przycisk = document.getElementById("wyslij");
var ostatnie = "";
var przewinNaDol = true;

function przewin() {
    lista.scrollTop = lista.scrollHeight;
}

function narysuj(wiadomosci) {
    var naDole = przewinNaDol || lista.scrollHeight - lista.scrollTop - lista.clientHeight < 50;
    przewinNaDol = false;
    lista.innerHTML = "";

    if (wiadomosci.length == 0) {
        var p = document.createElement("p");
        p.id = "pusto";
        p.textContent = "Brak wiadomości, napisz coś pierwszy!";
        lista.appendChild(p);
        return;
    }

    wiadomosci.forEach(function (w) {
        var div = document.createElement("div");
        div.className = "wiadomosc";
        if (w.moja) {
            div.className += " moja";
        }

        var autor = document.createElement("span");
        autor.className = "autor";
        autor.textContent = w.autor;
        div.appendChild(autor);

        var godzina = document.createElement("span");
        godzina.className = "godzina";
        godzina.textContent = w.godzina + (w.edytowana ? " (edytowana)" : "");
        div.appendChild(godzina);

        if (w.tresc) {
            var tresc = document.createElement("div");
            tresc.className = "tresc";
            tresc.textContent = w.tresc;
            div.appendChild(tresc);
        }

        if (w.zdjecie) {
            var link = document.createElement("a");
            link.href = w.zdjecie;
            link.target = "_blank";
            var img = document.createElement("img");
            img.src = w.zdjecie;
            img.className = "zdjecie";
            img.alt = "zdjęcie";
            img.onload = function () {
                if (naDole) {
                    przewin();
                }
            };
            link.appendChild(img);
            div.appendChild(link);
        }

        lista.appendChild(div);
    });

    if (naDole) {
        przewin();
    }
}

function pobierz() {
    fetch("/api/wiadomosci")
        .then(function (res) {
            if (res.status == 401) {
                window.location.href = "/login";
                return null;
            }
            return res.json();
        })
        .then(function (dane) {
            if (!dane) {
                return;
            }
            var tekst = JSON.stringify(dane);
            if (tekst != ostatnie) {
                ostatnie = tekst;
                narysuj(dane);
            }
        })
        .catch(function (err) {
            console.log(err);
        });
}

plik.addEventListener("change", function () {
    if (plik.files.length > 0) {
        wybrane.textContent = "Wybrane zdjęcie: " + plik.files[0].name;
    } else {
        wybrane.textContent = "";
    }
});

formularz.addEventListener("submit", function (e) {
    e.preventDefault();
    if (pole.value.trim() == "" && plik.files.length == 0) {
        return;
    }
    if (plik.files.length > 0 && plik.files[0].size > 5 * 1024 * 1024) {
        alert("Zdjęcie może mieć maksymalnie 5 MB");
        return;
    }

    var dane = new FormData(formularz);
    przycisk.disabled = true;
    fetch("/api/wiadomosci", { method: "POST", body: dane })
        .then(function (res) {
            if (!res.ok) {
                return res.text().then(function (tekst) {
                    alert(tekst);
                });
            }
            formularz.reset();
            wybrane.textContent = "";
            przewinNaDol = true;
            pobierz();
        })
        .catch(function () {
            alert("Nie udało się wysłać wiadomości");
        })
        .finally(function () {
            przycisk.disabled = false;
            pole.focus();
        });
});

pobierz();
setInterval(pobierz, 2000);
