class Event {
  index int
  taken bool
}

function Summarize(events: Event[], name: string) -> string {
  client Fast
  prompt #"
    <instructions role="system">Summarize the events for {{ name }}.</instructions>
    {% for event in events %}
    - event {{ event.index }} was {% if event.taken %}taken{% else %}skipped{% endif %}
    {% endfor %}
    {{ ctx.output_format }}
  "#
}

function Ask(text: string) -> string {
  client Fast
  prompt `Summarize \`this\`: ${text} for $5
${ctx.output_format()}`
}
