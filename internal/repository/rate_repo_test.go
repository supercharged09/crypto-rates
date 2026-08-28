package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/supercharged09/crypto-rates/internal/model"
)

/*
TestRateRepository_Integration интеграц. тест с реальной БД
перед запуском, запустить postgreSQL
*/
func TestRateRepository_Integration(t *testing.T) {
	//подключение к тестовой БД
	db, err := setupTestDB()
	if err != nil {
		t.Skipf("Skipping integration test: DB not available: %v", err)
	}
	defer db.Close()

	repo := NewRateRepository(db)
	ctx := context.Background()

	//очистка таблицы перед тестом
	db.ExecContext(ctx, "DELETE FROM rates WHERE cryptocurrency = 'testcoin'")

	//test save
	rate := model.Rate{
		Cryptocurrency: "testcoin",
		PriceUSD:       50000.00,
		Timestamp:      time.Now(),
	}

	err = repo.Save(ctx, rate)
	assert.NoError(t, err)

	//test GetCurrentPrice
	current, err := repo.GetCurrentPrice(ctx, "testcoin")
	assert.NoError(t, err)
	assert.NotNil(t, current)
	assert.Equal(t, 50000.00, current.PriceUSD)
	assert.Equal(t, "testcoin", current.Cryptocurrency)

	//test GetMinMax24h
	min, max, err := repo.GetMinMax24h(ctx, "testcoin")
	assert.NoError(t, err)
	assert.Equal(t, 50000.00, min)
	assert.Equal(t, 50000.00, max)

	//test GetPriceHourAgo (данных нет)
	price, err := repo.GetPriceHourAgo(ctx, "testcoin")
	assert.NoError(t, err)
	assert.Equal(t, 0.0, price)

	//очистка
	db.ExecContext(ctx, "DELETE FROM rates WHERE cryptocurrency = 'testcoin'")

}
