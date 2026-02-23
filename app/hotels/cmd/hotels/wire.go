//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"trivgoo-backend/app/hotels/internal/biz"
	"trivgoo-backend/app/hotels/internal/data"
	"trivgoo-backend/app/hotels/internal/server"
	"trivgoo-backend/app/hotels/internal/service"
	"trivgoo-backend/internal/conf"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
