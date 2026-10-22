package outbound

import (
	"github.com/datosbolivia/open-knowledge-format/src/domain"
)

// ConfigLoaderPort abstracts reading configurations from files and environment variables.
type ConfigLoaderPort interface {
	Load(explicitFilePath string) (domain.Config, error)
}
