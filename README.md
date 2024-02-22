# xdomea Aussonderungsmanager

## Kurzbeschreibung

Der xdomea Aussonderungsmanager (kurz x-man) ermöglicht die Ansicht, Bewertung und Archivierung von xdomea Aussonderungsnachrichten. Die Nutzeroberfläche wird als Webanwendung ausgeliefert und funktioniert in allen modernen Browsern. Die Metadaten von Anbietungen und Abgaben werden in einer Baumansicht dargestellt und können in jedem Prozessschritt eingesehen werden. Die Bewertung der Anbietung erfolgt direkt in der Anwendung. Bei der Bewertung sind die Metadaten, der zu bewerteten Schriftgutobjekte sichtbar. Die Bewertung kann jederzeit unterbrochen und zu einem späteren Zeitpunkt fortgesetzt werden. Alle Anbietungen, Abgaben und Schriftgutobjekte sind durch eine URL direkt adressierbar, die Links können in der Anwendung kopiert und geteilt werden. Fehler bei der Verarbeitung der Aussonderungsnachrichten werden in einer Steuerungsstelle angezeigt und können fehlerabhängig behandelt werden. Die Anwendung führt verschiedene Qualitätskontrollen für empfangene Nachrichten durch, wie bspw. ein Abgleich zwischen Bewertung und Abgabe, XML-Schemaprüfungen und Formaterkennung- und -validierung von Primärdateien durch. Die Formaterkennung und -validierung von Primärdateien wird von einem externen Tool (borgFormat, ebenfalls entwickelt vom LATh) durchgeführt. Ergebnisse der Formaterkennung und -validierung werden in der Anwendung angezeigt und mit den Primärdateien und Metadaten archiviert. Die Aussonderungsnachrichten werden von den Abgebenden Stellen über Transferverzeichnisse übertragen. Die Transferverzeichnisse werden von der Anwendung dauerhaft überwacht. Neue Nachrichten werden automatisch eingelesen und verarbeitet. Die technischen Nachrichten, die im Aussonderungsworkflow von xdomea vorgesehen sind (Bewertungsnachricht, diverse Empfangs- bzw. Importbestätigungen), werden von der Anwendung im entsprechenden Prozessschritt automatisch erstellt und in das Transferverzeichnis der abgebenden Stelle übertragen. Die Anwendung ermöglicht die Anbindung verschiedener Systeme (Repositories bzw. digitale Magazine) für die dauerhafte Speicherung des Archivguts.

## Running with Docker Compose

There is a development configuration and a production configuration for running x-man using Docker Compose.

```sh
# Run for development
docker compose up --build
# Run for production
docker compose -f compose.yml -f compose.prod.yml up --build
```

## Development

When run with development configuration, there are some additional options to help with testing and debugging.

### Frontend Development Server

To run with a auto-refreshing development server for frontend development, run

```sh
# Run a minimal backend configuration. You can also start the complete stack without specifying "server".
docker compose up --build server
# Start the frontend development server
cd gui
npm start
```

### Test without LDAP

You can use X-Man without LDAP configuration. In `.env`, set

```sh
ACCEPT_ANY_LOGIN_CREDENTIALS=true
```

You can now log in with any username/password combination. Use 'admin' as
password to log in with administration privileges.

### Debug Database with [pgweb](https://github.com/sosedoff/pgweb)

pgweb is automatically enabled by the development configuration. Go to http://localhost:8081.
