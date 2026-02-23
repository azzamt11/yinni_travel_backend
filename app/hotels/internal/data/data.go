package data

import (
	"context"
	"yinni_travel_backend/ent"
	"yinni_travel_backend/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewHotelsRepo)

// Data .
type Data struct {
	ent *ent.Client
}

// NewData .
func NewData(c *conf.Data) (*Data, func(), error) {
	client, err := ent.Open(c.Database.Driver, c.Database.Source)
	if err != nil {
		return nil, nil, err
	}
	if err := client.Schema.Create(context.Background()); err != nil {
		_ = client.Close()
		return nil, nil, err
	}

	cleanup := func() {
		log.Info("closing the data resources")
		_ = client.Close()
	}
	return &Data{ent: client}, cleanup, nil
}
