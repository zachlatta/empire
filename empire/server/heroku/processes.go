package heroku

import (
	"net/http"

	"github.com/bgentry/heroku-go"
	"github.com/remind101/empire/empire"
	"github.com/remind101/pkg/httpx"
	"golang.org/x/net/context"
)

type Dyno heroku.Dyno

func newDyno(j *empire.ProcessState) *Dyno {
	return &Dyno{
		Command:   string(j.Command),
		Name:      string(j.Name),
		State:     j.State,
		Size:      j.Constraints.String(),
		UpdatedAt: j.UpdatedAt,
	}
}

func newDynos(js []*empire.ProcessState) []*Dyno {
	dynos := make([]*Dyno, len(js))

	for i := 0; i < len(js); i++ {
		dynos[i] = newDyno(js[i])
	}

	return dynos
}

type GetProcesses struct {
	*empire.Empire
}

func (h *GetProcesses) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	a, err := findApp(ctx, h)
	if err != nil {
		return err
	}

	// Retrieve job states
	js, err := h.JobStatesByApp(ctx, a)
	if err != nil {
		return err
	}

	w.WriteHeader(200)
	return Encode(w, newDynos(js))
}

type PostProcessForm struct {
	Command string            `json:"command"`
	Attach  bool              `json:"attach"`
	Env     map[string]string `json:"env"`
	Size    string            `json:"size"`
}

type PostProcess struct {
	*empire.Empire
}

func (h *PostProcess) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var form PostProcessForm

	a, err := findApp(ctx, h)
	if err != nil {
		return err
	}

	if err := Decode(r, &form); err != nil {
		return err
	}

	relay, err := h.ProcessesRun(ctx, a, form.Command, empire.ProcessesRunOpts{
		Attach: form.Attach,
		Env:    form.Env,
		Size:   form.Size,
	})

	dyno := &heroku.Dyno{
		Name:      relay.Name,
		AttachURL: &relay.AttachURL,
		Command:   relay.Command,
		State:     relay.State,
		Type:      relay.Type,
		Size:      relay.Size,
		CreatedAt: relay.CreatedAt,
	}

	w.WriteHeader(201)
	return Encode(w, dyno)
}

type DeleteProcesses struct {
	*empire.Empire
}

func (h *DeleteProcesses) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	vars := httpx.Vars(ctx)
	ptype := empire.ProcessType(vars["ptype"])
	pid := vars["pid"]

	a, err := findApp(ctx, h)
	if err != nil {
		return err
	}

	err = h.ProcessesRestart(ctx, a, ptype, pid)
	if err != nil {
		return err
	}

	return NoContent(w)
}
