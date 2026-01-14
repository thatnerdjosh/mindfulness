# Mindfulness (mt)

CLI journal for daily reflections on the Five Mindfulness Trainings.

## Design

* [X] - Multiple entries (format below) per-day, stored in `$XDG_DATA_DIR/mt/journal.json`
```json
[
	{
		"date": "YYYY-MM-DD",
		"reflections": {
			"1st precept": "",
			...
		}
	}
	...
]
```
* [X] - Ability to keep track of precept adherence, stored in `$XDG_DATA_DIR/mt/adherence.json` - adherence to precepts defaults to `true`

```json
{
	"1st precept": true,
	...
}
```

* [X] - Log file when adherence is modified (true <-> false)
* [X] - Guided adherence toggle interface (defaults will be based on the current adherence)
* [X] - Optional journal note when any of the adherences are toggled (specifically for the precept which is toggled)
