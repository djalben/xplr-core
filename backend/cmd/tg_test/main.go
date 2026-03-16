package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

// Standalone Telegram Bot test — verifies token, prints bot info, and optionally sends a test message.
// Usage: go run backend/cmd/tg_test/main.go [optional_chat_id]
func main() {
	if err := godotenv.Load("backend/.env"); err != nil {
		log.Printf("Warning: %v", err)
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")

	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("  XPLR Telegram Bot Diagnostic")
	fmt.Println("═══════════════════════════════════════════")

	if token == "" {
		fmt.Println("\n  ❌ TELEGRAM_BOT_TOKEN is NOT SET in .env")
		fmt.Println("  → Add it: TELEGRAM_BOT_TOKEN=123456789:ABCDefGhIjKlMnOpQrStUvWxYz")
		fmt.Println("  → Get a token from @BotFather in Telegram")
		return
	}
	fmt.Printf("  Token: %s***%s\n", token[:8], token[len(token)-4:])

	// 1. Test getMe — verify token is valid
	fmt.Println("\n── 1. Testing getMe (token validity) ──")
	resp, err := http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/getMe", token))
	if err != nil {
		fmt.Printf("  ❌ HTTP error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 401 {
		fmt.Println("  ❌ 401 Unauthorized — Token is INVALID or REVOKED")
		fmt.Println("  → Go to @BotFather → /mybots → select your bot → API Token → Revoke & regenerate")
		return
	}
	if resp.StatusCode != 200 {
		fmt.Printf("  ❌ API returned %d: %s\n", resp.StatusCode, string(body))
		return
	}

	var meResp struct {
		OK     bool `json:"ok"`
		Result struct {
			ID        int64  `json:"id"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
			IsBot     bool   `json:"is_bot"`
		} `json:"result"`
	}
	json.Unmarshal(body, &meResp)

	if !meResp.OK {
		fmt.Printf("  ❌ API returned ok=false: %s\n", string(body))
		return
	}
	fmt.Printf("  ✅ Bot is valid!\n")
	fmt.Printf("  🤖 Name:     %s\n", meResp.Result.FirstName)
	fmt.Printf("  📛 Username: @%s\n", meResp.Result.Username)
	fmt.Printf("  🆔 Bot ID:   %d\n", meResp.Result.ID)

	// 2. Check for pending updates (to find your chat ID)
	fmt.Println("\n── 2. Checking recent updates (to find your chat ID) ──")
	resp2, err := http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?limit=5", token))
	if err != nil {
		fmt.Printf("  ⚠️  Could not fetch updates: %v\n", err)
	} else {
		defer resp2.Body.Close()
		body2, _ := io.ReadAll(resp2.Body)
		var updResp struct {
			OK     bool `json:"ok"`
			Result []struct {
				Message struct {
					Chat struct {
						ID        int64  `json:"id"`
						FirstName string `json:"first_name"`
						Username  string `json:"username"`
						Type      string `json:"type"`
					} `json:"chat"`
					Text string `json:"text"`
				} `json:"message"`
			} `json:"result"`
		}
		json.Unmarshal(body2, &updResp)

		if len(updResp.Result) == 0 {
			fmt.Println("  ⚠️  No recent messages to the bot")
			fmt.Println("  → Open Telegram, find your bot, press /start, then re-run this script")
		} else {
			seen := map[int64]bool{}
			for _, u := range updResp.Result {
				cid := u.Message.Chat.ID
				if cid == 0 || seen[cid] {
					continue
				}
				seen[cid] = true
				fmt.Printf("  📩 Chat ID: %d | Name: %s | @%s | Type: %s | Msg: %q\n",
					cid, u.Message.Chat.FirstName, u.Message.Chat.Username, u.Message.Chat.Type, u.Message.Text)
			}
			fmt.Println("\n  💡 Use the Chat ID above to set telegram_chat_id for your admin user in DB:")
			fmt.Println("     UPDATE users SET telegram_chat_id = <CHAT_ID> WHERE email = 'your-admin@email.com';")
		}
	}

	// 3. If a chat_id was provided as argument, send a test message
	if len(os.Args) > 1 {
		chatID := os.Args[1]
		fmt.Printf("\n── 3. Sending test message to chat %s ──\n", chatID)

		msg := "🔔 <b>XPLR Admin Notification Test</b>\n\nЭто тестовое уведомление из системы XPLR.\nЕсли вы видите это сообщение — Telegram-бот работает корректно! ✅"
		encodedMsg := url.QueryEscape(msg)
		apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&parse_mode=HTML&text=%s",
			token, chatID, encodedMsg)

		resp3, err := http.Get(apiURL)
		if err != nil {
			fmt.Printf("  ❌ HTTP error: %v\n", err)
			return
		}
		defer resp3.Body.Close()
		body3, _ := io.ReadAll(resp3.Body)

		if resp3.StatusCode == 403 {
			fmt.Println("  ❌ 403 Forbidden — The user has BLOCKED the bot or hasn't pressed /start")
			fmt.Println("  → Open Telegram, find your bot, press /start, then retry")
			return
		}
		if resp3.StatusCode != 200 {
			fmt.Printf("  ❌ API returned %d: %s\n", resp3.StatusCode, string(body3))
			return
		}
		fmt.Println("  ✅ Test message sent successfully!")
		fmt.Println("  → Check your Telegram for the notification")
	} else {
		fmt.Println("\n── 3. Send test message ──")
		fmt.Println("  ℹ️  To send a test message, re-run with a chat ID:")
		fmt.Println("     go run backend/cmd/tg_test/main.go <CHAT_ID>")
	}

	fmt.Println("\n═══════════════════════════════════════════")
	fmt.Println("  Diagnostic complete")
	fmt.Println("═══════════════════════════════════════════")
}
