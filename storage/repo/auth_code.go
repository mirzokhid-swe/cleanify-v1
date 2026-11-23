package repo

import "time"

// AuthCodeI: Telegram auth bot beradigan bir martalik kirish kodlari.
type AuthCodeI interface {
	Create(userID int64, code string, expiresAt time.Time) error
	// Use: faol (ishlatilmagan, muddati o'tmagan) kodni bir martalik ishlatadi
	// va egasining user_id sini qaytaradi; topilmasa sql.ErrNoRows.
	Use(code string) (userID int64, err error)
}
