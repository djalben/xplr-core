// Файл: utils/general.go
package utils

import (
	"net/http"
	"strings"
	"net"
)

// GetJWTSecret возвращает секретный ключ для подписи JWT
// (Предположим, что эта функция у вас уже где-то была, возможно, в utils/jwt.go,
// но на всякий случай, если вы ее сюда скопируете, это будет нормально)
// func GetJWTSecret() []byte {
//     return []byte(os.Getenv("JWT_SECRET")) 
// }


// GetClientIP пытается определить IP-адрес клиента, учитывая прокси-заголовки
func GetClientIP(r *http.Request) string {
	// 1. Проверяем стандартный заголовок прокси (часто используется балансировщиками)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// X-Forwarded-For может содержать список IP-адресов. Берем первый.
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	// 2. Проверяем заголовок, часто используемый Nginx
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// 3. Используем стандартный RemoteAddr
	// RemoteAddr имеет формат "ip:port"
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // Возвращаем полный адрес, если не удалось разделить
	}
	return ip
}