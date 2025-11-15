// @kthulu:module:verifactu
package modules

import (
	"os"

	"go.uber.org/fx"

	"github.com/pmaojo/kthulu-go/backend/core"
	adapterhttp "github.com/pmaojo/kthulu-go/backend/internal/adapters/http"
	db "github.com/pmaojo/kthulu-go/backend/internal/infrastructure/db"
	vf "github.com/pmaojo/kthulu-go/backend/internal/modules/verifactu"
)

// VerifactuModule wires VeriFactu dependencies and routes.
var VerifactuModule = fx.Options(
	// Services
	fx.Provide(
		db.NewVerifactuRepository,
		func() vf.Signer {
			key := []byte(os.Getenv("VERIFACTU_SIGN_KEY"))
			if len(key) == 0 {
				key = []byte("changeme")
			}
			return vf.NewHMACSigner(key)
		},
		func(repo vf.Repository, signer vf.Signer, cfg *core.Config) *vf.Service {
			return vf.NewService(repo, signer, cfg.VerifactuSIFCode, cfg.VerifactuMode)
		},
	),

	// HTTP handlers
	fx.Provide(
		adapterhttp.NewVerifactuHandler,
	),

	// Register routes
	fx.Invoke(func(handler *adapterhttp.VerifactuHandler, registry *RouteRegistry) {
		registry.Register(handler)
	}),
)
