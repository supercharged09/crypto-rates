package model

// FloatToCents конвертер дуллеров в центы
func FloatToCents(price float64) int64 {
	return int64(price * 100)
}

// CentsToFloat конвертер центов в дуллеры
func CentsToFloat(cents int64) float64 {
	return float64(cents) / 100.0
}
