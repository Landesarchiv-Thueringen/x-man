#import "@preview/cetz:0.3.1"
#import "@preview/cetz-plot:0.1.0": chart
#import "shared.typ": formatDate, formatDateTime, formatContentStatsElement, appraisalStatsTable, appraisalStatsGraph

#let data = json("appraisal-data.json")

#let topMatter() = [
  #block(spacing: 2em)[
    #set text(2em)
    *Bewertungsbericht*
  ]
  #table(
    columns: 2,
    stroke: none,
    [Abgegebene Stelle:], data.Process.agency.name,
    [Objektart:], [E-Akte],
    [Prozess-ID:], data.Process.processId,
    [Anbietung erhalten:], formatDateTime(data.Process.processState.receive0501.completedAt),
    [Bewertung versendet:], formatDateTime(data.Process.processState.appraisal.completedAt),
    [Bewertung durch:], data.Process.processState.appraisal.completedBy,
  )
]

#let overview() = [
  #v(1fr)
  = Übersicht
  #appraisalStatsTable(data)
  #v(1fr)
  #appraisalStatsGraph(data)
  #v(2fr)
]

#let title = [
  Bewertungsbericht --
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

#topMatter()
#overview()