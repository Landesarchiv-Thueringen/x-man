#import "@preview/cetz:0.3.1"
#import "@preview/cetz-plot:0.1.0": chart

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
      Attachments: "Anhänge",
    ).at(type)
  }
}

#let appraisalStatsTable(data) = [
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
      columns.map(c => [#formatContentStatsElement("Files", data.AppraisalStats.Files.at(c.key))])
    },
    ..if data.AppraisalStats.Processes.Total > 0 {
      columns.map(c => [#formatContentStatsElement("Processes", data.AppraisalStats.Processes.at(c.key))])
    },
    ..if data.AppraisalStats.Documents.Total > 0 {
      columns.map(c => [#formatContentStatsElement("Documents", data.AppraisalStats.Documents.at(c.key))])
    },
  )
]

#let appraisalStatsGraph(data) = [
  #set align(center)
  #cetz.canvas({
    let all = (
      data.AppraisalStats.Files,
      data.AppraisalStats.Processes,
      data.AppraisalStats.Documents,
    )
    let total = all.map(el => el.Total).sum()
    let values = (
      (
        label: [übernommen],
        value: all.map(el => el.Archived).sum(),
        backgroundColor: rgb("#005cbb"),
        textColor: rgb("#ffffff"),
      ),
      (
        label: [teilweise],
        value: all.map(el => el.PartiallyArchived).sum(),
        backgroundColor: rgb("#0074e9"),
        textColor: rgb("#ffffff"),
      ),
      (
        label: [kassiert],
        value: all.map(el => el.Discarded).sum(),
        backgroundColor: rgb("#d7e3ff"),
        textColor: rgb("#000000"),
      ),
      (
        label: [fehlend],
        value: all.map(el => el.Missing).sum(),
        backgroundColor: rgb("#ffdad6"),
        textColor: rgb("#000000"),
      ),
      (
        label: [zusätzlich],
        value: all.map(el => el.Surplus).sum(),
        backgroundColor: rgb("#ba1a1a"),
        textColor: rgb("#ffffff"),
      ),
    ).filter(v => v.value > 0)
    chart.piechart(
      values,
      label-key: "label",
      value-key: "value",
      radius: 4,
      inner-radius: 1,
      slice-style: (
        index => (
          fill: values.at(index).backgroundColor,
          stroke: none,
        )
      ),
      inner-label: (
        content: (value, label) => if value / total >= 0.2 [
          #set text(fill: values.find(v => v.label == label).textColor)
          #set align(center)
          #label \
          (#(calc.round(value / total * 100))%)
        ],
        radius: 100%,
      ),
      outer-label: (
        content: (value, label) => if value / total < 0.2 [
          #label (#(calc.round(value / total * 100))%)
        ],
        radius: 110%,
        anchor: "west",
      ),
      legend: (label: none),
      start: 0deg,
      stop: 360deg,
      clockwise: false,
    )
  })
]