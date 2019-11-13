package sim

type iInputReproducer interface {
	next() (int, float64, string, float64, float64)
}

type inputReproducer struct {
	index   int
	warmed  bool
	entries []InputEntry
}

type warmedinputReproducer struct {
	index   int
	entries []InputEntry
}

func newInputReproducer(input []InputEntry) iInputReproducer {
	return &inputReproducer{entries: input}
}

func newWarmedInputReproducer(input []InputEntry) iInputReproducer {
	if len(input) > 1 {
		input = input[1:]
	}
	return &warmedinputReproducer{entries: input}
}

func (r *inputReproducer) next() (int, float64, string, float64, float64) {
	e := r.entries[r.index]
	r.index = (r.index + 1) % len(r.entries)
	r.setWarm()
	return e.Status, e.Duration, e.Body, e.TsBefore, e.TsAfter
}

func (r *inputReproducer) setWarm() {
	if !r.warmed {
		r.warmed = true
		if len(r.entries) > 1 {
			r.entries = r.entries[1:] // remove first entry
			r.index = 0
		}
	}
}

func (r *warmedinputReproducer) next() (int, float64, string, float64, float64) {
	e := r.entries[r.index]
	r.index = (r.index + 1) % len(r.entries)
	return e.Status, e.Duration, e.Body, e.TsBefore, e.TsAfter
}

// InputEntry packs information about one response.
type InputEntry struct {
	Status   int
	Duration float64
	Body     string
	TsBefore float64
	TsAfter  float64
}
