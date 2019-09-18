package sim

type IInputReproducer interface {
	next() (int, float64)
}

type InputReproducer struct {
	index   int
	warmed  bool
	entries []InputEntry
}

type WarmedInputReproducer struct {
	index   int
	entries []InputEntry
}

func newInputReproducer(input []InputEntry) IInputReproducer {
	return &InputReproducer{entries: input}
}

func newWarmedInputReproducer(input []InputEntry) IInputReproducer {
	if len(input) > 1 {
		input = input[1:]
	}
	return &WarmedInputReproducer{entries: input}
}

func (r *InputReproducer) next() (int, float64) {
	e := r.entries[r.index]
	r.index = (r.index + 1) % len(r.entries)
	r.setWarm()
	return e.Status, e.Duration
}

func (r *InputReproducer) setWarm() {
	if !r.warmed {
		r.warmed = true
		if len(r.entries) > 1 {
			r.entries = r.entries[1:] // remove first entry
			r.index = 0
		}
	}
}

func (r *WarmedInputReproducer) next() (int, float64) {
	e := r.entries[r.index]
	r.index = (r.index + 1) % len(r.entries)
	return e.Status, e.Duration
}

// InputEntry packs information about one response.
type InputEntry struct {
	Status   int
	Duration float64
}
