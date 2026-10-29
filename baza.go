package main

import "log"

type Uzytkownik struct {
	ID          int
	Login       string
	Email       string
	Rola        string
	Zablokowany bool
}

func (u *Uzytkownik) JestAdminem() bool {
	return u.Rola == "admin"
}

func (u *Uzytkownik) JestModeratorem() bool {
	return u.Rola == "moderator" || u.Rola == "admin"
}

func utworzTabele() {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS uzytkownicy (
	id SERIAL PRIMARY KEY,
	login VARCHAR(30) NOT NULL,
	email VARCHAR(200) NOT NULL,
	haslo TEXT NOT NULL,
	rola VARCHAR(20) NOT NULL DEFAULT 'user',
	zablokowany BOOLEAN NOT NULL DEFAULT false,
	utworzony TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS uzytkownicy_login ON uzytkownicy (lower(login));
CREATE UNIQUE INDEX IF NOT EXISTS uzytkownicy_email ON uzytkownicy (lower(email));

CREATE TABLE IF NOT EXISTS sesje (
	token VARCHAR(64) PRIMARY KEY,
	uzytkownik_id INT NOT NULL REFERENCES uzytkownicy(id) ON DELETE CASCADE,
	wygasa TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS wiadomosci (
	id SERIAL PRIMARY KEY,
	autor_id INT NOT NULL REFERENCES uzytkownicy(id) ON DELETE CASCADE,
	tresc TEXT NOT NULL DEFAULT '',
	zdjecie VARCHAR(100) NOT NULL DEFAULT '',
	data TIMESTAMPTZ NOT NULL DEFAULT now(),
	edytowana BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS resety_hasel (
	token VARCHAR(64) PRIMARY KEY,
	uzytkownik_id INT NOT NULL REFERENCES uzytkownicy(id) ON DELETE CASCADE,
	wygasa TIMESTAMPTZ NOT NULL
);
`)
	if err != nil {
		log.Fatal("nie udalo sie utworzyc tabel: ", err)
	}
}
