#import "@preview/cetz:0.4.2"
#import "@preview/cetz-plot:0.1.3": chart
#import "shared.typ": formatDate, formatDateTime, formatContentStatsElement, appraisalStatsTable, appraisalStatsGraph

#let data = json("appraisal-data.json")

#let title = [
  Bewertungsbericht --
  #data.Process.agency.abbreviation -- E-Akte --
  #formatDate(data.Process.processState.appraisal.completedAt)
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

#let appraisalEntries(nodes, level) = {
  if (nodes == none) {
    ()
  } else {
    nodes
      .map(node => (
          table.cell(inset: (left: 1em * level))[
            #[
              #set text(weight: "bold") if node.AppraisalDecision == "A"
              #node.Title
            ]
            #if node.AppraisalNote != "" [
              #set text(size: 8pt)
              \ Bewertungsnotiz: #node.AppraisalNote
            ]
          ],
          table.cell(inset: (right: 0pt))[
            #set text(weight: "bold") if node.AppraisalDecision == "A"
            #show "V": "K"
            #node.AppraisalDecision
          ],
          ..appraisalEntries(node.Children, level + 1),
        ))
      .flatten()
  }
}

#let appraisals() = [
  = Bewertungen
  #table(
    columns: (1fr, auto),
    stroke: none,
    ..appraisalEntries(data.AppraisalInfo, 0)
  )
]



#topMatter()
#overview()
#pagebreak()
#appraisals()