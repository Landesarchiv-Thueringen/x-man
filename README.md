# xdomea-Aussonderungsmanager

Der xdomea-Aussonderungsmanager (kurz x-man) ermöglicht die Ansicht, Bewertung und Archivierung von E-Akten in Form von xdomea Aussonderungsnachrichten.

Die Nutzeroberfläche wird als Webanwendung ausgeliefert und funktioniert in allen modernen Browsern.
Die Anwendung kommuniziert über automatisierte Schnittstellen mit externen Diensten und die Bedienung geschieht medienbruchfrei in der Anwendung.

Siehe [Dokumentation](https://landesarchiv-thueringen.github.io/x-man/) für weitere Informationen zu Bedienung und Betrieb.

![Nachrichten-Ansicht](./docs/img/message-page.png)

## Funktionen

### Hauptfunktionen

- Annahme von Aussonderungen und Austausch von Nachrichten mit abgebenden Stellen nach dem [xdomea-Standard](https://www.xrepository.de/details/urn:xoev-de:xdomea:kosit:standard:xdomea) (Versionen 2.3 bis 3.1) im zwei- und vierstufigen Aussonderungsverfahren
- Ansicht und Bewertung von Anbietungen in der Weboberfläche
- Formaterkennung und -validierung mit Hilfe von [BorgFormat](https://github.com/Landesarchiv-Thueringen/borg)
- Import von Abgaben in das [DIMAG Kernmodul](https://gitlab.la-bw.de/dimag/core/kernmodul) für die dauerhafte Archivierung

### Ergänzende Funktionen

- Erstellen eines Bewertungsberichtes nach Abschluss der archivischen Bewertung
- Erstellen eines Übernahmeberichtes nach erfolgreicher Archivierung
- E-Mail-Benachrichtigungen bei neuen xdomea-Nachrichten oder im Fehlerfall
- Nutzerverwaltung über ein bestehendes LDAP / Active Directory
- Administration und Fehlerbehandlung durch gesonderte Administratoren
- Zuordnung von Archivarinnen zu abgebenden Stellen

## Entwicklungsstand

Die Anwendung wurde bisher nur mit generierten Testdaten geprüft. Ein Test mit Echtdaten ist noch nicht abgeschlossen. Für den produktiven Einsatz sollten Backups der Aussonderungsnachrichten erstellt werden, weiterhin sollten die Aussonderungsnachrichten erst gelöscht werden, wenn die vollständige Archivierung geprüft werden konnte.

## Bedienung

- [Benutzerhandbuch](https://landesarchiv-thueringen.github.io/x-man/benutzerhandbuch/)
- [Betriebshandbuch](https://landesarchiv-thueringen.github.io/x-man/betriebshandbuch/)

## Betrieb und Entwicklung

- [Installation (en)](https://landesarchiv-thueringen.github.io/x-man/installation/)
- [Development (en)](https://landesarchiv-thueringen.github.io/x-man/development/)
- [Tests (en)](./test/README.md)

## Roadmap

- Unterstützung xdomea 4.0
- Hashwertprüfung
- AFIS-Schnittstelle für die automatische Bildung von Verzeichnungseinheiten bei der Archivierung

## Lizenz

Dieses Projekt wird unter der [GNU General Public License Version 3 (GPLv3)](https://www.gnu.org/licenses/gpl-3.0.de.html) veröffentlicht.

Von der Lizenz ausgenommen sind xdomea-Schemadateien (.xsd), deren Verwendung vom Rechteinhaber im Rahmen dieses Projekts uneingeschränkt erlaubt wurde.
