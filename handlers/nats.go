package handlers

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"gitlab.com/zeelrupapara/trade-engine/models"
	"go.uber.org/zap"
)

func (ec *EngineCore) SymbolsEventHandler(msg *nats.Msg) {
	ec.Logger.Sugar().Infof("Received Dealing Room event on %s subject \n", msg.Subject)

	// unmarshal data
	symbolEvent := &models.SymbolEvent{}
	err := json.Unmarshal(msg.Data, symbolEvent)
	if err != nil {
		ec.Logger.Error(err.Error(), zap.Any("Symbol", "Event Unmarshal"))
	}

	switch symbolEvent.Type {
	case models.SymbolEventType_AddSymbol:
		ec.Logger.Debug("Adding Symbol", zap.String("symbol", symbolEvent.Symbol), zap.String("exchange", symbolEvent.Exchange))
		ec.SubcribeSymbol(symbolEvent.Exchange, symbolEvent.Symbol)
	case models.SymbolEventType_RemoveSymbol:
		ec.Logger.Debug("Removing Symbol", zap.String("symbol", symbolEvent.Symbol), zap.String("exchange", symbolEvent.Exchange))
		ec.UnsubcribeSymbol(symbolEvent.Exchange, symbolEvent.Symbol)
	}
}
