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
- [Skript für die Datenbankerstellung](/psql/init_database.sql)
  - Nutzer psql hat initial keine Leserechte -->
    - `sudo mkdir /var/lib/postgres/scripts`
    - `sudo cp psql/init_database.sql /var/lib/postgres/scripts`
    - `sudo chown -R postgres:postgres /var/lib/postgres/scripts`
    - `\i /var/lib/postgres/scripts/init_database.sql`

  ## Go Bibliotheken

  - [Gin](https://gin-gonic.com/)
  - [Gorm](https://gorm.io/)
  - [fsnotify](https://github.com/fsnotify/fsnotify)
  - [libxml2](https://github.com/lestrrat-go/libxml2)
    - C-Bibliothek libxml2, libxml2-dev notwendig (in Ubuntu bereits vorinstalliert)