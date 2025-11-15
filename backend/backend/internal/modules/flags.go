package modules

import (
	flagcfg "github.com/kthulu/kthulu-go/backend/internal/modules/flags"
	"go.uber.org/fx"
)

// FlagsModule provides configuration for request flags.
var FlagsModule = fx.Options(
	fx.Provide(flagcfg.LoadHeaderConfig),
)
