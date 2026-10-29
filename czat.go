package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const maxZdjecie = 5 << 20

var dozwoloneZdjecia = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

var poprawnaNazwaZdjecia = regexp.MustCompile(`^[a-f0-9]{32}\.(jpg|png|gif|webp)$`)

type WiadomoscJSON struct {
	ID        int    `json:"id"`
	Autor     string `json:"autor"`
	Tresc     string `json:"tresc"`
	Zdjecie   string `json:"zdjecie"`
	Godzina   string `json:"godzina"`
	Moja      bool   `json:"moja"`
	Edytowana bool   `json:"edytowana"`
}

func formatujDate(t time.Time) string {
	t = t.In(strefa)
	teraz := time.Now().In(strefa)
	if t.Year() == teraz.Year() && t.YearDay() == teraz.YearDay() {
		return t.Format("15:04")
	}
	return t.Format("02.01 15:04")
}

func stronaCzatu(w http.ResponseWriter, r *http.Request, u *Uzytkownik) {
	renderuj(w, "czat", map[string]any{"User": u})
}

func listaWiadomosci(w http.ResponseWriter, r *http.Request, u *Uzytkownik) {
	wiersze, err := db.Query(`SELECT * FROM (
		SELECT w.id, u.login, w.tresc, w.zdjecie, w.data, w.edytowana, w.autor_id
		FROM wiadomosci w JOIN uzytkownicy u ON u.id = w.autor_id
		ORDER BY w.id DESC LIMIT 50
	) x ORDER BY id`)
	if err != nil {
		blad500(w, err)
		return
	}
	defer wiersze.Close()

	lista := []WiadomoscJSON{}
	for wiersze.Next() {
		var wj WiadomoscJSON
		var data time.Time
		var autorID int
		if err := wiersze.Scan(&wj.ID, &wj.Autor, &wj.Tresc, &wj.Zdjecie, &data, &wj.Edytowana, &autorID); err != nil {
			blad500(w, err)
			return
		}
		if wj.Zdjecie != "" {
			wj.Zdjecie = "/zdjecia/" + wj.Zdjecie
		}
		wj.Godzina = formatujDate(data)
		wj.Moja = autorID == u.ID
		lista = append(lista, wj)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lista)
}

func wyslijWiadomosc(w http.ResponseWriter, r *http.Request, u *Uzytkownik) {
	r.Body = http.MaxBytesReader(w, r.Body, maxZdjecie+(1<<20))
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		http.Error(w, "Plik jest za duży (maksymalnie 5 MB)", http.StatusBadRequest)
		return
	}

	tresc := strings.TrimSpace(r.FormValue("tresc"))
	if utf8.RuneCountInString(tresc) > 1000 {
		http.Error(w, "Wiadomość jest za długa", http.StatusBadRequest)
		return
	}

	nazwaPliku := ""
	plik, _, err := r.FormFile("zdjecie")
	if err == nil {
		defer plik.Close()
		dane, err := io.ReadAll(io.LimitReader(plik, maxZdjecie+1))
		if err != nil {
			http.Error(w, "Błąd przesyłania pliku", http.StatusBadRequest)
			return
		}
		if len(dane) > maxZdjecie {
			http.Error(w, "Plik jest za duży (maksymalnie 5 MB)", http.StatusBadRequest)
			return
		}
		rozszerzenie, ok := dozwoloneZdjecia[http.DetectContentType(dane)]
		if !ok {
			http.Error(w, "Można wysyłać tylko zdjęcia JPG, PNG, GIF lub WEBP", http.StatusBadRequest)
			return
		}
		nazwaPliku = losowyToken()[:32] + rozszerzenie
		if err := os.WriteFile(filepath.Join(katalogZdjec, nazwaPliku), dane, 0644); err != nil {
			blad500(w, err)
			return
		}
	} else if err != http.ErrMissingFile {
		http.Error(w, "Błąd przesyłania pliku", http.StatusBadRequest)
		return
	}

	if tresc == "" && nazwaPliku == "" {
		http.Error(w, "Wiadomość jest pusta", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO wiadomosci (autor_id, tresc, zdjecie) VALUES ($1, $2, $3)", u.ID, tresc, nazwaPliku)
	if err != nil {
		blad500(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func pokazZdjecie(w http.ResponseWriter, r *http.Request, u *Uzytkownik) {
	plik := r.PathValue("plik")
	if !poprawnaNazwaZdjecia.MatchString(plik) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, max-age=86400")
	http.ServeFile(w, r, filepath.Join(katalogZdjec, plik))
}

func usunPlikZdjecia(nazwa string) {
	if poprawnaNazwaZdjecia.MatchString(nazwa) {
		os.Remove(filepath.Join(katalogZdjec, nazwa))
	}
}
