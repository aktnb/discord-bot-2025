package version

import "context"

type Service struct {
	version string
}

func NewVersionService(version string) *Service {
	return &Service{version: version}
}

func (s *Service) Version(ctx context.Context) (string, error) {
	return s.version, nil
}
