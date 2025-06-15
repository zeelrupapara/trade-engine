package handlers

import (
	"fmt"
	"strings"
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
	// Watch Closing Positions
	ec.StartPriceWatcher()

	ec.StartScheduling()
}

func (ec *EngineCore) StartScheduling() {
	ec.Cron = gocron.NewScheduler(time.Local)

	// Every 30 minutes
	ec.Cron.Every(30).Minutes().Do(func() {
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

	type PnLEntry struct {
		Symbol string  `db:"symbol"`
		Profit float64 `db:"profit"`
	}

	var todayResults []PnLEntry
	var allResults []PnLEntry

	// Today's symbol-wise closed positions
	err := ec.DB.From("positions").
		Select("symbol", "profit").
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

	// All-time symbol-wise closed positions
	err = ec.DB.From("positions").
		Select("symbol", "profit").
		Where(goqu.C("profit").IsNotNull()).
		ScanStructs(&allResults)
	if err != nil {
		ec.Logger.Error("❌ Failed to fetch overall report data", zap.Error(err))
		return
	}

	// Aggregate today's symbol PnL
	todaySymbolStats := make(map[string]struct {
		Count  int
		Wins   int
		Profit float64
	})
	var todayPnL float64
	for _, r := range todayResults {
		s := todaySymbolStats[r.Symbol]
		s.Count++
		if r.Profit > 0 {
			s.Wins++
		}
		s.Profit += r.Profit
		todaySymbolStats[r.Symbol] = s
		todayPnL += r.Profit
	}

	// Aggregate all-time symbol PnL
	overallSymbolPnL := make(map[string]float64)
	var overallPnL float64
	for _, r := range allResults {
		overallSymbolPnL[r.Symbol] += r.Profit
		overallPnL += r.Profit
	}

	// Compute equity
	equity := ec.Account.Balance + overallPnL

	// Format symbol-wise stats
	var symbolReport strings.Builder
	for symbol, stats := range todaySymbolStats {
		winRate := 0.0
		if stats.Count > 0 {
			winRate = (float64(stats.Wins) / float64(stats.Count)) * 100
		}
		symbolReport.WriteString(fmt.Sprintf(
			"📌 *%s*: Trades: %d | Wins: %d | WinRate: %.2f%% | PnL: %.2f | TotalPnL: %.2f\n",
			symbol, stats.Count, stats.Wins, winRate, stats.Profit, overallSymbolPnL[symbol],
		))
	}

	// Final message
	msg := fmt.Sprintf(
		"📊 *Daily Trading Report*\n\n"+
			"🗓️ *Date:* %s\n"+
			"👤 *Account ID:* `%s`\n"+
			"💼 *Balance:* %.2f\n"+
			"📈 *Equity:* %.2f\n"+
			"💰 *Today PnL:* %.2f\n"+
			"💰 *Total PnL:* %.2f\n\n"+
			"🔍 *Per-Symbol Summary:*\n%s",
		now.Format("02 Jan 2006"),
		ec.Account.ID,
		ec.Account.Balance,
		equity,
		todayPnL,
		overallPnL,
		symbolReport.String(),
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
