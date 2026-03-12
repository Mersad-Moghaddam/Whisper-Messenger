package ports

import "context"

type DependencyProbe interface {
	Check(ctx context.Context) error
}

type Metrics interface {
	IncRequests(path string)
	IncRejected()
	IncErrors()
	Render() string
}
