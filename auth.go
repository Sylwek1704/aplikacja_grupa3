package main

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"log"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const nazwaCiasteczka = "sesja"

var poprawnyLogin = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

type handlerZUzytkownikiem func(http.ResponseWriter, *http.Request, *Uzytkownik)

func losowyToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func sha(tekst string) string {
	suma := sha256.Sum256([]byte(tekst))
	return hex.EncodeToString(suma[:])
}

func hashujHaslo(haslo string) (string, error) {
	sol := make([]byte, 16)
	rand.Read(sol)
	klucz, err := pbkdf2.Key(sha256.New, haslo, sol, 100000, 32)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sol) + "$" + hex.EncodeToString(klucz), nil
}

func sprawdzHaslo(haslo, zapisane string) bool {
	sol, hash, ok := strings.Cut(zapisane, "$")
	if !ok {
		return false
	}
	solBajty, err := hex.DecodeString(sol)
	if err != nil {
		return false
	}
	klucz, err := pbkdf2.Key(sha256.New, haslo, solBajty, 100000, 32)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(hex.EncodeToString(klucz)), []byte(hash)) == 1
}

func sprawdzNoweHaslo(haslo, haslo2 string) string {
	if utf8.RuneCountInString(haslo) < 8 {
		return "Hasło musi mieć co najmniej 8 znaków"
	}
	if len(haslo) > 200 {
		return "Hasło jest za długie"
	}
	if haslo != haslo2 {
		return "Hasła nie są takie same"
	}
	return ""
}

func aktualnyUzytkownik(r *http.Request) *Uzytkownik {
	c, err := r.Cookie(nazwaCiasteczka)
	if err != nil {
		return nil
	}
	u := &Uzytkownik{}
	err = db.QueryRow(`SELECT u.id, u.login, u.email, u.rola, u.zablokowany
		FROM sesje s JOIN uzytkownicy u ON u.id = s.uzytkownik_id
		WHERE s.token = $1 AND s.wygasa > now()`, sha(c.Value)).Scan(&u.ID, &u.Login, &u.Email, &u.Rola, &u.Zablokowany)
	if err != nil || u.Zablokowany {
		return nil
	}
	return u
}

func zaloguj(w http.ResponseWriter, id int) error {
	db.Exec("DELETE FROM sesje WHERE wygasa < now()")
	token := losowyToken()
	_, err := db.Exec("INSERT INTO sesje (token, uzytkownik_id, wygasa) VALUES ($1, $2, $3)", sha(token), id, time.Now().Add(24*time.Hour))
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     nazwaCiasteczka,
		Value:    token,
		Path:     "/",
		MaxAge:   24 * 60 * 60,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func tylkoZalogowani(h handlerZUzytkownikiem) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := aktualnyUzytkownik(r)
		if u == nil {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.Error(w, "Musisz się zalogować", http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		h(w, r, u)
	}
}

func tylkoModeratorzy(h handlerZUzytkownikiem) http.HandlerFunc {
	return tylkoZalogowani(func(w http.ResponseWriter, r *http.Request, u *Uzytkownik) {
		if !u.JestModeratorem() {
			http.Error(w, "Brak dostępu", http.StatusForbidden)
			return
		}
		h(w, r, u)
	})
}

func tylkoAdmin(h handlerZUzytkownikiem) http.HandlerFunc {
	return tylkoZalogowani(func(w http.ResponseWriter, r *http.Request, u *Uzytkownik) {
		if !u.JestAdminem() {
			http.Error(w, "Brak dostępu", http.StatusForbidden)
			return
		}
		h(w, r, u)
	})
}

func stronaLogowania(w http.ResponseWriter, r *http.Request) {
	if aktualnyUzytkownik(r) != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	dane := map[string]any{}
	if r.URL.Query().Get("info") == "haslo" {
		dane["Info"] = "Hasło zostało zmienione, możesz się zalogować"
	}
	renderuj(w, "login", dane)
}

func logowanie(w http.ResponseWriter, r *http.Request) {
	login := strings.TrimSpace(r.FormValue("login"))
	haslo := r.FormValue("haslo")

	var id int
	var zapisane string
	var zablokowany bool
	err := db.QueryRow("SELECT id, haslo, zablokowany FROM uzytkownicy WHERE lower(login) = lower($1)", login).Scan(&id, &zapisane, &zablokowany)
	if err != nil || !sprawdzHaslo(haslo, zapisane) {
		renderuj(w, "login", map[string]any{"Blad": "Zły login lub hasło", "Login": login})
		return
	}
	if zablokowany {
		renderuj(w, "login", map[string]any{"Blad": "Twoje konto zostało zablokowane", "Login": login})
		return
	}
	if err := zaloguj(w, id); err != nil {
		blad500(w, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func wylogowanie(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(nazwaCiasteczka); err == nil {
		db.Exec("DELETE FROM sesje WHERE token = $1", sha(c.Value))
	}
	http.SetCookie(w, &http.Cookie{Name: nazwaCiasteczka, Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func stronaRejestracji(w http.ResponseWriter, r *http.Request) {
	if aktualnyUzytkownik(r) != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderuj(w, "rejestracja", map[string]any{})
}

func rejestracja(w http.ResponseWriter, r *http.Request) {
	login := strings.TrimSpace(r.FormValue("login"))
	email := strings.TrimSpace(r.FormValue("email"))
	haslo := r.FormValue("haslo")
	haslo2 := r.FormValue("haslo2")
	dane := map[string]any{"Login": login, "Email": email}

	blad := ""
	if !poprawnyLogin.MatchString(login) {
		blad = "Login może mieć 3-30 znaków: litery, cyfry i _"
	} else if adres, err := mail.ParseAddress(email); err != nil || adres.Address != email || len(email) > 200 {
		blad = "Niepoprawny adres email"
	} else {
		blad = sprawdzNoweHaslo(haslo, haslo2)
	}
	if blad != "" {
		dane["Blad"] = blad
		renderuj(w, "rejestracja", dane)
		return
	}

	hash, err := hashujHaslo(haslo)
	if err != nil {
		blad500(w, err)
		return
	}

	rola := "user"
	var ile int
	db.QueryRow("SELECT count(*) FROM uzytkownicy").Scan(&ile)
	if ile == 0 {
		rola = "admin"
	}

	var id int
	err = db.QueryRow("INSERT INTO uzytkownicy (login, email, haslo, rola) VALUES ($1, $2, $3, $4) RETURNING id",
		login, email, hash, rola).Scan(&id)
	if err != nil {
		dane["Blad"] = "Taki login lub email jest już zajęty"
		renderuj(w, "rejestracja", dane)
		return
	}
	if err := zaloguj(w, id); err != nil {
		blad500(w, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func stronaZapomnialem(w http.ResponseWriter, r *http.Request) {
	renderuj(w, "zapomnialem", map[string]any{})
}

func zapomnialem(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.FormValue("email"))

	var id int
	var login, zapisanyEmail string
	err := db.QueryRow("SELECT id, login, email FROM uzytkownicy WHERE lower(email) = lower($1) AND NOT zablokowany", email).
		Scan(&id, &login, &zapisanyEmail)
	if err == nil {
		token := losowyToken()
		_, err = db.Exec("INSERT INTO resety_hasel (token, uzytkownik_id, wygasa) VALUES ($1, $2, $3)",
			sha(token), id, time.Now().Add(time.Hour))
		if err != nil {
			blad500(w, err)
			return
		}
		link := adresStrony + "/reset?token=" + token
		tresc := "Cześć " + login + ",\n\n" +
			"ktoś poprosił o zresetowanie hasła do Twojego konta na czacie.\n" +
			"Kliknij w link, żeby ustawić nowe hasło (link jest ważny 1 godzinę):\n\n" +
			link + "\n\n" +
			"Jeśli to nie Ty, po prostu zignoruj tę wiadomość.\n"
		if err := wyslijMaila(zapisanyEmail, "Reset hasła na czacie", tresc); err != nil {
			log.Println("nie udalo sie wyslac maila:", err)
			log.Println("link do resetu hasla dla", login+":", link)
		}
	} else if err != sql.ErrNoRows {
		blad500(w, err)
		return
	}

	renderuj(w, "zapomnialem", map[string]any{
		"Info": "Jeśli konto z tym adresem istnieje, wysłaliśmy na niego link do zmiany hasła.",
	})
}

func uzytkownikZTokenu(token string) (int, bool) {
	var id int
	err := db.QueryRow("SELECT uzytkownik_id FROM resety_hasel WHERE token = $1 AND wygasa > now()", sha(token)).Scan(&id)
	return id, err == nil
}

func stronaResetu(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if _, ok := uzytkownikZTokenu(token); !ok {
		renderuj(w, "reset", map[string]any{"ZlyLink": true})
		return
	}
	renderuj(w, "reset", map[string]any{"Token": token})
}

func resetHasla(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	id, ok := uzytkownikZTokenu(token)
	if !ok {
		renderuj(w, "reset", map[string]any{"ZlyLink": true})
		return
	}
	if blad := sprawdzNoweHaslo(r.FormValue("haslo"), r.FormValue("haslo2")); blad != "" {
		renderuj(w, "reset", map[string]any{"Token": token, "Blad": blad})
		return
	}
	hash, err := hashujHaslo(r.FormValue("haslo"))
	if err != nil {
		blad500(w, err)
		return
	}
	if _, err := db.Exec("UPDATE uzytkownicy SET haslo = $1 WHERE id = $2", hash, id); err != nil {
		blad500(w, err)
		return
	}
	db.Exec("DELETE FROM resety_hasel WHERE uzytkownik_id = $1", id)
	db.Exec("DELETE FROM sesje WHERE uzytkownik_id = $1", id)
	http.Redirect(w, r, "/login?info=haslo", http.StatusSeeOther)
}
