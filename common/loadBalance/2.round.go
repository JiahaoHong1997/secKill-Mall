package loadBalance

import "errors"

type RoundBalance struct {
	curIndex int
	rss      []string
}

func NewRoundBalance() *RoundBalance {
	return &RoundBalance{}
}

func (r *RoundBalance) Add(params ...string) error {
	if len(params) == 0 {
		return errors.New("param len 1 at least")
	}
	addr := params
	r.rss = append(r.rss, addr...)
	return nil
}

func (r *RoundBalance) Next() string {
	if len(r.rss) == 0 {
		return ""
	}

	lens := len(r.rss)
	if r.curIndex >= lens {
		r.curIndex = 0
	}
	curAddr := r.rss[r.curIndex]
	r.curIndex = (r.curIndex + 1) % lens
	return curAddr
}

func (r *RoundBalance) Get(key string) (string, error) {
	return r.Next(), nil
}
