package service

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// ═══════════════════════════════════════════════════
// SMTP transport — supports both port 465 (implicit TLS)
// and port 587 (STARTTLS).
// Env vars: SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, SMTP_FROM, APP_DOMAIN
// ═══════════════════════════════════════════════════

type smtpConfig struct {
	Host   string
	Port   string
	User   string
	Pass   string
	From   string
	Domain string
}

func loadSMTPConfig() *smtpConfig {
	c := &smtpConfig{
		Host:   os.Getenv("SMTP_HOST"),
		Port:   os.Getenv("SMTP_PORT"),
		User:   os.Getenv("SMTP_USER"),
		Pass:   os.Getenv("SMTP_PASS"),
		From:   os.Getenv("SMTP_FROM"),
		Domain: os.Getenv("APP_DOMAIN"),
	}
	if c.From == "" {
		if c.User != "" {
			c.From = c.User
		} else {
			c.From = "admin@xplr.pro"
		}
	}
	if c.Domain == "" {
		c.Domain = "https://xplr.pro"
	}
	return c
}

// sendMail — unified sender via admin@xplr.pro. Uses implicit TLS for port 465, STARTTLS for 587.
func sendMail(to, subject, htmlBody string) error {
	return sendMailWith(loadSMTPConfig(), to, subject, htmlBody)
}

func sendMailWith(cfg *smtpConfig, to, subject, htmlBody string) error {
	if cfg.Host == "" || cfg.Port == "" {
		log.Printf("[SMTP-ERROR] SMTP not configured (SMTP_HOST=%q, SMTP_PORT=%q). Email to %s skipped.", cfg.Host, cfg.Port, to)
		return fmt.Errorf("SMTP not configured: SMTP_HOST and SMTP_PORT are required")
	}
	if cfg.User == "" || cfg.Pass == "" {
		log.Printf("[SMTP-ERROR] SMTP credentials missing (SMTP_USER=%q, SMTP_PASS length=%d). Email to %s skipped.", cfg.User, len(cfg.Pass), to)
		return fmt.Errorf("SMTP credentials missing: SMTP_USER and SMTP_PASS are required")
	}

	log.Printf("[EMAIL] 📤 Sending email: to=%s, from=%s, host=%s:%s, user=%s, subject=%q",
		to, cfg.From, cfg.Host, cfg.Port, cfg.User, subject)

	headers := fmt.Sprintf(
		"From: XPLR <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n",
		cfg.From, to, subject,
	)
	msg := []byte(headers + htmlBody)
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	var err error
	if cfg.Port == "465" {
		err = sendImplicitTLS(cfg, addr, to, msg)
	} else {
		// STARTTLS (port 587 etc.) with timeout to prevent hangs
		err = sendSTARTTLS(cfg, addr, to, msg)
	}

	if err != nil {
		log.Printf("[EMAIL] ❌ FAILED to send to %s: %v", to, err)
	} else {
		log.Printf("[EMAIL] ✅ Sent successfully to %s", to)
	}
	return err
}

// sendImplicitTLS — connects with TLS first, then authenticates (Zoho, port 465).
func sendImplicitTLS(cfg *smtpConfig, addr, to string, msg []byte) error {
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	tlsConfig := &tls.Config{ServerName: cfg.Host}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls dial (timeout 15s): %w", err)
	}
	defer conn.Close()

	host, _, _ := net.SplitHostPort(addr)
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer client.Quit()

	auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err = client.Mail(cfg.From); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	return w.Close()
}

// sendSTARTTLS — connects with plain TCP + STARTTLS upgrade (port 587).
// Uses 15s dial timeout to prevent indefinite hangs.
func sendSTARTTLS(cfg *smtpConfig, addr, to string, msg []byte) error {
	conn, err := net.DialTimeout("tcp", addr, 15*time.Second)
	if err != nil {
		return fmt.Errorf("tcp dial (timeout 15s): %w", err)
	}
	defer conn.Close()

	// Set overall deadline for the entire SMTP conversation
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	host, _, _ := net.SplitHostPort(addr)
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer client.Quit()

	// Upgrade to TLS
	tlsConfig := &tls.Config{ServerName: cfg.Host}
	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("starttls: %w", err)
	}

	auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err = client.Mail(cfg.From); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	return w.Close()
}

// ═══════════════════════════════════════════════════
// Shared HTML template wrapper
// ═══════════════════════════════════════════════════

func wrapHTML(title, content string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="ru">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head>
<body style="margin:0;padding:0;background:#121212;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
<table width="100%%" cellpadding="0" cellspacing="0" style="background:#121212;padding:40px 0;">
<tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="background:#1A1A1A;border-radius:16px;border:1px solid rgba(255,255,255,0.08);overflow:hidden;">
  <!-- Header -->
  <tr><td style="padding:32px 40px 20px;text-align:center;border-bottom:1px solid rgba(255,255,255,0.08);">
    <div style="display:inline-block;background:linear-gradient(135deg,#3b82f6,#8b5cf6);border-radius:12px;width:48px;height:48px;line-height:48px;text-align:center;color:#fff;font-size:22px;font-weight:bold;">X</div>
    <h1 style="margin:12px 0 0;color:#FFFFFF;font-size:20px;font-weight:700;letter-spacing:-0.3px;">%s</h1>
  </td></tr>
  <!-- Body -->
  <tr><td style="padding:32px 40px;">%s</td></tr>
  <!-- Footer -->
  <tr><td style="padding:20px 40px 32px;text-align:center;border-top:1px solid rgba(255,255,255,0.08);">
    <p style="margin:0;color:#9ca3af;font-size:11px;">© XPLR — Premium Financial Services</p>
    <p style="margin:4px 0 0;color:#6b7280;font-size:10px;">Это автоматическое письмо, не отвечайте на него.</p>
  </td></tr>
</table>
</td></tr></table>
</body></html>`, title, content)
}

// ═══════════════════════════════════════════════════
// Email functions
// ═══════════════════════════════════════════════════

// SendVerificationEmail — ссылка подтверждения email.
func SendVerificationEmail(toEmail, token string) error {
	cfg := loadSMTPConfig()
	verifyURL := fmt.Sprintf("%s/verify?token=%s", cfg.Domain, token)

	content := fmt.Sprintf(`
    <p style="color:#cbd5e1;font-size:15px;line-height:1.6;margin:0 0 20px;">Здравствуйте!</p>
    <p style="color:#94a3b8;font-size:14px;line-height:1.6;margin:0 0 24px;">Для подтверждения вашего email нажмите на кнопку ниже:</p>
    <div style="text-align:center;margin:0 0 24px;">
      <a href="%s" style="display:inline-block;padding:14px 40px;background:linear-gradient(135deg,#3b82f6,#8b5cf6);color:#fff;text-decoration:none;border-radius:12px;font-size:14px;font-weight:600;">Подтвердить email</a>
    </div>
    <p style="color:#64748b;font-size:12px;line-height:1.5;margin:0;">Ссылка действительна 24 часа. Если вы не регистрировались на XPLR — проигнорируйте это письмо.</p>`, verifyURL)

	html := wrapHTML("Подтверждение email", content)

	if err := sendMail(toEmail, "XPLR — Подтверждение email", html); err != nil {
		log.Printf("[EMAIL] Failed to send verification email to %s: %v", toEmail, err)
		return err
	}
	log.Printf("[EMAIL] Verification email sent to %s", toEmail)
	return nil
}

// SendPasswordResetEmail — ссылка сброса пароля.
func SendPasswordResetEmail(toEmail, token string) error {
	cfg := loadSMTPConfig()
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", cfg.Domain, token)

	content := fmt.Sprintf(`
    <p style="color:#cbd5e1;font-size:15px;line-height:1.6;margin:0 0 20px;">Здравствуйте!</p>
    <p style="color:#94a3b8;font-size:14px;line-height:1.6;margin:0 0 24px;">Вы запросили сброс пароля для вашего аккаунта XPLR.</p>
    <div style="text-align:center;margin:0 0 24px;">
      <a href="%s" style="display:inline-block;padding:14px 40px;background:linear-gradient(135deg,#3b82f6,#8b5cf6);color:#fff;text-decoration:none;border-radius:12px;font-size:14px;font-weight:600;">Сбросить пароль</a>
    </div>
    <p style="color:#64748b;font-size:12px;line-height:1.5;margin:0;">Ссылка действительна 1 час. Если вы не запрашивали сброс — проигнорируйте это письмо.</p>`, resetURL)

	html := wrapHTML("Сброс пароля", content)

	if err := sendMail(toEmail, "XPLR — Сброс пароля", html); err != nil {
		log.Printf("[EMAIL] Failed to send password reset email to %s: %v", toEmail, err)
		return err
	}
	log.Printf("[EMAIL] Password reset email sent to %s", toEmail)
	return nil
}

// SendWelcomeEmail — приветственное письмо после регистрации.
func SendWelcomeEmail(toEmail string) error {
	cfg := loadSMTPConfig()

	content := fmt.Sprintf(`
    <p style="color:#cbd5e1;font-size:16px;line-height:1.6;margin:0 0 8px;">Добро пожаловать в <strong style="color:#fff;">XPLR</strong>! 🎉</p>
    <p style="color:#94a3b8;font-size:14px;line-height:1.7;margin:0 0 24px;">
      Ваш аккаунт успешно создан. Теперь вам доступны:<br/>
      — Выпуск виртуальных карт для онлайн-покупок<br/>
      — Управление балансом и пополнение кошелька<br/>
      — Реферальная программа с бонусами<br/>
      — Персональные условия с ростом грейда
    </p>
    <div style="text-align:center;margin:0 0 24px;">
      <a href="%s/dashboard" style="display:inline-block;padding:14px 40px;background:linear-gradient(135deg,#3b82f6,#8b5cf6);color:#fff;text-decoration:none;border-radius:12px;font-size:14px;font-weight:600;">Перейти в личный кабинет</a>
    </div>
    <p style="color:#64748b;font-size:12px;line-height:1.5;margin:0;">Если у вас есть вопросы — напишите в поддержку через личный кабинет.</p>`, cfg.Domain)

	html := wrapHTML("Добро пожаловать!", content)

	if err := sendMail(toEmail, "Добро пожаловать в XPLR! 🎉", html); err != nil {
		log.Printf("[EMAIL] Failed to send welcome email to %s: %v", toEmail, err)
		return err
	}
	log.Printf("[EMAIL] Welcome email sent to %s", toEmail)
	return nil
}

// SendGradeChangeEmail — уведомление о смене грейда.
func SendGradeChangeEmail(toEmail, newGrade string) error {
	gradeColors := map[string]string{
		"STANDARD": "#94a3b8",
		"SILVER":   "#c0c0c0",
		"GOLD":     "#fbbf24",
		"PLATINUM": "#a78bfa",
		"BLACK":    "#fff",
	}
	gradeEmoji := map[string]string{
		"STANDARD": "⚪",
		"SILVER":   "🥈",
		"GOLD":     "🥇",
		"PLATINUM": "💎",
		"BLACK":    "🖤",
	}
	gradeBg := map[string]string{
		"STANDARD": "linear-gradient(135deg,#475569,#64748b)",
		"SILVER":   "linear-gradient(135deg,#9ca3af,#d1d5db)",
		"GOLD":     "linear-gradient(135deg,#f59e0b,#fbbf24)",
		"PLATINUM": "linear-gradient(135deg,#8b5cf6,#a78bfa)",
		"BLACK":    "linear-gradient(135deg,#1e1e2e,#000)",
	}
	gradeBenefit := map[string]string{
		"STANDARD": "Базовая комиссия 6.70%",
		"SILVER":   "Сниженная комиссия 5.50%",
		"GOLD":     "Выгодная комиссия 4.50%",
		"PLATINUM": "Премиальная комиссия 3.50%",
		"BLACK":    "Минимальная комиссия 2.50% • Приоритетная поддержка • Эксклюзивные лимиты",
	}

	g := strings.ToUpper(newGrade)
	color := gradeColors[g]
	if color == "" {
		color = "#94a3b8"
	}
	emoji := gradeEmoji[g]
	bg := gradeBg[g]
	if bg == "" {
		bg = "linear-gradient(135deg,#475569,#64748b)"
	}
	benefit := gradeBenefit[g]

	cfg := loadSMTPConfig()

	content := fmt.Sprintf(`
    <p style="color:#cbd5e1;font-size:15px;line-height:1.6;margin:0 0 20px;">Здравствуйте!</p>
    <p style="color:#94a3b8;font-size:14px;line-height:1.6;margin:0 0 24px;">Ваш статус в XPLR был обновлён:</p>
    <!-- Grade badge -->
    <div style="text-align:center;margin:0 0 24px;">
      <div style="display:inline-block;padding:20px 48px;background:%s;border-radius:16px;border:1px solid rgba(255,255,255,0.1);">
        <span style="font-size:32px;">%s</span>
        <p style="margin:8px 0 0;color:%s;font-size:22px;font-weight:800;letter-spacing:2px;">%s</p>
      </div>
    </div>
    <div style="background:rgba(255,255,255,0.03);border:1px solid rgba(255,255,255,0.06);border-radius:12px;padding:16px 20px;margin:0 0 24px;">
      <p style="color:#94a3b8;font-size:13px;margin:0;">Ваши привилегии: <strong style="color:#e2e8f0;">%s</strong></p>
    </div>
    <div style="text-align:center;margin:0 0 24px;">
      <a href="%s/dashboard" style="display:inline-block;padding:14px 40px;background:linear-gradient(135deg,#3b82f6,#8b5cf6);color:#fff;text-decoration:none;border-radius:12px;font-size:14px;font-weight:600;">Открыть личный кабинет</a>
    </div>`,
		bg, emoji, color, g, benefit, cfg.Domain)

	html := wrapHTML("Ваш грейд обновлён", content)
	subject := fmt.Sprintf("XPLR — Ваш грейд: %s %s", g, emoji)

	if err := sendMail(toEmail, subject, html); err != nil {
		log.Printf("[EMAIL] Failed to send grade change email to %s: %v", toEmail, err)
		return err
	}
	log.Printf("[EMAIL] Grade change email sent to %s (grade=%s)", toEmail, g)
	return nil
}

// SendSupportTicketNotification — дублирует тикет на admin@xplr.pro через Zoho SMTP.
func SendSupportTicketNotification(ticketID int, userEmail, subject, message string) error {
	content := fmt.Sprintf(`
    <p style="color:#cbd5e1;font-size:15px;line-height:1.6;margin:0 0 20px;">Новый тикет поддержки</p>
    <div style="background:rgba(255,255,255,0.03);border:1px solid rgba(255,255,255,0.06);border-radius:12px;padding:16px 20px;margin:0 0 20px;">
      <p style="color:#94a3b8;font-size:13px;margin:0 0 8px;"><strong style="color:#e2e8f0;">Тикет:</strong> #%d</p>
      <p style="color:#94a3b8;font-size:13px;margin:0 0 8px;"><strong style="color:#e2e8f0;">Email:</strong> %s</p>
      <p style="color:#94a3b8;font-size:13px;margin:0 0 8px;"><strong style="color:#e2e8f0;">Тема:</strong> %s</p>
    </div>
    <div style="background:rgba(59,130,246,0.06);border:1px solid rgba(59,130,246,0.15);border-radius:12px;padding:16px 20px;margin:0 0 20px;">
      <p style="color:#94a3b8;font-size:13px;margin:0 0 4px;"><strong style="color:#e2e8f0;">Сообщение:</strong></p>
      <p style="color:#cbd5e1;font-size:14px;line-height:1.6;margin:0;white-space:pre-wrap;">%s</p>
    </div>
    <p style="color:#64748b;font-size:12px;">Ответьте пользователю через админ-панель или напрямую на email.</p>`,
		ticketID, userEmail, subject, message)

	html := wrapHTML("Новый тикет поддержки", content)
	emailSubject := fmt.Sprintf("XPLR Тикет #%d — %s", ticketID, subject)

	if err := sendMail("admin@xplr.pro", emailSubject, html); err != nil {
		log.Printf("[EMAIL] Failed to send support ticket notification: %v", err)
		return err
	}
	log.Printf("[EMAIL] Support ticket #%d notification sent to admin@xplr.pro", ticketID)
	return nil
}

// SendEmailVerifyCode — отправка 6-значного кода подтверждения email.
func SendEmailVerifyCode(toEmail, code string) error {
	content := fmt.Sprintf(`
    <p style="color:#cbd5e1;font-size:15px;line-height:1.6;margin:0 0 20px;">Здравствуйте!</p>
    <p style="color:#94a3b8;font-size:14px;line-height:1.6;margin:0 0 24px;">Ваш код подтверждения email:</p>
    <div style="text-align:center;margin:0 0 24px;">
      <div style="display:inline-block;padding:20px 48px;background:linear-gradient(135deg,#3b82f6,#8b5cf6);border-radius:16px;border:1px solid rgba(255,255,255,0.1);">
        <p style="margin:0;color:#fff;font-size:36px;font-weight:800;letter-spacing:8px;">%s</p>
      </div>
    </div>
    <p style="color:#64748b;font-size:12px;line-height:1.5;margin:0;">Код действителен 15 минут. Если вы не запрашивали подтверждение — проигнорируйте это письмо.</p>`, code)

	html := wrapHTML("Подтверждение email", content)

	if err := sendMail(toEmail, "XPLR — Код подтверждения email", html); err != nil {
		log.Printf("[EMAIL] Failed to send verify code to %s: %v", toEmail, err)
		return err
	}
	log.Printf("[EMAIL] Verify code sent to %s", toEmail)
	return nil
}

// SendEmergencyFreezeNotification — уведомление о блокировке аккаунта.
func SendEmergencyFreezeNotification(toEmail string, frozenCards int) error {
	content := fmt.Sprintf(`
    <div style="text-align:center;margin:0 0 24px;">
      <div style="display:inline-block;width:64px;height:64px;background:linear-gradient(135deg,#ef4444,#dc2626);border-radius:50%%;line-height:64px;font-size:28px;">🔒</div>
    </div>
    <p style="color:#fca5a5;font-size:16px;line-height:1.6;margin:0 0 12px;text-align:center;font-weight:700;">Ваш аккаунт заблокирован</p>
    <p style="color:#94a3b8;font-size:14px;line-height:1.7;margin:0 0 24px;text-align:center;">
      В целях безопасности все операции по вашему аккаунту приостановлены.
    </p>
    <div style="background:rgba(239,68,68,0.08);border:1px solid rgba(239,68,68,0.2);border-radius:12px;padding:16px 20px;margin:0 0 24px;">
      <p style="color:#fca5a5;font-size:13px;margin:0 0 4px;">Что произошло:</p>
      <p style="color:#94a3b8;font-size:13px;margin:0;">— Заморожено карт: <strong style="color:#fff;">%d</strong></p>
      <p style="color:#94a3b8;font-size:13px;margin:0;">— Статус аккаунта: <strong style="color:#ef4444;">BANNED</strong></p>
      <p style="color:#94a3b8;font-size:13px;margin:0;">— Баланс кошелька: <strong style="color:#ef4444;">обнулён</strong></p>
    </div>
    <p style="color:#64748b;font-size:12px;line-height:1.5;margin:0;text-align:center;">Если вы считаете, что это ошибка — свяжитесь с нами: admin@xplr.pro</p>`, frozenCards)

	html := wrapHTML("Аккаунт заблокирован", content)

	if err := sendMail(toEmail, "XPLR — Ваш аккаунт заблокирован 🔒", html); err != nil {
		log.Printf("[EMAIL] Failed to send freeze notification to %s: %v", toEmail, err)
		return err
	}
	log.Printf("[EMAIL] Emergency freeze notification sent to %s", toEmail)
	return nil
}

// SendPurchaseReceipt — premium purchase receipt email with order details.
// For eSIM purchases, includes activation instructions block.
func SendPurchaseReceipt(toEmail string, orderID int, productName string, priceUSD string, cardLast4 string, isESIM bool, activationData map[string]string) error {
	cfg := loadSMTPConfig()
	dateStr := fmt.Sprintf("%s", time.Now().Format("02.01.2006 15:04"))

	// Order details block
	orderBlock := fmt.Sprintf(`
    <div style="background:rgba(255,255,255,0.04);border:1px solid rgba(255,255,255,0.08);border-radius:12px;padding:20px 24px;margin:0 0 24px;">
      <table width="100%%" cellpadding="0" cellspacing="0">
        <tr><td style="padding:6px 0;color:#9ca3af;font-size:13px;">Номер заказа</td><td style="padding:6px 0;color:#FFFFFF;font-size:13px;text-align:right;font-weight:600;">#%d</td></tr>
        <tr><td style="padding:6px 0;color:#9ca3af;font-size:13px;">Дата</td><td style="padding:6px 0;color:#FFFFFF;font-size:13px;text-align:right;">%s</td></tr>
        <tr><td style="padding:6px 0;color:#9ca3af;font-size:13px;">Товар</td><td style="padding:6px 0;color:#FFFFFF;font-size:13px;text-align:right;font-weight:600;">%s</td></tr>
        <tr><td style="padding:6px 0;color:#9ca3af;font-size:13px;">Сумма</td><td style="padding:6px 0;color:#60a5fa;font-size:15px;text-align:right;font-weight:700;">$%s</td></tr>
        <tr><td style="padding:6px 0;color:#9ca3af;font-size:13px;">Оплата</td><td style="padding:6px 0;color:#FFFFFF;font-size:13px;text-align:right;">Карта •••• %s</td></tr>
        <tr><td colspan="2" style="padding:12px 0 0;"><div style="height:1px;background:rgba(255,255,255,0.06);"></div></td></tr>
        <tr><td style="padding:8px 0 0;color:#9ca3af;font-size:13px;">Статус</td><td style="padding:8px 0 0;text-align:right;"><span style="display:inline-block;padding:4px 12px;background:rgba(34,197,94,0.15);color:#4ade80;border-radius:8px;font-size:12px;font-weight:600;">Оплачено</span></td></tr>
      </table>
    </div>`, orderID, dateStr, productName, priceUSD, cardLast4)

	// eSIM activation instructions (optional)
	esimBlock := ""
	if isESIM && activationData != nil {
		qrData := activationData["qr_data"]
		smdp := activationData["smdp"]
		matchingID := activationData["matching_id"]
		iccid := activationData["iccid"]

		esimBlock = fmt.Sprintf(`
    <div style="background:rgba(59,130,246,0.06);border:1px solid rgba(59,130,246,0.15);border-radius:12px;padding:20px 24px;margin:0 0 24px;">
      <p style="margin:0 0 12px;color:#60a5fa;font-size:14px;font-weight:700;">📱 Инструкция по активации eSIM</p>
      <p style="margin:0 0 8px;color:#e2e8f0;font-size:13px;line-height:1.6;"><strong>Способ 1:</strong> Настройки → Сотовая связь → Добавить eSIM → Сканировать QR-код</p>
      <p style="margin:0 0 16px;color:#e2e8f0;font-size:13px;line-height:1.6;"><strong>Способ 2:</strong> Ввести данные вручную:</p>`)

		if smdp != "" {
			esimBlock += fmt.Sprintf(`
      <div style="background:rgba(0,0,0,0.2);border-radius:8px;padding:10px 14px;margin:0 0 8px;">
        <p style="margin:0 0 2px;color:#9ca3af;font-size:10px;text-transform:uppercase;letter-spacing:1px;">SM-DP+ адрес</p>
        <p style="margin:0;color:#FFFFFF;font-size:12px;font-family:monospace;word-break:break-all;">%s</p>
      </div>`, smdp)
		}
		if matchingID != "" {
			esimBlock += fmt.Sprintf(`
      <div style="background:rgba(0,0,0,0.2);border-radius:8px;padding:10px 14px;margin:0 0 8px;">
        <p style="margin:0 0 2px;color:#9ca3af;font-size:10px;text-transform:uppercase;letter-spacing:1px;">Код активации</p>
        <p style="margin:0;color:#FFFFFF;font-size:12px;font-family:monospace;word-break:break-all;">%s</p>
      </div>`, matchingID)
		}
		if qrData != "" {
			esimBlock += fmt.Sprintf(`
      <div style="text-align:center;margin:16px 0 8px;">
        <img src="https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=%s" alt="QR" style="width:180px;height:180px;border-radius:12px;" />
      </div>`, qrData)
		}
		if iccid != "" {
			esimBlock += fmt.Sprintf(`
      <p style="margin:8px 0 0;color:#9ca3af;font-size:11px;">ICCID: %s</p>`, iccid)
		}
		esimBlock += `
    </div>`
	}

	content := fmt.Sprintf(`
    <p style="color:#FFFFFF;font-size:16px;line-height:1.6;margin:0 0 8px;">Спасибо за покупку! 🎉</p>
    <p style="color:#d1d5db;font-size:14px;line-height:1.6;margin:0 0 24px;">Ваш заказ успешно оплачен. Ниже — детали:</p>
    %s
    %s
    <div style="text-align:center;margin:0 0 24px;">
      <a href="%s/purchases" style="display:inline-block;padding:14px 40px;background:linear-gradient(135deg,#3b82f6,#8b5cf6);color:#fff;text-decoration:none;border-radius:12px;font-size:14px;font-weight:600;">Мои покупки</a>
    </div>
    <p style="color:#6b7280;font-size:12px;line-height:1.5;margin:0;">Если у вас есть вопросы — обратитесь в поддержку через личный кабинет.</p>`,
		orderBlock, esimBlock, cfg.Domain)

	html := wrapHTML("Чек покупки", content)
	subject := fmt.Sprintf("XPLR — Чек #%d: %s", orderID, productName)

	if err := sendMail(toEmail, subject, html); err != nil {
		log.Printf("[EMAIL] Failed to send purchase receipt to %s: %v", toEmail, err)
		return err
	}
	log.Printf("[EMAIL] Purchase receipt #%d sent to %s", orderID, toEmail)
	return nil
}

// SendGenericEmail — sends a generic HTML email (used by NotifyUser).
func SendGenericEmail(toEmail, subject, htmlContent string) error {
	if toEmail == "" {
		log.Printf("[SMTP-ERROR] Cannot send email — recipient address is empty (subject=%q)", subject)
		return fmt.Errorf("recipient email is empty")
	}
	html := wrapHTML(subject, htmlContent)
	if err := sendMail(toEmail, "XPLR — "+subject, html); err != nil {
		log.Printf("[SMTP-ERROR] Failed to send email to %s: %v (subject=%q)", toEmail, err, subject)
		return err
	}
	log.Printf("[SMTP-OK] Email delivered to %s (subject=%q)", toEmail, subject)
	return nil
}
