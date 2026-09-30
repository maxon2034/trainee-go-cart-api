package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/maxon2034/trainee-go-cart-api/internal/errs"
	"github.com/maxon2034/trainee-go-cart-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCartRepository_AddCart(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)
		expectedID := uuid.New()

		mock.ExpectQuery(`INSERT INTO carts DEFAULT VALUES RETURNING id;`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedID))

		cart, err := repo.AddCart(context.Background())

		require.NoError(t, err)
		require.NotNil(t, cart)
		assert.Equal(t, expectedID, cart.ID)
		assert.NotNil(t, cart.Items)
		assert.Empty(t, cart.Items)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)
		expectedErr := errors.New("db connection timeout")

		mock.ExpectQuery(`INSERT INTO carts DEFAULT VALUES RETURNING id;`).
			WillReturnError(expectedErr)

		cart, err := repo.AddCart(context.Background())

		require.Error(t, err)
		assert.Nil(t, cart)
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorContains(t, err, "r.AddCart")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCartRepository_GetCart(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()
		itemID1 := uuid.New()
		itemID2 := uuid.New()

		mock.ExpectQuery(`SELECT c\.id AS cart_id, i\.id AS id, i\.product AS product, i\.price AS price FROM carts c LEFT JOIN cart_items i ON c\.id = i\.cart_id WHERE c\.id = \$1`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"cart_id", "id", "product", "price"}).
				AddRow(cartID, itemID1, "Laptop", float64(1000.00)).
				AddRow(cartID, itemID2, "Mouse", float64(50.00)))

		cart, err := repo.GetCart(context.Background(), cartID)

		require.NoError(t, err)
		require.NotNil(t, cart)
		assert.Equal(t, cartID, cart.ID)
		assert.Len(t, cart.Items, 2)

		require.NotNil(t, cart.Items[0].ID)
		assert.Equal(t, itemID1, *cart.Items[0].ID)
		require.NotNil(t, cart.Items[0].Product)
		assert.Equal(t, "Laptop", *cart.Items[0].Product)
		require.NotNil(t, cart.Items[0].Price)
		assert.Equal(t, float64(1000.00), *cart.Items[0].Price)

		require.NotNil(t, cart.Items[1].ID)
		assert.Equal(t, itemID2, *cart.Items[1].ID)
		require.NotNil(t, cart.Items[1].Product)
		assert.Equal(t, "Mouse", *cart.Items[1].Product)
		require.NotNil(t, cart.Items[1].Price)
		assert.Equal(t, float64(50.00), *cart.Items[1].Price)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("cart not found", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()

		mock.ExpectQuery(`SELECT c\.id AS cart_id, i\.id AS id, i\.product AS product, i\.price AS price FROM carts c LEFT JOIN cart_items i ON c\.id = i\.cart_id WHERE c\.id = \$1`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"cart_id", "id", "product", "price"}))

		cart, err := repo.GetCart(context.Background(), cartID)

		require.Error(t, err)
		assert.Nil(t, cart)
		assert.ErrorIs(t, err, errs.ErrCartNotFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()
		expectedErr := errors.New("db connection lost")

		mock.ExpectQuery(`SELECT c\.id AS cart_id, i\.id AS id, i\.product AS product, i\.price AS price FROM carts c LEFT JOIN cart_items i ON c\.id = i\.cart_id WHERE c\.id = \$1`).
			WithArgs(cartID).
			WillReturnError(expectedErr)

		cart, err := repo.GetCart(context.Background(), cartID)

		require.Error(t, err)
		assert.Nil(t, cart)
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorContains(t, err, "r.GetCart")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCartRepository_AddCartItem(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()
		itemID := uuid.New()
		product := "Mechanical Keyboard"
		price := float64(150.00)
		itemLimit := 5

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM cart_items WHERE cart_id=\$1 FOR UPDATE`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		mock.ExpectQuery(`INSERT INTO cart_items \(cart_id, product, price\) VALUES \(\$1, \$2, \$3\) RETURNING id,cart_id, product, price`).
			WithArgs(cartID, product, price).
			WillReturnRows(sqlmock.NewRows([]string{"id", "cart_id", "product", "price"}).
				AddRow(itemID, cartID, product, price))

		item, err := repo.AddCartItem(context.Background(), cartID, product, price, itemLimit)

		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Equal(t, itemID, *item.ID)
		assert.Equal(t, cartID, item.CartID)
		assert.Equal(t, product, *item.Product)
		assert.Equal(t, price, *item.Price)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error - cart not found (fk violation)", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		nonExistentCartID := uuid.New()
		product := "Headphones"
		price := float64(80.00)
		itemLimit := 5

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM cart_items WHERE cart_id=\$1 FOR UPDATE`).
			WithArgs(nonExistentCartID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		mock.ExpectQuery(`INSERT INTO cart_items \(cart_id, product, price\) VALUES \(\$1, \$2, \$3\) RETURNING id,cart_id, product, price`).
			WithArgs(nonExistentCartID, product, price).
			WillReturnError(&pgconn.PgError{Code: "23503"})

		item, err := repo.AddCartItem(context.Background(), nonExistentCartID, product, price, itemLimit)

		require.Error(t, err)
		assert.Nil(t, item)
		assert.ErrorIs(t, err, errs.ErrCartNotFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error - cart limit reached", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()
		itemLimit := 5

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM cart_items WHERE cart_id=\$1 FOR UPDATE`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

		item, err := repo.AddCartItem(context.Background(), cartID, "Headphones", 80.00, itemLimit)

		require.Error(t, err)
		assert.Nil(t, item)
		assert.ErrorIs(t, err, errs.ErrFullCart)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error - count query failed", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()
		expectedErr := errors.New("connection failed")
		itemLimit := 5

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM cart_items WHERE cart_id=\$1 FOR UPDATE`).
			WithArgs(cartID).
			WillReturnError(expectedErr)

		item, err := repo.AddCartItem(context.Background(), cartID, "Monitor", 300.00, itemLimit)

		require.Error(t, err)
		assert.Nil(t, item)
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorContains(t, err, "r.AddCartItem")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error - insert query failed", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()
		product := "Webcam"
		price := float64(60.00)
		itemLimit := 5
		expectedErr := errors.New("some unexpected db error")

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM cart_items WHERE cart_id=\$1 FOR UPDATE`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(`INSERT INTO cart_items \(cart_id, product, price\) VALUES \(\$1, \$2, \$3\) RETURNING id,cart_id, product, price`).
			WithArgs(cartID, product, price).
			WillReturnError(expectedErr)

		item, err := repo.AddCartItem(context.Background(), cartID, product, price, itemLimit)

		require.Error(t, err)
		assert.Nil(t, item)
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorContains(t, err, "r.AddCartItem")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCartRepository_UpdateCartItem(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		itemID := uuid.New()
		cartID := uuid.New()
		oldProduct := "Old Shoes"
		oldPrice := float64(1000.00)

		newProduct := "Shoes"
		newPrice := float64(5000.50)

		mock.ExpectQuery(`SELECT c\.id AS cart_id, i\.id AS id, i\.product, i\.price FROM carts c LEFT JOIN cart_items i ON i\.cart_id = c\.id AND i\.id =\$1 WHERE c\.id = \$2`).
			WithArgs(itemID, cartID).
			WillReturnRows(sqlmock.NewRows([]string{"cart_id", "id", "product", "price"}).
				AddRow(cartID, itemID, oldProduct, oldPrice))

		mock.ExpectQuery(`UPDATE cart_items SET product=\$1, price=\$2 WHERE id=\$3 RETURNING id,cart_id,product,price`).
			WithArgs(newProduct, newPrice, itemID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "cart_id", "product", "price"}).
				AddRow(itemID, cartID, newProduct, newPrice))

		item, err := repo.UpdateCartItem(context.Background(), cartID, itemID, newProduct, newPrice)

		require.NoError(t, err)
		require.NotNil(t, item)

		// Для указателей проверяем Nil перед разыменованием
		require.NotNil(t, item.ID)
		assert.Equal(t, itemID, *item.ID)

		// CartID — обычное значение, передаем без &
		assert.Equal(t, cartID, item.CartID)

		require.NotNil(t, item.Product)
		assert.Equal(t, newProduct, *item.Product)

		require.NotNil(t, item.Price)
		assert.Equal(t, newPrice, *item.Price)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should fail if cart does not exist", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()
		itemID := uuid.New()

		mock.ExpectQuery(`SELECT c\.id AS cart_id, i\.id AS id, i\.product, i\.price FROM carts c LEFT JOIN cart_items i ON i\.cart_id = c\.id AND i\.id =\$1 WHERE c\.id = \$2`).
			WithArgs(itemID, cartID).
			WillReturnError(sql.ErrNoRows)

		item, err := repo.UpdateCartItem(context.Background(), cartID, itemID, "Shoes", 5000.50)

		require.Error(t, err)
		assert.Nil(t, item)
		assert.ErrorIs(t, err, errs.ErrCartNotFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should fail if product does not exist in cart (item_id is nil)", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()
		nonExistentID := uuid.New()

		mock.ExpectQuery(`SELECT c\.id AS cart_id, i\.id AS id, i\.product, i\.price FROM carts c LEFT JOIN cart_items i ON i\.cart_id = c\.id AND i\.id =\$1 WHERE c\.id = \$2`).
			WithArgs(nonExistentID, cartID).
			WillReturnRows(sqlmock.NewRows([]string{"cart_id", "id", "product", "price"}).
				AddRow(cartID, nil, nil, nil))

		item, err := repo.UpdateCartItem(context.Background(), cartID, nonExistentID, "Shoes", 5000.50)

		require.Error(t, err)
		assert.Nil(t, item)
		assert.ErrorIs(t, err, errs.ErrCartItemNotFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error - SELECT query failed", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()
		itemID := uuid.New()
		expectedErr := errors.New("db connection lost")

		mock.ExpectQuery(`SELECT c\.id AS cart_id, i\.id AS id, i\.product, i\.price FROM carts c LEFT JOIN cart_items i ON i\.cart_id = c\.id AND i\.id =\$1 WHERE c\.id = \$2`).
			WithArgs(itemID, cartID).
			WillReturnError(expectedErr)

		item, err := repo.UpdateCartItem(context.Background(), cartID, itemID, "Shoes", 5000.50)

		require.Error(t, err)
		assert.Nil(t, item)
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorContains(t, err, "r.UpdateCartItem")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error - UPDATE query failed", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		itemID := uuid.New()
		cartID := uuid.New()
		newProduct := "Shoes"
		newPrice := float64(5000.50)
		expectedErr := errors.New("db update failure")

		mock.ExpectQuery(`SELECT c\.id AS cart_id, i\.id AS id, i\.product, i\.price FROM carts c LEFT JOIN cart_items i ON i\.cart_id = c\.id AND i\.id =\$1 WHERE c\.id = \$2`).
			WithArgs(itemID, cartID).
			WillReturnRows(sqlmock.NewRows([]string{"cart_id", "id", "product", "price"}).
				AddRow(cartID, itemID, "Old Product", 100.00))

		mock.ExpectQuery(`UPDATE cart_items SET product=\$1, price=\$2 WHERE id=\$3 RETURNING id,cart_id,product,price`).
			WithArgs(newProduct, newPrice, itemID).
			WillReturnError(expectedErr)

		item, err := repo.UpdateCartItem(context.Background(), cartID, itemID, newProduct, newPrice)

		require.Error(t, err)
		assert.Nil(t, item)
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorContains(t, err, "r.UpdateCartItem")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
func TestCartRepository_RemoveCartItem(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()
		itemID := uuid.New()

		mock.ExpectExec(`DELETE FROM cart_items WHERE id=\$1 AND cart_id=\$2`).
			WithArgs(itemID, cartID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = repo.RemoveCartItem(context.Background(), cartID, itemID)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should fail if cart does not exist (fk violation)", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		nonExistentCartID := uuid.New()
		itemID := uuid.New()

		mock.ExpectExec(`DELETE FROM cart_items WHERE id=\$1 AND cart_id=\$2`).
			WithArgs(itemID, nonExistentCartID).
			WillReturnError(&pgconn.PgError{Code: "23503"})

		err = repo.RemoveCartItem(context.Background(), nonExistentCartID, itemID)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrCartNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should fail if item does not exist (0 rows affected)", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()
		nonExistentItemID := uuid.New()

		mock.ExpectExec(`DELETE FROM cart_items WHERE id=\$1 AND cart_id=\$2`).
			WithArgs(nonExistentItemID, cartID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err = repo.RemoveCartItem(context.Background(), cartID, nonExistentItemID)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrCartItemNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error - DELETE query execution failed", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()
		itemID := uuid.New()
		expectedErr := errors.New("db delete execution error")

		mock.ExpectExec(`DELETE FROM cart_items WHERE id=\$1 AND cart_id=\$2`).
			WithArgs(itemID, cartID).
			WillReturnError(expectedErr)

		err = repo.RemoveCartItem(context.Background(), cartID, itemID)

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorContains(t, err, "r.RemoveCartItem")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error - RowsAffected failed", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()
		itemID := uuid.New()
		expectedErr := errors.New("rows affected error")

		mock.ExpectExec(`DELETE FROM cart_items WHERE id=\$1 AND cart_id=\$2`).
			WithArgs(itemID, cartID).
			WillReturnResult(sqlmock.NewErrorResult(expectedErr))

		err = repo.RemoveCartItem(context.Background(), cartID, itemID)

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorContains(t, err, "r.RemoveCartItem")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
