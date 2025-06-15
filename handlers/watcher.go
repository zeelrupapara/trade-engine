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

	// Fetch today's closed positions with symbol
	err := ec.DB.From("positions").
		Select("symbol", "profit").
		Where(goqu.And(
			goqu.C("closed_at").Gte(startOfDay),
			goqu.C("closed_at").Lt(now),
			goqu.C("profit").IsNotNull(),
		)).
		ScanStructs(&todayResults)
	if err != nil {
		ec.Logger.Error("❌ Failed to fetch today's data", zap.Error(err))
		return
	}

	// Fetch all closed positions
	err = ec.DB.From("positions").
		Select("symbol", "profit").
		Where(goqu.C("profit").IsNotNull()).
		ScanStructs(&allResults)
	if err != nil {
		ec.Logger.Error("❌ Failed to fetch overall data", zap.Error(err))
		return
	}

	// Symbol-wise today stats
	todaySymbolStats := make(map[string]struct {
		Count  int
		Wins   int
		Profit float64
	})
	todayPnL := 0.0

	for _, r := range todayResults {
		stat := todaySymbolStats[r.Symbol]
		stat.Count++
		if r.Profit > 0 {
			stat.Wins++
		}
		stat.Profit += r.Profit
		todaySymbolStats[r.Symbol] = stat
		todayPnL += r.Profit
	}

	// Symbol-wise all-time stats
	overallSymbolPnL := make(map[string]float64)
	overallPnL := 0.0
	for _, r := range allResults {
		overallSymbolPnL[r.Symbol] += r.Profit
		overallPnL += r.Profit
	}

	// Compute current equity
	equity := ec.Account.Balance + overallPnL

	// Build per-symbol report
	var symbolReport string
	if len(todaySymbolStats) == 0 {
		symbolReport = "No trades today."
	} else {
		for symbol, stats := range todaySymbolStats {
			winRate := 0.0
			if stats.Count > 0 {
				winRate = (float64(stats.Wins) / float64(stats.Count)) * 100
			}
			totalPnL := overallSymbolPnL[symbol]
			symbolReport += fmt.Sprintf(
				"📌 %s: Trades: %d | Wins: %d | WinRate: %.2f%% | TodayPnL: %.2f | TotalPnL: %.2f\n",
				symbol, stats.Count, stats.Wins, winRate, stats.Profit, totalPnL,
			)
		}
	}

	// Final report
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
		symbolReport,
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
