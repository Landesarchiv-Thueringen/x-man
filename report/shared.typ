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
      Attachments: "Anähnge",
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
]

#let appraisalStatsGraph(data) = [
   #set align(center)
      #cetz.canvas(
        {
          let all = (
            data.AppraisalStats.Files,
            data.AppraisalStats.Processes,
            data.AppraisalStats.Documents,
          )
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
          chart.piechart(
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
            legend: (label: none),
          )
        },
      )
]