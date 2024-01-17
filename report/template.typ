#set document(
  title: [Übernahmebericht],
)

#set page(
  numbering: "1",
)

#set text(
  lang: "de",
  font: "Noto Sans"
)


#let topMatter(data) = [
  #block(spacing: 2em)[
    #set text(2em)
    *Übernahmebericht*
  ]
  #table(
    columns: 2,
    inset: 0.5em,
    stroke: none,
    [Abgegebene Stelle], data.Institution,
    [Zeitpunkt der Abgabe], data.CreationTime
  )
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
    ..fileStats.ByFileType.pairs().map(pair => (
      raw(pair.at(0)),
      [#pair.at(1)]
    )).flatten(),
    [*Gesamt*], [*#fileStats.Total*],
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

#let report(data) = [
  #topMatter(data)
  #outline(
    indent: 2em
  )
  #fileStats(data.FileStats)
  #recordObjects(data.Message)
]

#report(json("data.json"))