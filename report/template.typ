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
  #let values = dateString.split(regex("[-T:.]"))
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
      [#formatContentStatsElement("Files", data.Message0503Stats.Files)],
    )
    #v(10fr)
  ] else [
    #table(
      columns: (1fr, 1fr, 1fr),
      stroke: none,
      [*Angeboten*],
      [*Übernommen*],
      [*Kassiert*],
      [#formatContentStatsElement("Files", data.AppraisalStats.Files.Total)],
      [#formatContentStatsElement("Files", data.AppraisalStats.Files.Archived)],
      [#formatContentStatsElement("Files", data.AppraisalStats.Files.Discarded)],
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
          let values = ()
          if (data.AppraisalStats.Files.Archived > 0) {
            values.push(([übernommen], data.AppraisalStats.Files.Archived, (fill: olive)))
          }
          if (data.AppraisalStats.Files.Discarded > 0) {
            values.push((
              [kassiert],
              data.AppraisalStats.Files.Discarded,
              (fill: rgb("#e53a31")),
            ))
          }
          cetz.chart.piechart(
            values,
            label-key: 0,
            value-key: 1,
            radius: 4,
            slice-style: (
              // slice-style as a somewhat peculiar indexing strategy...
              index => values.at(calc.rem-euclid(values.len() - index - 1, values.len())).at(2)
            ),
            inner-label: (content: (value, label) => [#text(white, label)], radius: 120%),
            outer-label: (content: "%", radius: 120%),
          )
        },
      )
      #v(2fr)
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

#let archivePackages(elements) = [
  = Archivierte Pakete
  #for el in elements [
    == #el.Title

    #table(
      columns: (auto, 1fr, auto, 1fr),
      stroke: none,
      [Laufzeit:],
      formatLifetime(el.Lifetime),
      [],
      [],
      [Speicher-?volumen:],
      formatFilesize(el.TotalFileSize),
      [Paket-ID:],
      fallback(el.PackageID),
    )
    #table(
      columns: 2,
      stroke: none,
      [Bewertungs-?notiz:],
      [#fallback(el.AppraisalNote)],
    )
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
  #pagebreak()
  #archivePackages(data.ArchivePackages)
  // #pagebreak()
  // #fileStats(data.FileStats)
]

#report(json("data.json"))