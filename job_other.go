//go:build !windows

package cove

type limit struct{}

func prepare(memory uint64) (*limit, error) {
	if memory > 0 {
		return nil, ErrLimited
	}
	return &limit{}, nil
}

func (l *limit) assign(int) error { return nil }

func (l *limit) cancel() error { return nil }

func (l *limit) close() {}
