package outbound

import (
	"context"

	"github.com/datosbolivia/open-knowledge-format/src/domain"
)

// AuthProviderPort defines outbound operations for injecting authentication headers into outgoing requests.
type AuthProviderPort interface {
	GetAuthHeaders(ctx context.Context) (map[string]string, error)
	GetAuthType() string
}

// WellKnownResolverPort defines outbound operations for discovering catalog configurations.
type WellKnownResolverPort interface {
	FetchWellKnown(ctx context.Context, targetURL string) (*domain.WellKnownDiscovery, error)
}
