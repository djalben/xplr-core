package shop

import (
	"log"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// ══════════════════════════════════════════════════════════════
// Deposit Monitor — periodically checks supplier balances and
// notifies admins when any balance drops below the threshold.
// ══════════════════════════════════════════════════════════════

const (
	depositCheckInterval  = 15 * time.Minute
	lowBalanceThresholdUS = 20.0 // $20 threshold
)

var lowBalanceThreshold = decimal.NewFromFloat(lowBalanceThresholdUS)

// AdminNotifier is a function that sends a notification to all admins.
// Injected from service.NotifyAdmins to avoid circular imports.
type AdminNotifier func(subject string, htmlMsg string)

// DepositMonitor checks supplier deposit balances on a schedule.
type DepositMonitor struct {
	registry      *Registry
	notifyAdmins  AdminNotifier
	stopCh        chan struct{}
	once          sync.Once
	alerted       map[string]time.Time // provider → last alert time (debounce)
	alertDebounce time.Duration
}

// NewDepositMonitor creates a new monitor.
func NewDepositMonitor(registry *Registry, notifyAdmins AdminNotifier) *DepositMonitor {
	return &DepositMonitor{
		registry:      registry,
		notifyAdmins:  notifyAdmins,
		stopCh:        make(chan struct{}),
		alerted:       make(map[string]time.Time),
		alertDebounce: 1 * time.Hour, // don't spam: max 1 alert per provider per hour
	}
}

// Start begins the background monitoring loop.
func (dm *DepositMonitor) Start() {
	dm.once.Do(func() {
		go dm.loop()
		log.Println("[DEPOSIT-MONITOR] ✅ Started (interval=15m, threshold=$20)")
	})
}

// Stop signals the monitor to stop.
func (dm *DepositMonitor) Stop() {
	select {
	case dm.stopCh <- struct{}{}:
	default:
	}
}

func (dm *DepositMonitor) loop() {
	// Run immediately on start
	dm.checkAll()

	ticker := time.NewTicker(depositCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			dm.checkAll()
		case <-dm.stopCh:
			log.Println("[DEPOSIT-MONITOR] Stopped")
			return
		}
	}
}

func (dm *DepositMonitor) checkAll() {
	providers := dm.registry.All()
	for _, p := range providers {
		if p.Name() == "demo" {
			continue // skip demo provider
		}
		dm.checkProvider(p)
	}
}

func (dm *DepositMonitor) checkProvider(p ProductProvider) {
	balance, err := p.GetBalance()
	if err != nil {
		log.Printf("[DEPOSIT-MONITOR] ⚠️ Failed to get balance for %s: %v", p.Name(), err)
		return
	}
	if balance == nil {
		return // provider doesn't support balance queries
	}

	log.Printf("[DEPOSIT-MONITOR] %s balance: $%s", p.Name(), balance.BalanceUSD.StringFixed(2))

	if balance.BalanceUSD.LessThan(lowBalanceThreshold) {
		dm.sendLowBalanceAlert(p.Name(), balance.BalanceUSD)
	}
}

func (dm *DepositMonitor) sendLowBalanceAlert(providerName string, balance decimal.Decimal) {
	// Debounce: don't send more than 1 alert per provider per hour
	if lastAlert, ok := dm.alerted[providerName]; ok {
		if time.Since(lastAlert) < dm.alertDebounce {
			return
		}
	}

	dm.alerted[providerName] = time.Now()

	subject := "⚠️ Низкий баланс у поставщика " + providerName
	htmlMsg := `<b>⚠️ Внимание: низкий баланс депозита!</b>` + "\n\n" +
		`Поставщик: <b>` + providerName + `</b>` + "\n" +
		`Текущий баланс: <b>$` + balance.StringFixed(2) + `</b>` + "\n" +
		`Минимальный порог: <b>$` + lowBalanceThreshold.StringFixed(2) + `</b>` + "\n\n" +
		`Необходимо пополнить депозит для бесперебойной работы магазина.`

	log.Printf("[DEPOSIT-MONITOR] 🚨 Low balance alert for %s: $%s < $%s",
		providerName, balance.StringFixed(2), lowBalanceThreshold.StringFixed(2))

	if dm.notifyAdmins != nil {
		dm.notifyAdmins(subject, htmlMsg)
	}
}
