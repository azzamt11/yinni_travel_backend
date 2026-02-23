//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"yinni_travel_backend/app/booking/internal/biz"
	"yinni_travel_backend/app/booking/internal/data"
	"yinni_travel_backend/app/booking/internal/server"
	"yinni_travel_backend/app/booking/internal/service"
	"yinni_travel_backend/internal/conf"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Auth, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
