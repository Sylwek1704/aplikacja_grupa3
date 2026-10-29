# Czat - projekt grupa 3

Aplikacja w Go + PostgreSQL. Po zalogowaniu jest czat z innymi użytkownikami (tekst i zdjęcia).

Co jest:
- rejestracja, logowanie, wylogowanie
- czat ze zdjęciami (jpg, png, gif, webp do 5 MB)
- reset hasła przez maila
- panel moderatora (edycja i usuwanie wiadomości)
- panel admina (role i blokowanie kont)

## Kompilacja

Na WSL:

    ./build.sh

albo na Windowsie `build.bat`. Wychodzą pliki `bin/czat` i `bin/czat.exe`.

## Środowisko testowe w dockerze

    ./start.sh

Kompiluje aplikację i stawia bazę, skrzynkę na maile i czat.

- czat: http://localhost:8081
- maile: http://localhost:1081
- baza: localhost:5433 (postgres / postgres)

Zatrzymanie: `./stop.sh`, a `./stop.sh -v` dodatkowo kasuje bazę i zdjęcia.

## Uruchomienie bez dockera

Strona jest na http://localhost:8080. Ustawienia przez zmienne środowiskowe:
`DB_URL`, `PORT`, `ADRES`, `ZDJECIA`, `SMTP_ADRES`, `SMTP_OD`, `SMTP_LOGIN`, `SMTP_HASLO`.
