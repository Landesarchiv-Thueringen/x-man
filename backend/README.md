# LATh xdomea Server

## Abhängigkeiten

- Golang (Programmiersprache)
- Gin (Web Framework)
- Gorm (ORM)
- PostgreSQL (Datenbank)
  - Vorteil gegenüber MariaDB ist Performance und native Datentyp XML Unterstützung

## Datenbank

- Nutzer: lath_xdomea (kein Passwort, Besitzer der Datenbank)
- Name: lath_xdomea
- Passwort muss für Produktiveinsatz neu vergeben werden
- Migration der Datenbank muss bei der ersten Ausführung mit der Flag -init durchgeführt werden
- Datenbank und Nutzer erstellen:
  - `create user lath_xdomea;`
  - `create database lath_xdomea owner lath_xdomea;`
  - `grant all privileges on database lath_xdomea to postgres;`
- Erweiterung für UUID-Generierung erstellen
  - `\c lath_xdomea`
  - `create extension if not exists "uuid-ossp";`
- als postgres root Nutzer anmelden
  - `sudo -u postgres psql`