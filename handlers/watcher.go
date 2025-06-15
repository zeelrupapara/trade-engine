package handlers

import (
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/go-co-op/gocron"
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

	// Subcribe to the papar trading engine simulator for fake order to test the bot
	ec.SubscribeToOrders()

	ec.StartScheduling()
}

func (ec *EngineCore) StartScheduling() {
	ec.Cron = gocron.NewScheduler(time.Local)

	// Every 6 hours
	ec.Cron.Every(6).Hours().Do(func() {
		ec.sendDailyReport()
	})

	// Hourly health check at minute 0
	ec.Cron.Every(1).Hour().At("00:00").Do(func() {
		ec.sendHealthCheck()
	})

	ec.Cron.StartAsync()
}

// sendDailyReport fetches today's deals and sends a summary via Telegram
func (ec *EngineCore) sendDailyReport() {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var todayResults []struct {
		Profit float64 `db:"profit"`
	}
	var allResults []struct {
		Profit float64 `db:"profit"`
	}

	// Fetch today's closed positions
	err := ec.DB.From("positions").
		Select("profit").
		Where(goqu.And(
			goqu.C("closed_at").Gte(startOfDay),
			goqu.C("closed_at").Lt(now),
			goqu.C("profit").IsNotNull(),
		)).
		ScanStructs(&todayResults)
	if err != nil {
		ec.Logger.Error("❌ Failed to fetch today's report data", zap.Error(err))
		return
	}

	// Fetch all-time closed positions
	err = ec.DB.From("positions").
		Select("profit").
		Where(goqu.C("profit").IsNotNull()).
		ScanStructs(&allResults)
	if err != nil {
		ec.Logger.Error("❌ Failed to fetch overall report data", zap.Error(err))
		return
	}

	// Compute today's stats
	var todayPnL float64
	var todayWins int
	for _, r := range todayResults {
		todayPnL += r.Profit
		if r.Profit > 0 {
			todayWins++
		}
	}
	todayTotal := len(todayResults)
	todayWinRate := 0.0
	if todayTotal > 0 {
		todayWinRate = (float64(todayWins) / float64(todayTotal)) * 100
	}

	// Compute overall PnL
	var overallPnL float64
	for _, r := range allResults {
		overallPnL += r.Profit
	}

	// Compute equity (assuming only realized PnL added to balance)
	equity := ec.Account.Balance + overallPnL

	// Format message
	msg := fmt.Sprintf(
		"📊 *Daily Trading Report*\n\n"+
			"🗓️ *Date:* %s\n"+
			"👤 *Account ID:* `%s`\n"+
			"💼 *Balance:* %.2f\n"+
			"📊 *Equity:* %.2f\n\n"+
			"🔁 *Trades:* %d\n"+
			"✅ *Wins:* %d\n"+
			"📈 *Win Rate:* %.2f%%\n"+
			"💰 *Today PnL:* %.2f\n"+
			"💼 *Total PnL:* %.2f",
		now.Format("02 Jan 2006"),
		ec.Account.ID,
		ec.Account.Balance,
		equity,
		todayTotal,
		todayWins,
		todayWinRate,
		todayPnL,
		overallPnL,
	)

	// Send via Telegram
	if err := ec.Telegram.Send(msg); err != nil {
		ec.Logger.Error("❌ Failed to send Telegram report", zap.Error(err))
	}
}

// sendHealthCheck pings the DB and reports orders count in the last hour via Telegram pings the DB and reports orders count in the last hour via Telegram
func (ec *EngineCore) sendHealthCheck() {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)

	cnt, err := ec.DB.From("orders").
		Where(goqu.C("created_at").Gt(oneHourAgo)).
		Count()
	if err != nil {
		ec.Logger.Error("❌ Failed to count recent orders for health check", zap.Error(err))
		return
	}

	msg := fmt.Sprintf(
		"🛡 *Health Check Report*\n\n"+
			"🕒 Time: %s\n"+
			"📦 Orders in Last Hour: %d",
		now.Format("02 Jan 2006 15:04:05"),
		cnt,
	)

	if err := ec.Telegram.Send(msg); err != nil {
		ec.Logger.Error("❌ Failed to send health check to Telegram", zap.Error(err))
	}
}
