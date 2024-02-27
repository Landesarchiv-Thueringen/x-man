#set document(
  title: [Übernahmebericht],
)

#set page(
  numbering: "1",
  // margin: (
  //   top: 2cm,
  //   bottom: 2cm,
  //   x: 1cm,
  // )
)

#set text(
  lang: "de",
  font: "Noto Sans"
)

#let fallbackEmpty(string, fallback: "-") = {
  if string == "" or string == none { fallback } else { string }
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

#let formatAppraisalCode(code) = [
  #[
    #show "A": [archivieren]
    #show "V": [vernichten]
    #code
  ] (#code)
]

#let topMatter(data) = [
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
      columns: (1fr, 1fr, 1fr, 1fr, 1fr, 1fr),
      inset: 0.5em,
      align: (x, y) => (left, right).at(calc.rem(x, 2)),
      stroke: none,
      [Akten:], [#data.Message0501Stats.TotalFiles],
      [Vorgänge:], [#data.Message0501Stats.TotalProcesses],
      [Dokumente:], [#data.Message0501Stats.TotalDocuments],
      [Unterakten:], [#data.Message0501Stats.TotalSubFiles],
      [Untervorgänge:], [#data.Message0501Stats.TotalSubProcesses],
      [Anhänge:], [#data.Message0501Stats.TotalAttachments],
    )
  ]
  #if data.AppraisalStats != none [
    == Bewertung
    #table(
      columns: (1fr, 1fr, 1fr, 1fr, 1fr, 1fr),
      inset: 0.5em,
      align: (x, y) => (left, right).at(calc.rem(x, 2)),
      stroke: none,
      [*Akten*], [],
      [*Vorgänge*], [],
      [*Dokumente*], [],
      block(inset: (left: 1em))[archivieren:], [#data.AppraisalStats.Files.Archived],
      block(inset: (left: 1em))[archivieren:], [#data.AppraisalStats.Processes.Archived],
      block(inset: (left: 1em))[archivieren:], [#data.AppraisalStats.Documents.Archived],
      block(inset: (left: 1em))[kassieren:], [#data.AppraisalStats.Files.Discarded],
      block(inset: (left: 1em))[kassieren:], [#data.AppraisalStats.Processes.Discarded],
      block(inset: (left: 1em))[kassieren:], [#data.AppraisalStats.Documents.Discarded],
      [*Unterakten*], [],
      [*Untervorgänge*], [],
      [*Anhänge*], [],
      block(inset: (left: 1em))[archivieren:], [#data.AppraisalStats.SubFiles.Archived],
      block(inset: (left: 1em))[archivieren:], [#data.AppraisalStats.SubProcesses.Archived],
      block(inset: (left: 1em))[archivieren:], [#data.AppraisalStats.Attachments.Archived],
      block(inset: (left: 1em))[kassieren:], [#data.AppraisalStats.SubFiles.Discarded],
      block(inset: (left: 1em))[kassieren:], [#data.AppraisalStats.SubProcesses.Discarded],
      block(inset: (left: 1em))[kassieren:], [#data.AppraisalStats.Attachments.Discarded],
    )
  ]
  == Übernahme
  #table(
    columns: (1fr, 1fr, 1fr, 1fr, 1fr, 1fr),
    inset: 0.5em,
    align: (x, y) => (left, right).at(calc.rem(x, 2)),
    stroke: none,
    [Akten:], [#data.Message0503Stats.TotalFiles],
    [Vorgänge:], [#data.Message0503Stats.TotalProcesses],
    [Dokumente:], [#data.Message0503Stats.TotalDocuments],
    [Unterakten:], [#data.Message0503Stats.TotalSubFiles],
    [Untervorgänge:], [#data.Message0503Stats.TotalSubProcesses],
    [Anhänge:], [#data.Message0503Stats.TotalAttachments],
  )
  #table(
    columns: (1fr, 1fr, 1fr, 1fr, 1fr, 1fr),
    inset: 0.5em,
    align: (x, y) => (left, right).at(calc.rem(x, 2)),
    stroke: none,
    [Dateien:], [#data.FileStats.TotalFiles],
    [Gesamtgröße:], [#formatFilesize(data.FileStats.TotalBytes)]
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

#let fileRecordObjectsTable(fileRecordObjects, level) = [
  #if fileRecordObjects.len() > 0 [
    #heading(level: level)[
      Akten (#fileRecordObjects.len())
    ]
    #table(
      fill: rgb("#3f51b520"),
      columns: (auto, 1fr, auto),
      [*Aktenzeichen*], [*Betreff*], [*Bewertung*],
      ..fileRecordObjects.map(f => (
        f.generalMetadata.xdomeaID,
        f.generalMetadata.subject,
        f.archiveMetadata.appraisalCode,
      )).flatten()
    )
  ]
]

#let processRecordObjectsTable(processRecordObjects, level) = [
  #if processRecordObjects.len() > 0 [
    #heading(level: level)[
      Vorgänge (#processRecordObjects.len())
    ]
    #table(
      fill: rgb("#00800020"),
      columns: (auto, 1fr, auto),
      [*Aktenzeichen*], [*Betreff*], [*Bewertung*],
      ..processRecordObjects.map(f => (
        f.generalMetadata.xdomeaID,
        f.generalMetadata.subject,
        f.archiveMetadata.appraisalCode,
      )).flatten()
    )
  ]
]

#let documentRecordObjectsTable(documentRecordObjects, level) = [
  #if documentRecordObjects.len() > 0 [
    #heading(level: level)[
      Dokumente (#documentRecordObjects.len())
    ]
    #table(
      fill: rgb("#ffa50020"),
      columns: (auto, 1fr),
      [*Aktenzeichen*], [*Betreff*], 
      ..documentRecordObjects.map(f => (
        f.generalMetadata.xdomeaID,
        f.generalMetadata.subject,
      )).flatten()
    )
  ]
]

#let processRecordObjects(processRecordObjects, level) = [
  #for p in processRecordObjects [
    #if p.archiveMetadata.appraisalCode == "A" [
      #heading(level: level, text(rgb("#008000"))[#p.generalMetadata.xdomeaID #p.generalMetadata.subject])
      #processRecordObjectsTable(p.subprocesses, level + 1)
      #documentRecordObjectsTable(p.documents, level + 1)
    ]
  ]
]

#let fileRecordObjects(fileRecordObjects, level) = [
  #for f in fileRecordObjects [
    #if f.archiveMetadata.appraisalCode == "A" [
      #heading(level: level, text(rgb("#3f51b5"))[#f.generalMetadata.xdomeaID #f.generalMetadata.subject])
      #fileRecordObjectsTable(f.subfiles, level + 1)
      #processRecordObjectsTable(f.processes, level + 1)
      #processRecordObjects(f.processes, level + 1)
    ]
  ]
]

#let recordObjects(message) = [
  #let rootLevel = 1
  #fileRecordObjectsTable(message.fileRecordObjects, rootLevel)
  #processRecordObjectsTable(message.processRecordObjects, rootLevel)
  #documentRecordObjectsTable(message.documentRecordObjects, rootLevel)
  #fileRecordObjects(message.fileRecordObjects, rootLevel)
  #processRecordObjects(message.processRecordObjects, rootLevel)
]




#let contents(message) = [
  // #let fileCounter = counter("files")
  = Inhalte
  // === Akten (#message.fileRecordObjects.len())
  #for f in message.fileRecordObjects [
    // #fileCounter.step()
    == Akte #f.generalMetadata.xdomeaID: #f.generalMetadata.subject
    #table(
      columns: 2,
      inset: 0.5em,
      // align: (x, y) => (left, right).at(calc.rem(x, 2)),
      stroke: none,
      [Laufzeit:], [#formatDate(f.lifetime.start) -- #formatDate(f.lifetime.end)],
      [Bewertung:], [#formatAppraisalCode(f.archiveMetadata.appraisalCode)],
      [Bewertungsnotiz:], [#fallbackEmpty(f.archiveMetadata.internalAppraisalNote)],
      [Umfang:], [TODO],
      [Speichervolumen:], [TODO],
      [Signatur:], [TODO],
    )
    
  ]
]

#let report(data) = [
  #topMatter(data)
  #overview(data)
  #pagebreak()
  // #recordObjects(data.Message0503)
  #contents(data.Message0501)
  #pagebreak()
  #fileStats(data.FileStats)
]

#report(json("data.json"))