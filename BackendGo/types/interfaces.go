package types

import (
	"CryptoLens_Backend/models"
	"context"
)

type BybitInstrumentRepositoryInterface interface {
	GetBySymbol(ctx context.Context, symbol string) (*models.BybitInstrument, error)
}
