# Changelog

## Next

- Feature: Anpassbare Paketierung
- Feature: Bestätige Erhalt der Abgabe mit 0507-Nachricht (xdomea >= 3.0)
- Feature: Überarbeiteter Übernahmebericht
- Feature: Verbesserte Fehlerausgabe in verschiedenen Fällen
- Feature: Laufzeit bei Sammelpaketen für Dokumente ohne Akte oder Vorgang
- Fix: Benachrichtigungen und Bestätigungen von xdomea-Nachrichten bei Fehlern
- Fix: Import nach DIMAG mit unbekannter Laufzeit
- Fix: Einstellung NO_PROXY wird nicht angewendet
- Fix: Kleinere UI-Fixes
- Fix: Fehler bei Erstellung des Übernahmeberichtes in manchen Fällen
- Fix: Löschen von unbekannten Ordnern in Transfer-Verzeichnis
- Intern: Verbesserte Kompatibilität mit Podman
- Intern: Verbesserte Fehlerbehandlung beim Versenden von xdomea-Nachrichten
- Intern: Prüfe Borg-Version
- Intern: Migration zum Signals-Mechanismus in Angular

## v1.0.0

- Feature: Formatverifikation und Archivierung können pausiert und abgebrochen werden
- Feature: Verbesserungen bei der Anzeige der Ergebnisse der Formatverifikation
- Feature: Zusätzlich angezeigtes Metadatum: Aktenplan-Titel
- Feature: Server-Konfiguration wird beim Programm-Start getestet
- Fix: Zugriff auf WebDav mit nicht-leerem Wurzel-Pfad
- Intern: Migration von PostgreSQL nach MongoDB
- Intern: Import nach DIMAG mittels BagIt
- Intern: Asynchrone API für Import nach DIMAG
- Intern: Fehlerbehandlung überarbeitet
- Intern: Automatische Seitenaktualisierung optimiert
- Intern: Abhängigkeiten aktualisiert

## v0.9.1

- Fix: Fehler bei Datenbank-Migration
- Fix: Nachricht beim Löschen nicht vollständig aus Datenbank entfernt
- Test: Automatisierter Test für kompletten Durchlauf
