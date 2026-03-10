package ports

import (
	"context"

	domainports "whisper/libs/domain/ports"
)

type DependencyProbe interface {
	Check(ctx context.Context) error
}

type Metrics interface {
	IncRequests(path string)
	IncRejected()
	IncErrors()
	Render() string
}

type GatewayService interface {
	domainports.GatewayUseCase
}
