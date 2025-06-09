package handlers

import (
	"gitlab.com/zeelrupapara/trade-engine/models"
	"go.uber.org/zap"
)

func (ec *EngineCore) InitWatcher() {
	// Subscribe to symbol events
	subSymbol := models.SubjectEngineSymbol
	_, err := ec.Nats.QueueSubscribe(subSymbol, models.GroupSymbol, ec.SymbolsEventHandler)
	if err != nil {
		ec.Logger.Fatal("NATS subscription failed", zap.String("subject", subSymbol), zap.Error(err))
	}
}
