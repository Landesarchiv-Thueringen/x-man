#set document(
  title: [Übernahmebericht],
)

#set page(
  numbering: "1",
  margin: 1.5cm,
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


#let formatAppraisalCode(code) = [
  #[
    #show "A": [archivieren]
    #show "V": [vernichten]
    #code
  ] (#code)
]

#let topMatter(data) = [
  #v(2em)
  #block(spacing: 2em)[
    #set text(2em)
    *Übernahmebericht*
  ]
  #table(
    columns: 2,
    inset: 0.5em,
    stroke: none,
    [Abgegebene Stelle], data.Process.institution,
    [Objektart], [E-Akte],
    [Aussonderungs-ID], data.Process.xdomeaID,
    [Anbietung erhalten], 
    if data.Process.processState.receive0501.complete {
      formatDateTime(data.Process.processState.receive0501.completionTime)
    } else [-],
    [Bewertung versendet], formatDateTime(data.Process.processState.appraisal.completionTime),
    [Bewertung durch], data.Process.processState.appraisal.completedBy,
    [Abgabe archiviert], formatDateTime(data.Process.processState.archiving.completionTime),
    [Archivierung durch], data.Process.processState.archiving.completedBy,
  )
]

#let overview(data) = [
  = Übersicht
  #if data.Message0501Stats != none [
    == Anbietung
    #table(
      columns: (1fr, 1fr, 1fr, 1fr, 1fr, 1fr, 1fr),
      inset: 0.5em,
      stroke: none,
      [],
      [*Akten*],
      [*Teilakten*],
      [*Vorgänge*],
      [*Teilvorgänge*],
      [*Dokumente*],
      [*Anhänge*],
      [],
      [#fallback(data.Message0501Stats.Files)],
      [#fallback(data.Message0501Stats.SubFiles)],
      [#fallback(data.Message0501Stats.Processes)],
      [#fallback(data.Message0501Stats.SubProcesses)],
      [#fallback(data.Message0501Stats.Documents)],
      [#fallback(data.Message0501Stats.Attachments)],
    )
  ]
  #if data.AppraisalStats != none [
    == Bewertung
    #table(
      columns: (1fr, 1fr, 1fr, 1fr, 1fr, 1fr, 1fr),
      inset: 0.5em,
      stroke: none,
      [],
      [*Akten*],
      [*Teilakten*],
      [*Vorgänge*],
      [*Teilvorgänge*],
      [*Dokumente*],
      [*Anhänge*],
      [*Archivieren*],
      [#fallback(data.AppraisalStats.Files.Archived)],
      [#fallback(data.AppraisalStats.SubFiles.Archived)],
      [#fallback(data.AppraisalStats.Processes.Archived)],
      [#fallback(data.AppraisalStats.SubProcesses.Archived)],
      [#fallback(data.AppraisalStats.Documents.Archived)],
      [#fallback(data.AppraisalStats.Attachments.Archived)],
      [*Kassieren*],
      [#fallback(data.AppraisalStats.Files.Discarded)],
      [#fallback(data.AppraisalStats.SubFiles.Discarded)],
      [#fallback(data.AppraisalStats.Processes.Discarded)],
      [#fallback(data.AppraisalStats.SubProcesses.Discarded)],
      [#fallback(data.AppraisalStats.Documents.Discarded)],
      [#fallback(data.AppraisalStats.Attachments.Discarded)],
    )
  ]
  == Übernahme
  #table(
    columns: (1fr, 1fr, 1fr, 1fr, 1fr, 1fr, 1fr),
    inset: 0.5em,
    stroke: none,
    [],
    [*Akten*],
    [*Teilakten*],
    [*Vorgänge*],
    [*Teilvorgänge*],
    [*Dokumente*],
    [*Anhänge*],
    [],
    [#fallback(data.Message0503Stats.Files)],
    [#fallback(data.Message0503Stats.SubFiles)],
    [#fallback(data.Message0503Stats.Processes)],
    [#fallback(data.Message0503Stats.SubProcesses)],
    [#fallback(data.Message0503Stats.Documents)],
    [#fallback(data.Message0503Stats.Attachments)],
  )
  #table(
    columns: (1fr, 1fr, 1fr, 1fr, 1fr, 1fr, 1fr),
    inset: 0.5em,
    stroke: none,
    [], [*Dateien*], [*Gesamtgröße*], [], [], [],[],
    [], [#data.FileStats.TotalFiles], [#formatFilesize(data.FileStats.TotalBytes)],
  )
  == Archivierung
  TODO
]

#let fileStats(fileStats) = [
  = Datei-Statistik
  #set align(center)
  #table(
    columns: 2,
    align: (x, y) => (left, right).at(x),
    inset: 0.5em,
    stroke: none,
    [*Dateityp*], [*Archivierte Dateien*],
    ..fileStats.FilesByFileType.pairs().map(pair => (
      raw(pair.at(0)),
      [#pair.at(1)]
    )).flatten(),
    [*Gesamt*], [*#fileStats.TotalFiles*],
  )
]


#let contentList(elements, level) = [
  #for el in elements [
    #heading(
      level: level
    )[#formatRecordObjectType(el.recordObjectType) #el.generalMetadata.xdomeaID: #el.generalMetadata.subject]
    #[
      #set block(above: 0.2em)
      #table(
        columns: (1fr),
        inset: 0.5em,
        stroke: none,
        [*Laufzeit*],
        [#formatDate(el.lifetime.start) -- #formatDate(el.lifetime.end)],
      )
      #table(
        columns: (1fr, 5fr),
        inset: 0.5em,
        stroke: none,
        [*Bewertung*],
        [*Bewertungsnotiz*],
        [#formatAppraisalCode(el.archiveMetadata.appraisalCode)],
        [#fallback(el.archiveMetadata.internalAppraisalNote)],
      )
      #if el.recordObjectType == "file" or el.recordObjectType == "subFile" [
        #if  el.children == none [
          #table(
            columns: (1fr, 1fr, 1fr, 1fr, 1fr, 1fr),
            inset: 0.5em,
            stroke: none,
            [*Teilakten*],
            [*Vorgänge*],
            [*Teilvorgänge*],
            [*Dokumente*],
            [*Anhänge*],
            [],
            [#fallback(el.contentStats.SubFiles.Total)],
            [#fallback(el.contentStats.Processes.Total)],
            [#fallback(el.contentStats.SubProcesses.Total)],
            [#fallback(el.contentStats.Documents.Total)],
            [#fallback(el.contentStats.Attachments.Total)],
          )
        ] else [
          #table(
            columns: (1fr, 1fr, 1fr, 1fr, 1fr, 1fr),
            inset: 0.5em,
            stroke: none,
            [],
            [*Teilakten*],
            [*Vorgänge*],
            [*Teilvorgänge*],
            [*Dokumente*],
            [*Anhänge*],
            [*Gesamt*],
            [#fallback(el.contentStats.SubFiles.Total)],
            [#fallback(el.contentStats.Processes.Total)],
            [#fallback(el.contentStats.SubProcesses.Total)],
            [#fallback(el.contentStats.Documents.Total)],
            [#fallback(el.contentStats.Attachments.Total)],
            [*Archivieren*],
            [#fallback(el.contentStats.SubFiles.Archived)],
            [#fallback(el.contentStats.Processes.Archived)],
            [#fallback(el.contentStats.SubProcesses.Archived)],
            [#fallback(el.contentStats.Documents.Archived)],
            [#fallback(el.contentStats.Attachments.Archived)],
            [*Kassieren*],
            [#fallback(el.contentStats.SubFiles.Discarded)],
            [#fallback(el.contentStats.Processes.Discarded)],
            [#fallback(el.contentStats.SubProcesses.Discarded)],
            [#fallback(el.contentStats.Documents.Discarded)],
            [#fallback(el.contentStats.Attachments.Discarded)],
          )
        ]
      ]
      // TODO: handle processes
      #table(
        columns: (1fr, 1fr),
        inset: 0.5em,
        stroke: none,
        [*Speichervolumen*],
        [*Signatur*],
        [TODO],
        [TODO],
      )
    ]
    // #table(
    //   columns: 2,
    //   inset: 0.5em,
    //   stroke: none,
    //   [Laufzeit:], [#formatDate(el.lifetime.start) -- #formatDate(el.lifetime.end)],
    //   [Bewertung:], [#formatAppraisalCode(el.archiveMetadata.appraisalCode)],
    //   [Bewertungsnotiz:], [#fallback(el.archiveMetadata.internalAppraisalNote)],
    //   [Umfang:], [TODO],
    //   [Speichervolumen:], [TODO],
    //   [Signatur:], [TODO],
    // )
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