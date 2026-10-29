package main

import (
	"database/sql"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"
	_ "time/tzdata"

	_ "github.com/lib/pq"
)

//go:embed szablony/*.html
var szablonyFS embed.FS

//go:embed static
var staticFS embed.FS

var (
	db           *sql.DB
	szablony     = map[string]*template.Template{}
	strefa       *time.Location
	katalogZdjec string
	adresStrony  string
)

func env(nazwa, domyslna string) string {
	wartosc := os.Getenv(nazwa)
	if wartosc == "" {
		return domyslna
	}
	return wartosc
}

func main() {
	var err error
	strefa, err = time.LoadLocation("Europe/Warsaw")
	if err != nil {
		log.Fatal(err)
	}
	katalogZdjec = env("ZDJECIA", "zdjecia")
	port := env("PORT", "8080")
	adresStrony = env("ADRES", "http://localhost:"+port)

	db, err = sql.Open("postgres", env("DB_URL", "postgres://postgres:postgres@localhost:5432/czat?sslmode=disable"))
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("nie mozna polaczyc sie z baza: ", err)
	}
	utworzTabele()
	wczytajSzablony()
	if err = os.MkdirAll(katalogZdjec, 0755); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	statyczne, _ := fs.Sub(staticFS, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(statyczne)))

	mux.HandleFunc("GET /login", stronaLogowania)
	mux.HandleFunc("POST /login", logowanie)
	mux.HandleFunc("POST /wyloguj", wylogowanie)
	mux.HandleFunc("GET /rejestracja", stronaRejestracji)
	mux.HandleFunc("POST /rejestracja", rejestracja)
	mux.HandleFunc("GET /zapomnialem", stronaZapomnialem)
	mux.HandleFunc("POST /zapomnialem", zapomnialem)
	mux.HandleFunc("GET /reset", stronaResetu)
	mux.HandleFunc("POST /reset", resetHasla)

	mux.HandleFunc("GET /{$}", tylkoZalogowani(stronaCzatu))
	mux.HandleFunc("GET /api/wiadomosci", tylkoZalogowani(listaWiadomosci))
	mux.HandleFunc("POST /api/wiadomosci", tylkoZalogowani(wyslijWiadomosc))
	mux.HandleFunc("GET /zdjecia/{plik}", tylkoZalogowani(pokazZdjecie))

	mux.HandleFunc("GET /panel", tylkoModeratorzy(stronaPanelu))
	mux.HandleFunc("POST /panel/wiadomosci/{id}/edytuj", tylkoModeratorzy(edytujWiadomosc))
	mux.HandleFunc("POST /panel/wiadomosci/{id}/usun", tylkoModeratorzy(usunWiadomosc))
	mux.HandleFunc("POST /panel/wiadomosci/{id}/usun-zdjecie", tylkoModeratorzy(usunZdjecie))
	mux.HandleFunc("POST /panel/uzytkownicy/{id}/rola", tylkoAdmin(zmienRole))
	mux.HandleFunc("POST /panel/uzytkownicy/{id}/blokada", tylkoAdmin(zmienBlokade))

	log.Println("serwer dziala na " + adresStrony)
	log.Fatal(http.ListenAndServe(":"+port, http.NewCrossOriginProtection().Handler(mux)))
}

func wczytajSzablony() {
	for _, nazwa := range []string{"login", "rejestracja", "zapomnialem", "reset", "czat", "panel"} {
		szablony[nazwa] = template.Must(template.ParseFS(szablonyFS, "szablony/base.html", "szablony/"+nazwa+".html"))
	}
}

func renderuj(w http.ResponseWriter, nazwa string, dane map[string]any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := szablony[nazwa].ExecuteTemplate(w, "base", dane); err != nil {
		log.Println("blad szablonu:", err)
	}
}

func blad500(w http.ResponseWriter, err error) {
	log.Println("blad:", err)
	http.Error(w, "Coś poszło nie tak, spróbuj ponownie", http.StatusInternalServerError)
}
