package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type WiadomoscPanel struct {
	ID        int
	Autor     string
	Tresc     string
	Zdjecie   string
	Data      string
	Edytowana bool
}

var komunikatyPanelu = map[string]string{
	"zapisano": "Zapisano zmiany w wiadomości",
	"pusta":    "Wiadomość nie może być pusta, jeśli chcesz ją usunąć kliknij Usuń",
	"usunieto": "Usunięto wiadomość",
	"zdjecie":  "Usunięto zdjęcie",
	"rola":     "Zmieniono rolę użytkownika",
	"blokada":  "Zmieniono status konta",
}

var role = map[string]bool{"user": true, "moderator": true, "admin": true}

func idZeSciezki(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	return id, err == nil
}

func wrocDoPanelu(w http.ResponseWriter, r *http.Request, info string) {
	http.Redirect(w, r, "/panel?info="+info, http.StatusSeeOther)
}

func stronaPanelu(w http.ResponseWriter, r *http.Request, u *Uzytkownik) {
	szukaj := strings.TrimSpace(r.URL.Query().Get("szukaj"))

	wiersze, err := db.Query(`SELECT w.id, u.login, w.tresc, w.zdjecie, w.data, w.edytowana
		FROM wiadomosci w JOIN uzytkownicy u ON u.id = w.autor_id
		WHERE $1 = '' OR w.tresc ILIKE '%' || $1 || '%' OR u.login ILIKE '%' || $1 || '%'
		ORDER BY w.id DESC LIMIT 200`, szukaj)
	if err != nil {
		blad500(w, err)
		return
	}
	defer wiersze.Close()

	wiadomosci := []WiadomoscPanel{}
	for wiersze.Next() {
		var wp WiadomoscPanel
		var data time.Time
		if err := wiersze.Scan(&wp.ID, &wp.Autor, &wp.Tresc, &wp.Zdjecie, &data, &wp.Edytowana); err != nil {
			blad500(w, err)
			return
		}
		wp.Data = data.In(strefa).Format("02.01.2006 15:04")
		wiadomosci = append(wiadomosci, wp)
	}

	uzytkownicy := []Uzytkownik{}
	if u.JestAdminem() {
		wiersze2, err := db.Query("SELECT id, login, email, rola, zablokowany FROM uzytkownicy ORDER BY id")
		if err != nil {
			blad500(w, err)
			return
		}
		defer wiersze2.Close()
		for wiersze2.Next() {
			var uz Uzytkownik
			if err := wiersze2.Scan(&uz.ID, &uz.Login, &uz.Email, &uz.Rola, &uz.Zablokowany); err != nil {
				blad500(w, err)
				return
			}
			uzytkownicy = append(uzytkownicy, uz)
		}
	}

	renderuj(w, "panel", map[string]any{
		"User":        u,
		"Wiadomosci":  wiadomosci,
		"Uzytkownicy": uzytkownicy,
		"Szukaj":      szukaj,
		"Info":        komunikatyPanelu[r.URL.Query().Get("info")],
	})
}

func edytujWiadomosc(w http.ResponseWriter, r *http.Request, u *Uzytkownik) {
	id, ok := idZeSciezki(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	tresc := strings.TrimSpace(r.FormValue("tresc"))
	wynik, err := db.Exec(`UPDATE wiadomosci SET tresc = $1, edytowana = true
		WHERE id = $2 AND ($1 <> '' OR zdjecie <> '')`, tresc, id)
	if err != nil {
		blad500(w, err)
		return
	}
	if ile, _ := wynik.RowsAffected(); ile == 0 {
		wrocDoPanelu(w, r, "pusta")
		return
	}
	wrocDoPanelu(w, r, "zapisano")
}

func usunWiadomosc(w http.ResponseWriter, r *http.Request, u *Uzytkownik) {
	id, ok := idZeSciezki(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	var zdjecie string
	err := db.QueryRow("DELETE FROM wiadomosci WHERE id = $1 RETURNING zdjecie", id).Scan(&zdjecie)
	if err == nil {
		usunPlikZdjecia(zdjecie)
	}
	wrocDoPanelu(w, r, "usunieto")
}

func usunZdjecie(w http.ResponseWriter, r *http.Request, u *Uzytkownik) {
	id, ok := idZeSciezki(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	var zdjecie string
	if err := db.QueryRow("SELECT zdjecie FROM wiadomosci WHERE id = $1", id).Scan(&zdjecie); err != nil {
		wrocDoPanelu(w, r, "zdjecie")
		return
	}
	db.Exec("UPDATE wiadomosci SET zdjecie = '', edytowana = true WHERE id = $1", id)
	db.Exec("DELETE FROM wiadomosci WHERE id = $1 AND tresc = ''", id)
	usunPlikZdjecia(zdjecie)
	wrocDoPanelu(w, r, "zdjecie")
}

func zmienRole(w http.ResponseWriter, r *http.Request, u *Uzytkownik) {
	id, ok := idZeSciezki(r)
	rola := r.FormValue("rola")
	if !ok || id == u.ID || !role[rola] {
		http.Error(w, "Niepoprawne dane", http.StatusBadRequest)
		return
	}
	if _, err := db.Exec("UPDATE uzytkownicy SET rola = $1 WHERE id = $2", rola, id); err != nil {
		blad500(w, err)
		return
	}
	wrocDoPanelu(w, r, "rola")
}

func zmienBlokade(w http.ResponseWriter, r *http.Request, u *Uzytkownik) {
	id, ok := idZeSciezki(r)
	if !ok || id == u.ID {
		http.Error(w, "Niepoprawne dane", http.StatusBadRequest)
		return
	}
	if _, err := db.Exec("UPDATE uzytkownicy SET zablokowany = NOT zablokowany WHERE id = $1", id); err != nil {
		blad500(w, err)
		return
	}
	db.Exec("DELETE FROM sesje WHERE uzytkownik_id = $1", id)
	wrocDoPanelu(w, r, "blokada")
}
