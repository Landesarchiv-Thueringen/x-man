# xdomea Aussonderungsmanager

## Kurzbeschreibung

Der xdomea Aussonderungsmanager (kurz x-man) ermöglicht die Ansicht, Bewertung und Archivierung von xdomea Aussonderungsnachrichten. Die Nutzeroberfläche wird als Webanwendung ausgeliefert und funktioniert in allen modernen Browsern. Die Metadaten von Anbietungen und Abgaben werden in einer Baumansicht dargestellt und können in jedem Prozessschritt eingesehen werden. Die Bewertung der Anbietung erfolgt direkt in der Anwendung. Bei der Bewertung sind die Metadaten, der zu bewerteten Schriftgutobjekte sichtbar. Die Bewertung kann jederzeit unterbrochen und zu einem späteren Zeitpunkt fortgesetzt werden. Alle Anbietungen, Abgaben und Schriftgutobjekte sind durch eine URL direkt adressierbar, die Links können in der Anwendung kopiert und geteilt werden. Fehler bei der Verarbeitung der Aussonderungsnachrichten werden in einer Steuerungsstelle angezeigt und können fehlerabhängig behandelt werden. Die Anwendung führt verschiedene Qualitätskontrollen für empfangene Nachrichten durch, wie bspw. ein Abgleich zwischen Bewertung und Abgabe, XML-Schemaprüfungen und Formaterkennung- und -validierung von Primärdateien durch. Die Formaterkennung und -validierung von Primärdateien wird von einem externen Tool (borgFormat, ebenfalls entwickelt vom LATh) durchgeführt. Ergebnisse der Formaterkennung und -validierung werden in der Anwendung angezeigt und mit den Primärdateien und Metadaten archiviert. Die Aussonderungsnachrichten werden von den Abgebenden Stellen über Transferverzeichnisse übertragen. Die Transferverzeichnisse werden von der Anwendung dauerhaft überwacht. Neue Nachrichten werden automatisch eingelesen und verarbeitet. Die technischen Nachrichten, die im Aussonderungsworkflow von xdomea vorgesehen sind (Bewertungsnachricht, diverse Empfangs- bzw. Importbestätigungen), werden von der Anwendung im entsprechenden Prozessschritt automatisch erstellt und in das Transferverzeichnis der abgebenden Stelle übertragen. Die Anwendung ermöglicht die Anbindung verschiedener Systeme (Repositories bzw. digitale Magazine) für die dauerhafte Speicherung des Archivguts.

## Requirements

x-man runs on any Linux system that provides the following dependencies:

- Docker (configured and with running daemon)
- Docker Compose
- OpenSSL (for testing setup)
- NodeJS and NPM (for frontend development)

Testing on Windows is possible with a Linux virtual machine or using the [Windows Subsystem for Linux (WSL)](https://learn.microsoft.com/en-us/windows/wsl/install) version 2 (WSL 2).

## Getting Started

For a minimal testing setup run the following commands:

```sh
# Create TLS certificates for testing
./scripts/generate-test-certificates.sh
# Create initial configuration
cp .env.example .env
# Build and run for development / testing
docker compose up --build
```

In case you are behind a http proxy, you will need to provide its address with
`HTTP_PROXY` in `.env` before running the last command.

The application will be exposed on http://localhost:8080.

Login:

- fry / fry (user)
- hermes / hermes (administrator)

### Tear Down

```sh
docker compose down --volumes
```

## Test Setup

When following the steps of [Getting Started](#getting-started), some outside
services will be replaced by dummy implementations that are not intended for
production use but allow to create a testing environment without much
configuration.

### LDAP

The development configuration comes with an [OpenLDAP Docker Image for testing](https://github.com/rroemhild/docker-test-openldap).

You can login as any member of the group `ship_crew` and `admin_staff` while the
latter has administration privileges.

### E-Mail

Notification e-mails are sent to a [Mailhog](https://github.com/mailhog/MailHog) instance.

A web UI is available on http://localhost:8025.

### Certificates

The script `generate-test-certificates.sh` creates dummy certificates for the services above.

A root certificate is created and installed in the application container, so it
can verify the certificates and establish secure connections.

## Production Use

### Build and Run

```sh
# Build and run for production
docker compose -f compose.yml -f compose.prod.yml up --build -d
```

Re-run the command after changing configuration.

### Configuration

Copy `.env.example` to `.env` and adjust values as described in the file.

### Custom TLS Certificates

In case any servers you provide in `.env` cannot present certificates that are
signed by a commonly accepted CA, you need to provide any involved root- and
intermediate certificates to x-man.

Copy certificates in PEM format with the file ending `.crt` to
`server/data/ca-certificates` and rerun `docker compose build`. This allows you
to remove the `TLS_INSECURE_SKIP_VERIFY` option for URLs that you provided
certificates for.

Remember to remove the generated `Test-RootCA.crt` for production use.

## Development

When run with development configuration, there are some additional options to help with testing and debugging.

### Frontend Development Server

To run with a auto-refreshing development server for frontend development, run

```sh
# Run a minimal backend configuration. You can also start the complete stack without specifying "server".
docker compose up --build -d server
# Start the frontend development server
cd gui
npm start
```

### Debug the Database

The development configuration starts an instance of [pgweb](https://github.com/sosedoff/pgweb) and connects it to the application database.

Its web UI is available on http://localhost:8081.
