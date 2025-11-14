package decorator

import "backend/internal/cli/builder/testdata/interfaces"

func DecorateService(s interfaces.Service) interfaces.Service { return s }
