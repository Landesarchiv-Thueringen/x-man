# xdomea Aussonderungsmanager

## Kurzbeschreibung

Der xdomea Aussonderungsmanager (kurz x-man) ermöglicht die Ansicht, Bewertung und Archivierung von xdomea Aussonderungsnachrichten. Die Nutzeroberfläche wird als Webanwendung ausgeliefert und funktioniert in allen modernen Browsern. Die Metadaten von Anbietungen und Abgaben werden in einer Baumansicht dargestellt und können in jedem Prozessschritt eingesehen werden. Die Bewertung der Anbietung erfolgt direkt in der Anwendung. Bei der Bewertung sind die Metadaten, der zu bewerteten Schriftgutobjekte sichtbar. Die Bewertung kann jederzeit unterbrochen und zu einem späteren Zeitpunkt fortgesetzt werden. Alle Anbietungen, Abgaben und Schriftgutobjekte sind durch eine URL direkt adressierbar, die Links können in der Anwendung kopiert und geteilt werden. Fehler bei der Verarbeitung der Aussonderungsnachrichten werden in einer Steuerungsstelle angezeigt und können fehlerabhängig behandelt werden. Die Anwendung führt verschiedene Qualitätskontrollen für empfangene Nachrichten durch, wie bspw. ein Abgleich zwischen Bewertung und Abgabe, XML-Schemaprüfungen und Formaterkennung- und -validierung von Primärdateien durch. Die Formaterkennung und -validierung von Primärdateien wird von einem externen Tool (borgFormat, ebenfalls entwickelt vom LATh) durchgeführt. Ergebnisse der Formaterkennung und -validierung werden in der Anwendung angezeigt und mit den Primärdateien und Metadaten archiviert. Die Aussonderungsnachrichten werden von den Abgebenden Stellen über Transferverzeichnisse übertragen. Die Transferverzeichnisse werden von der Anwendung dauerhaft überwacht. Neue Nachrichten werden automatisch eingelesen und verarbeitet. Die technischen Nachrichten, die im Aussonderungsworkflow von xdomea vorgesehen sind (Bewertungsnachricht, diverse Empfangs- bzw. Importbestätigungen), werden von der Anwendung im entsprechenden Prozessschritt automatisch erstellt und in das Transferverzeichnis der abgebenden Stelle übertragen. Die Anwendung ermöglicht die Anbindung verschiedener Systeme (Repositories bzw. digitale Magazine) für die dauerhafte Speicherung des Archivguts.

## Development

### Debug Database with [pgweb](https://github.com/sosedoff/pgweb)

Run docker compose with the additional file `docker-compose-debug-db.yml` and go to http://localhost:8081

```sh
# Build and run everything with additional database debugging
docker-compose -f docker-compose.yml -f docker-compose-debug-db.yml up --build
# Run only database debugging
docker compose -f docker-compose-debug-db.yml up
```
