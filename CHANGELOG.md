# Changelog

## Next

-   Feature: Berichte als PDF/A-2b
-   Fix: Versionsangabe
-   Fix: Versionsprüfung für Borg
-   Intern: Abhängigkeiten aktualisiert
-   Intern: Versionen für Docker-Images festgeschrieben

## v1.2.0

-   Feature: Bewertungsbericht
-   Feature: Unterstützung für tiefere Paketierungsebenen in Übernahmebericht
-   Fix: Fehler beim Einlesen von nicht UTF-8 kodierten Dateinamen aus der Nachricht
-   Fix: Setze Dateiname für Meta-Dateien beim Import nach DIMAG
-   Fix: Kleinere UI-Fixes
-   Fix: Zeitangaben teilweise in UTC
-   Fix: Kompatibilität Borg Version 1.4.0
-   Intern: Benutze "Import"-Ordner für Upload zu DIMAG
-   Intern: Docker-Image von Server optimieren
-   Intern: Abhängigkeiten aktualisiert
-   Intern: Netzwerkanfragen optimiert

## v1.1.0

-   Feature: Anpassbare Paketierung
-   Feature: Bestätige Erhalt der Abgabe mit 0507-Nachricht (xdomea >= 3.0)
-   Feature: Überarbeiteter Übernahmebericht
-   Feature: Verbesserte Fehlerausgabe in verschiedenen Fällen
-   Feature: Laufzeit bei Sammelpaketen für Dokumente ohne Akte oder Vorgang
-   Feature: Warnung bei nicht zugeordneten Dokumenten
-   Feature: Prüfsummen bei Archivierung in Dateisystem
-   Feature: Archiviere Ergebnisse der Formatverifikation
-   Feature: xdomea-Nachrichten an zentrale Poststelle weiterleiten
-   Feature: Filtern nach Laufzeit
-   Feature: Anzeige und Filter für feder- und aktenführende Organisationseinheit
-   Fix: Fehler beim Übertragen von Bewertungsentscheidungen im Objekt-Baum
-   Fix: Benachrichtigungen und Bestätigungen von xdomea-Nachrichten bei Fehlern
-   Fix: Import nach DIMAG mit unbekannter Laufzeit
-   Fix: Einstellung NO_PROXY wird nicht angewendet
-   Fix: Kleinere UI-Fixes
-   Fix: Fehler bei Erstellung des Übernahmeberichtes in manchen Fällen
-   Fix: Löschen von unbekannten Ordnern in Transfer-Verzeichnis
-   Fix: Login mit Umlauten im Nutzernamen
-   Intern: Verbesserte Kompatibilität mit Podman
-   Intern: Verbesserte Fehlerbehandlung beim Versenden von xdomea-Nachrichten
-   Intern: Prüfe Borg-Version
-   Intern: Migration zum Signals-Mechanismus in Angular
-   Intern: Konfigurierbare Ports für Dienste
-   Intern: Binde Volumes für persistente Daten ins Dateisystem
-   Intern: Migration auf PNPM

## v1.0.0

-   Feature: Formatverifikation und Archivierung können pausiert und abgebrochen werden
-   Feature: Verbesserungen bei der Anzeige der Ergebnisse der Formatverifikation
-   Feature: Zusätzlich angezeigtes Metadatum: Aktenplan-Titel
-   Feature: Server-Konfiguration wird beim Programm-Start getestet
-   Fix: Zugriff auf WebDav mit nicht-leerem Wurzel-Pfad
-   Intern: Migration von PostgreSQL nach MongoDB
-   Intern: Import nach DIMAG mittels BagIt
-   Intern: Asynchrone API für Import nach DIMAG
-   Intern: Fehlerbehandlung überarbeitet
-   Intern: Automatische Seitenaktualisierung optimiert
-   Intern: Abhängigkeiten aktualisiert

## v0.9.1

-   Fix: Fehler bei Datenbank-Migration
-   Fix: Nachricht beim Löschen nicht vollständig aus Datenbank entfernt
-   Test: Automatisierter Test für kompletten Durchlauf
