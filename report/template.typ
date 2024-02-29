#set document(
  title: [Übernahmebericht],
)

#set page(
  numbering: "1",
  margin: (x: 2cm),
)

#set text(
  lang: "de",
  font: "Noto Sans",
  size: 10pt,
)

#let fallback(input, fallback: "-") = {
  if input == "" or input == none or input == 0 { fallback } else { input }
}

#let formatDate(dateString) = [
    #let values = dateString.split(regex("[-]"))
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

#let formatRecordObjectType(type) = (
  file: "Akte",
  subFile: "Teilakte",
  process: "Vorgang",
  subProcess: "Teilvorgang",
).at(type)

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

#let formatContentStats(stats) = [
  #let statsList = (
    "Files", "SubFiles", "Processes", "SubProcesses", "Documents",
  ).map(key => (key, stats.at(key))).filter(el => el.last().Total > 0)
  #if stats.HasDeviatingAppraisals [
    #statsList.map(el => [
      #formatContentStatsElement(el.first(), el.last().Total)
      (davon #el.last().Archived übernommen, #el.last().Discarded kassiert)
    ]).join([\
    ])
  ] else [
    #statsList.map(el => formatContentStatsElement(el.first(), el.last().Total)).join(", ")
  ]
]

#let formatAppraisalCode(code) = (
  A: "Archivieren",
  V: "Vernichten",
).at(code)

#let formatValidity(validity) = {
  if (validity == none) { "-" }
  else if (validity) { "valide" }
  else { "invalide" }
}

#let topMatter(data) = [
  // #v(2em)
  #block(spacing: 2em)[
    #set text(2em)
    *Übernahmebericht*
  ]
  #table(
    columns: 2,
    stroke: none,
    [Abgegebene Stelle:], data.Process.institution,
    [Objektart:], [E-Akte],
    [Aussonderungs-ID:], data.Process.xdomeaID,
    [Anbietung erhalten:], 
    if data.Process.processState.receive0501.complete {
      formatDateTime(data.Process.processState.receive0501.completionTime)
    } else [-],
    [Bewertung versendet:], formatDateTime(data.Process.processState.appraisal.completionTime),
    [Bewertung durch:], data.Process.processState.appraisal.completedBy,
    [Abgabe archiviert:], formatDateTime(data.Process.processState.archiving.completionTime),
    [Archivierung durch:], data.Process.processState.archiving.completedBy,
  )
]

#let overview(data) = [
  = Übersicht
  #if data.Message0501Stats != none [
    == Anbietung
    #table(
      columns: (1fr, 1fr, 1fr, 1fr, 1fr, 1fr),
      stroke: none,
      [],
      [*Akten*],
      [*Teilakten*],
      [*Vorgänge*],
      [*Teilvorgänge*],
      [*Dokumente*],
      [],
      [#fallback(data.Message0501Stats.Files)],
      [#fallback(data.Message0501Stats.SubFiles)],
      [#fallback(data.Message0501Stats.Processes)],
      [#fallback(data.Message0501Stats.SubProcesses)],
      [#fallback(data.Message0501Stats.Documents)],
    )
  ]
  #if data.AppraisalStats != none [
    == Bewertung
    #table(
      columns: (1fr, 1fr, 1fr, 1fr, 1fr, 1fr),
      stroke: none,
      [],
      [*Akten*],
      [*Teilakten*],
      [*Vorgänge*],
      [*Teilvorgänge*],
      [*Dokumente*],
      [*Archivieren*],
      [#fallback(data.AppraisalStats.Files.Archived)],
      [#fallback(data.AppraisalStats.SubFiles.Archived)],
      [#fallback(data.AppraisalStats.Processes.Archived)],
      [#fallback(data.AppraisalStats.SubProcesses.Archived)],
      [#fallback(data.AppraisalStats.Documents.Archived)],
      [*Kassieren*],
      [#fallback(data.AppraisalStats.Files.Discarded)],
      [#fallback(data.AppraisalStats.SubFiles.Discarded)],
      [#fallback(data.AppraisalStats.Processes.Discarded)],
      [#fallback(data.AppraisalStats.SubProcesses.Discarded)],
      [#fallback(data.AppraisalStats.Documents.Discarded)],
    )
  ]
  == Übernahme
  #table(
    columns: (1fr, 1fr, 1fr, 1fr, 1fr, 1fr),
    stroke: none,
    [],
    [*Akten*],
    [*Teilakten*],
    [*Vorgänge*],
    [*Teilvorgänge*],
    [*Dokumente*],
    [],
    [#fallback(data.Message0503Stats.Files)],
    [#fallback(data.Message0503Stats.SubFiles)],
    [#fallback(data.Message0503Stats.Processes)],
    [#fallback(data.Message0503Stats.SubProcesses)],
    [#fallback(data.Message0503Stats.Documents)],
  )
  #table(
    columns: (1fr, 1fr, 1fr, 1fr, 1fr, 1fr),
    stroke: none,
    [], [*Dateien*], [*Gesamtgröße*], [], [], [],
    [], [#data.FileStats.TotalFiles], [#formatFilesize(data.FileStats.TotalBytes)],
  )
  == Archivierung
  TODO
]

#let fileStats(fileStats) = [
  = Primärdateien
  #table(
    columns: 5,
    stroke: none,
    align: (x, y) => (if x == 4 and  y > 0 { right } else { left }),
    [*PUID*], [*MIME-Type*], [*Formatversion*], [*Validität*], [*Dateien*],
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
    [*Gesamt*], [], [], [], [*#fileStats.TotalFiles*],
  )
]


#let contentList(elements, level) = [
  // #set text(size: 8pt)
  #for el in elements [
    #heading(
      level: level
    )[
      #formatRecordObjectType(el.recordObjectType): #el.generalMetadata.xdomeaID
      (#el.archiveMetadata.appraisalCode)\
      #el.generalMetadata.subject
    ]
    #table(
      columns: (auto, 1fr, auto, 1fr),
      stroke: none,
      [Laufzeit:], [#formatDate(el.lifetime.start) -- #formatDate(el.lifetime.end)],
      [Umfang:], [#formatContentStats(el.contentStats)],
      [Bewertung:], [#formatAppraisalCode(el.archiveMetadata.appraisalCode)],
      [Bewertungs-?notiz:], [#fallback(el.archiveMetadata.internalAppraisalNote)],
      [Speicher-?volumen:],
      [TODO],
      [Signatur:],
      [TODO],
    )
    #if el.children != none [
      #block(inset: (left: 2.4em))[  
        #contentList(el.children, level + 1)
      ]
    ]
  ]
]

#let contents(elements) = [
  = Inhalte
  #let rootLevel = 2
  #contentList(elements, rootLevel)
]

#let report(data) = [
  #topMatter(data)
  #overview(data)
  #pagebreak()
  // #recordObjects(data.Message0503)
  #contents(data.Content)
  #pagebreak()
  #fileStats(data.FileStats)
]

#report(json("data.json"))