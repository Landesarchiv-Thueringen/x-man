
#import "shared.typ": formatDate, formatDateTime, formatContentStatsElement, appraisalStatsTable, appraisalStatsGraph

#let fallback(input, fallback: "-") = {
  if input == "" or input == none or input == 0 { fallback } else { input }
}

#let formatFloat(f, digitsAfterPoint) = {
  let factor = calc.pow(10, digitsAfterPoint)
  let beforePoint = calc.floor(f)
  let afterPoint = f - beforePoint
  str(beforePoint) + "," + str(calc.round(afterPoint * factor))
}

#let formatFileSize(nBytes) = {
  let v = nBytes
  let suffix = [B]
  for c in "KMGTPE" {
    let newV = v / 1024
    if newV > 1 {
      suffix = c + "B"
      v = newV
    } else {
      break
    }
  }
  [#formatFloat(v, 2) #suffix]
}


#let formatValidity(validity) = {
  if (validity == none) { "-" } else if (validity) { "valide" } else { "invalide" }
}

#let formatLifetime(lifetime) = {
  if (lifetime == none) {
    return [-]
  }
  let keys = lifetime.keys()
  if (lifetime.start != "" and lifetime.end != "") {
    [#formatDate(lifetime.start) -- #formatDate(lifetime.end)]
  } else if (lifetime.start != "") {
    [ab #formatDate(lifetime.start)]
  } else if (lifetime.end != "") {
    [bis #formatDate(lifetime.end)]
  } else {
    [-]
  }
}

#let topMatter(data) = [
  #block(spacing: 2em)[
    #set text(2em)
    *Übernahmebericht*
  ]
  #table(
    columns: 2,
    stroke: none,
    [Abgegebene Stelle:],
    data.Process.agency.name,
    [Objektart:],
    [E-Akte],
    [Prozess-ID:],
    data.Process.processId,
    [Aussonderungsverfahren:],
    if data.Process.processState.receive0501.complete [
      4-stufig
    ] else [
      2-stufig
    ],
    ..if data.Process.processState.appraisal.complete {
      (
        [Anbietung erhalten:],
        formatDateTime(data.Process.processState.receive0501.completedAt),
        [Bewertung versendet:],
        formatDateTime(data.Process.processState.appraisal.completedAt),
        [Bewertung durch:],
        data.Process.processState.appraisal.completedBy,
      )
    } else {
      (
        [Abgabe erhalten:],
        formatDateTime(data.Process.processState.receive0503.completedAt),
      )
    },
    [Abgabe archiviert:],
    formatDateTime(data.Process.processState.archiving.completedAt),
    [Archivierung durch:],
    data.Process.processState.archiving.completedBy,
  )
]

#let overview(data) = [
  #v(1fr)
  = Übersicht
  #if data.AppraisalStats == none [
    #table(
      columns: 1fr,
      stroke: none,
      [*Übernommen*],
      ..if data.Message0503Stats.Files > 0 {
        ([#formatContentStatsElement("Files", data.Message0503Stats.Files)],)
      },
      ..if data.Message0503Stats.Processes > 0 {
        (
          [#formatContentStatsElement("Processes", data.Message0503Stats.Processes)],
        )
      },
      ..if data.Message0503Stats.Documents > 0 {
        (
          [#formatContentStatsElement("Documents", data.Message0503Stats.Documents)],
        )
      },
    )
    #table(
      columns: 2,
      stroke: none,
      [Gesamt-?speicher-?volumen übernommen:], [#formatFileSize(data.FileStats.TotalBytes)],
    )
    #v(10fr)
  ] else [
    #appraisalStatsTable(data)

    #table(
      columns: 2,
      stroke: none,
      [Gesamt-?speicher-?volumen übernommen:], [#formatFileSize(data.FileStats.TotalBytes)],
    )

    #[
      #v(1fr)
      #appraisalStatsGraph(data)
      #v(2fr)
    ]
  ]
]

#let discrepancies(discrepancies) = [
  = Diskrepanzen
  #if discrepancies.MissingRecords != none [
    == Fehlende Schriftgutobjekte (#discrepancies.MissingRecords.len())

    In der Abgabe fehlen folgende Schriftgutobjekte, die in der Anbietung als zu
    archivieren bewertet wurden.

    #for el in discrepancies.MissingRecords [
      - #el
    ]
  ]
  #if discrepancies.SurplusRecords != none [
    == Zusätzliche Schriftgutobjekte (#discrepancies.SurplusRecords.len())

    Die Abgabe enthält folgende Schriftgutobjekte, die entweder in der Anbietung
    nicht vorhanden waren, oder als zu vernichten bewertet wurden.

    #for el in discrepancies.SurplusRecords [
      - #el
    ]
  ]
  #if discrepancies.MissingPrimaryDocuments != none [
    == Fehlende Primärdokumente (#discrepancies.MissingPrimaryDocuments.len())

    Folgende Dateien sind nicht in der Abgabe enthalten, obwohl die xdomea-Nachricht
    die Primärdokumente referenziert.
    #for el in discrepancies.MissingPrimaryDocuments [
      - #el.replace("_", "_" + str(sym.zws))
    ]
  ]
]

#let fileStats(fileStats) = [
  = Formatstatistik
  #table(
    columns: 5,
    stroke: none,
    align: (x, y) => (if x == 4 and y > 0 { right } else { left }),
    [*PUID*],
    [*MIME-Type*],
    [*Formatversion*],
    [*Validität*],
    [*Dateien*],
    ..fileStats.PUIDEntries.map(p => {
      let rows = ()
      let first = true
      for e in p.Entries {
        if first {
          rows.push(p.PUID)
        } else {
          rows.push([])
        }
        rows.push(fallback(e.MimeType))
        rows.push(fallback(e.FormatVersion))
        rows.push(formatValidity(e.Valid))
        rows.push([#e.NumberFiles])
        first = false
      }
      rows
    }).flatten(),
    [*Gesamt*],
    [],
    [],
    [],
    [*#fileStats.TotalFiles*],
  )
]

#let archivePackageColor(recordType) = {
  (
    file: rgb("#3f51b5"),
    process: rgb("#008000"),
    document: rgb("#ffa500"),
  ).at(recordType)
}

#let archivePackage(aipData) = [
  #box(
    stroke: 0.5pt + archivePackageColor(aipData.Type),
    fill: archivePackageColor(aipData.Type).transparentize(80%),
    table(
      columns: (auto, 1fr, auto, 1fr),
      stroke: none,
      table.cell(colspan: 4)[*#aipData.Title*],
      [Laufzeit:],
      formatLifetime(aipData.Lifetime),
      [],
      [],
      [Speicher-?volumen:],
      formatFileSize(aipData.TotalFileSize),
      [Paket-ID:],
      fallback(aipData.PackageID),
    ),
  )
]


#let archivePackagesInner(elements, level) = [
  // Sort AIPs before sub sections, so it is clear they don't belong to a sub section.
  #for el in elements [
    #if el.AIP != none [
      #archivePackage(el.AIP)
    ]
  ]
  #for el in elements [
    #if el.AIP == none [
      #heading(level: level + 1, el.Title)
      #archivePackagesInner(el.Children, level + 1)
    ]
  ]
]

#let archivePackages(elements) = [
  = Archivierte Pakete
  #archivePackagesInner(elements, 1)
]

#let report(data) = [
  #let title = [
    Übernahmebericht --
    #data.Process.agency.abbreviation -- E-Akte --
    #formatDate(data.Process.processState.archiving.completedAt)
  ]

  #set document(title: title)
  #set page(
    numbering: "1",
    margin: (x: 2cm),
    header: context {
      let (page,) = counter(page).at(here())
      if page > 1 {
        show sym.dash.en: "/"
        [#h(1fr) #title]
      }
    },
  )
  #set text(lang: "de", font: "Noto Sans", size: 10pt)

  #topMatter(data)
  #overview(data)
  #if data.Discrepancies.values().any(a => a != none) [
    #pagebreak()
    #discrepancies(data.Discrepancies)
  ]
  #pagebreak()
  #archivePackages(data.ArchivePackages)
  // #pagebreak()
  // #fileStats(data.FileStats)
]

#report(json("submission-data.json"))