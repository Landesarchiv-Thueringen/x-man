#import "@preview/cetz:0.2.1"

#let fallback(input, fallback: "-") = {
  if input == "" or input == none or input == 0 { fallback } else { input }
}

#let formatDate(dateString) = [
  #let values = dateString.split(regex("[-T]"))
  #let date = datetime(
    year: int(values.at(0)),
    month: int(values.at(1)),
    day: int(values.at(2)),
  )
  #date.display("[day].[month].[year]")
]

#let formatDateTime(dateString) = [
  #let values = dateString.split(regex("[-T:.Z]"))
  #let date = datetime(
    year: int(values.at(0)),
    month: int(values.at(1)),
    day: int(values.at(2)),
    hour: int(values.at(3)),
    minute: int(values.at(4)),
    second: int(values.at(5)),
  )
  #date.display("[day].[month].[year] [hour]:[minute] Uhr")
]

#let formatFloat(f, digitsAfterPoint) = {
  let factor = calc.pow(10, digitsAfterPoint)
  let beforePoint = calc.floor(f)
  let afterPoint = f - beforePoint
  str(beforePoint) + "," + str(calc.round(afterPoint * factor))
}

#let formatFilesize(nbytes) = {
  let v = nbytes
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

#let formatContentStatsElement(type, number) = {
  if number == 1 {
    "1 " + (
      Files: "Akte",
      SubFiles: "Teilakte",
      Processes: "Vorgang",
      SubProcesses: "Teilvorgang",
      Documents: "Dokument",
      Attachments: "Anhang",
    ).at(type)
  } else {
    str(number) + " " + (
      Files: "Akten",
      SubFiles: "Teilakten",
      Processes: "Vorgänge",
      SubProcesses: "Teilvorgänge",
      Documents: "Dokumente",
      Attachments: "Anähnge",
    ).at(type)
  }
}

#let formatAppraisalCode(code) = (A: "Archivieren", V: "Vernichten").at(code)

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
      columns: (1fr),
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
      [Gesamt-?speicher-?volumen übernommen:],
      [#formatFilesize(data.FileStats.TotalBytes)],
    )
    #v(10fr)
  ] else [
    #let columns = ()
    #let all = (
      data.AppraisalStats.Files,
      data.AppraisalStats.Processes,
      data.AppraisalStats.Documents,
    )
    #columns.push((label: "Angeboten", key: "Offered"))
    #if all.any(el => el.PartiallyArchived > 0) [
      #columns.push((label: "Vollständig übernommen", key: "Archived"))
      #columns.push((label: "Teilweise übernommen", key: "PartiallyArchived"))
    ] else [
      #columns.push((label: "Übernommen", key: "Archived"))
    ]
    #columns.push((label: "Kassiert", key: "Discarded"))
    #if all.any(el => el.Missing > 0) [
      #columns.push((label: "Fehlend", key: "Missing"))
    ]
    #if all.any(el => el.Surplus > 0) [
      #columns.push((label: "Zusätzlich", key: "Surplus"))
    ]
    #table(
      columns: range(columns.len()).map(_ => 1fr),
      stroke: none,
      ..columns.map(c => [*#c.label*]),
      ..if data.AppraisalStats.Files.Total > 0 {
        columns.map(
          c => [#formatContentStatsElement("Files", data.AppraisalStats.Files.at(c.key))],
        )
      },
      ..if data.AppraisalStats.Processes.Total > 0 {
        columns.map(
          c => [#formatContentStatsElement("Processes", data.AppraisalStats.Processes.at(c.key))],
        )
      },
      ..if data.AppraisalStats.Documents.Total > 0 {
        columns.map(
          c => [#formatContentStatsElement("Documents", data.AppraisalStats.Documents.at(c.key))],
        )
      },
    )

    #table(
      columns: 2,
      stroke: none,
      [Gesamt-?speicher-?volumen übernommen:],
      [#formatFilesize(data.FileStats.TotalBytes)],
    )

    #[
      #v(1fr)
      #set align(center)
      #cetz.canvas(
        {
          let values = ((
            label: [übernommen],
            value: all.map(el => el.Archived).sum(),
            backgroundColor: rgb("#005cbb"),
            textColor: rgb("#ffffff"),
          ), (
            label: [teilweise],
            value: all.map(el => el.PartiallyArchived).sum(),
            backgroundColor: rgb("#0074e9"),
            textColor: rgb("#ffffff"),
          ), (
            label: [kassiert],
            value: all.map(el => el.Discarded).sum(),
            backgroundColor: rgb("#d7e3ff"),
            textColor: rgb("#410002"),
          ), (
            label: [fehlend],
            value: all.map(el => el.Missing).sum(),
            backgroundColor: rgb("#ffdad6"),
            textColor: rgb("#001b3f"),
          ), (
            label: [zusätzlich],
            value: all.map(el => el.Surplus).sum(),
            backgroundColor: rgb("#ba1a1a"),
            textColor: rgb("#ffffff"),
          )).filter(v => v.value > 0)
          cetz.chart.piechart(
            values,
            label-key: "label",
            value-key: "value",
            radius: 4,
            slice-style: (
              index => (
                // slice-style has a somewhat peculiar indexing strategy...
                fill: values.at(calc.rem-euclid(values.len() - index - 1, values.len())).backgroundColor,
                stroke: none,
              )
            ),
            inner-label: (
              content: (value, label) => [#text(values.find(v => v.label == label).textColor, label)],
              radius: 120%,
            ),
            outer-label: (content: "%", radius: 120%),
          )
        },
      )
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
      formatFilesize(aipData.TotalFileSize),
      [Paket-ID:],
      fallback(aipData.PackageID),
    ),
  )
]

#let archivePackageSection(structureData, level) = [
  #heading(level: level + 1, structureData.Title)
  #if structureData.AppraisalNote != "" [
    Bewertungsnotiz: #structureData.AppraisalNote
  ]

  #for el in structureData.Children [
    #if el.AIP == none [
      #archivePackageSection(el, level + 1)
    ] else [
      #archivePackage(el.AIP)
      #if el.AppraisalNote != "" [
        Bewertungsnotiz: #el.AppraisalNote
      ]
    ]
  ]
]

#let archivePackages(elements) = [
  = Archivierte Pakete
  #for el in elements [
    #if el.AIP == none [
      #archivePackageSection(el, 1)
    ] else [
      #archivePackage(el.AIP)
      #if el.AppraisalNote != "" [
        Bewertungsnotiz: #el.AppraisalNote
      ]
    ]
  ]
]

#let report(data) = [
  #let title = [
    Übernahmebericht --
    #data.Process.agency.abbreviation -- E-Akte --
    #formatDate(data.Process.processState.archiving.completedAt)
  ]

  #set document(title: title)
  #set page(numbering: "1", margin: (x: 2cm), header: locate(loc => {
    let (page,) = counter(page).at(loc)
    if page > 1 {
      show sym.dash.en: "/"
      [#h(1fr) #title]
    }
  }))
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

#report(json("data.json"))