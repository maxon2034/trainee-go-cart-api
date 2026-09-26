package tests

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
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

		mock.ExpectQuery(`SELECT id FROM carts WHERE id = \$1`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(cartID))

		mock.ExpectQuery(`SELECT id,product,price FROM cart_items WHERE cart_id=\$1`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "product", "price"}).
				AddRow(itemID1, "Laptop", float64(1000.00)).
				AddRow(itemID2, "Mouse", float64(50.00)))

		cart, err := repo.GetCart(context.Background(), cartID)

		require.NoError(t, err)
		require.NotNil(t, cart)
		assert.Equal(t, cartID, cart.ID)
		assert.Len(t, cart.Items, 2)

		assert.Equal(t, itemID1, cart.Items[0].ID)
		assert.Equal(t, "Laptop", cart.Items[0].Product)
		assert.Equal(t, float64(1000.00), cart.Items[0].Price)

		assert.Equal(t, itemID2, cart.Items[1].ID)
		assert.Equal(t, "Mouse", cart.Items[1].Product)
		assert.Equal(t, float64(50.00), cart.Items[1].Price)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("cart not found", func(t *testing.T) {
		var db sqlx.DB
		var err error
		var mock sqlmock.Sqlmock
		db.DB, mock, err = sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db.DB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()

		mock.ExpectQuery(`SELECT id FROM carts WHERE id = \$1`).
			WithArgs(cartID).
			WillReturnError(sql.ErrNoRows)

		cart, err := repo.GetCart(context.Background(), cartID)

		require.Error(t, err)
		assert.Nil(t, cart)
		assert.ErrorIs(t, err, errs.ErrCartNotFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("items fetch error", func(t *testing.T) {
		var db sqlx.DB
		var err error
		var mock sqlmock.Sqlmock
		db.DB, mock, err = sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db.DB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()
		expectedErr := errors.New("db connection lost")

		mock.ExpectQuery(`SELECT id FROM carts WHERE id = \$1`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(cartID))

		mock.ExpectQuery(`SELECT id,product,price FROM cart_items WHERE cart_id=\$1`).
			WithArgs(cartID).
			WillReturnError(expectedErr)

		cart, err := repo.GetCart(context.Background(), cartID)

		require.Error(t, err)
		assert.Nil(t, cart)
		assert.ErrorIs(t, err, expectedErr)

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

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM cart_items WHERE cart_id=\$1`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		mock.ExpectQuery(`INSERT INTO cart_items \(cart_id, product, price\) VALUES \(\$1, \$2, \$3\) RETURNING id,cart_id, product, price`).
			WithArgs(cartID, product, price).
			WillReturnRows(sqlmock.NewRows([]string{"id", "cart_id", "product", "price"}).
				AddRow(itemID, cartID, product, price))

		item, err := repo.AddCartItem(context.Background(), cartID, product, price)

		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Equal(t, itemID, item.ID)
		assert.Equal(t, cartID, item.CartID)
		assert.Equal(t, product, item.Product)
		assert.Equal(t, price, item.Price)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("validation error - empty product", func(t *testing.T) {
		mockDB, _, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		item, err := repo.AddCartItem(context.Background(), uuid.New(), "", 100.00)

		require.Error(t, err)
		assert.Nil(t, item)
		assert.ErrorIs(t, err, errs.ErrEmptyProduct)
	})

	t.Run("validation error - negative price", func(t *testing.T) {
		mockDB, _, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		item, err := repo.AddCartItem(context.Background(), uuid.New(), "Mouse", -10.00)

		require.Error(t, err)
		assert.Nil(t, item)
		assert.ErrorIs(t, err, errs.ErrNegativePrice)
	})

	t.Run("error - cart limit reached", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM cart_items WHERE cart_id=\$1`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

		item, err := repo.AddCartItem(context.Background(), cartID, "Headphones", 80.00)

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

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM cart_items WHERE cart_id=\$1`).
			WithArgs(cartID).
			WillReturnError(expectedErr)

		item, err := repo.AddCartItem(context.Background(), cartID, "Monitor", 300.00)

		require.Error(t, err)
		assert.Nil(t, item)
		assert.ErrorIs(t, err, expectedErr)

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
		expectedErr := errors.New("foreign key constraint violation")

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM cart_items WHERE cart_id=\$1`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(`INSERT INTO cart_items \(cart_id, product, price\) VALUES \(\$1, \$2, \$3\) RETURNING id,cart_id, product, price`).
			WithArgs(cartID, product, price).
			WillReturnError(expectedErr)

		item, err := repo.AddCartItem(context.Background(), cartID, product, price)

		require.Error(t, err)
		assert.Nil(t, item)
		assert.ErrorIs(t, err, expectedErr)

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

		// 1. Ожидаем SELECT по ID товара
		mock.ExpectQuery(`SELECT cart_id,id,product,price FROM cart_items WHERE id=\$1`).
			WithArgs(itemID).
			WillReturnRows(sqlmock.NewRows([]string{"cart_id", "id", "product", "price"}).
				AddRow(cartID, itemID, oldProduct, oldPrice))

		// 2. Ожидаем UPDATE: SET product=$1, price=$2 WHERE id=$3 RETURNING...
		// Аргументы в порядке $1, $2, $3: newProduct, newPrice, itemID
		mock.ExpectQuery(`UPDATE cart_items SET product=\$1, price=\$2 WHERE id=\$3 RETURNING id,cart_id,product,price`).
			WithArgs(newProduct, newPrice, itemID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "cart_id", "product", "price"}).
				AddRow(itemID, cartID, newProduct, newPrice))

		item, err := repo.UpdateCartItem(context.Background(), itemID, newProduct, newPrice)

		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Equal(t, itemID, item.ID)
		assert.Equal(t, cartID, item.CartID)
		assert.Equal(t, newProduct, item.Product)
		assert.Equal(t, newPrice, item.Price)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should fail if product does not exist (item not found in DB)", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		nonExistentID := uuid.New()

		// SELECT возвращает sql.ErrNoRows
		mock.ExpectQuery(`SELECT cart_id,id,product,price FROM cart_items WHERE id=\$1`).
			WithArgs(nonExistentID).
			WillReturnError(sql.ErrNoRows)

		item, err := repo.UpdateCartItem(context.Background(), nonExistentID, "Shoes", 5000.50)

		require.Error(t, err)
		assert.Nil(t, item)
		assert.ErrorIs(t, err, errs.ErrCartItemNotFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should fail if fetched item has negative price in DB", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		itemID := uuid.New()
		cartID := uuid.New()

		// Возвращаем из БД запись с отрицательной ценой
		mock.ExpectQuery(`SELECT cart_id,id,product,price FROM cart_items WHERE id=\$1`).
			WithArgs(itemID).
			WillReturnRows(sqlmock.NewRows([]string{"cart_id", "id", "product", "price"}).
				AddRow(cartID, itemID, "Broken Product", -10.00))

		item, err := repo.UpdateCartItem(context.Background(), itemID, "Shoes", 5000.50)

		require.Error(t, err)
		assert.Nil(t, item)
		assert.ErrorIs(t, err, errs.ErrNegativePrice)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should fail if fetched item product name is blank in DB", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		itemID := uuid.New()
		cartID := uuid.New()

		// Возвращаем из БД запись с пустым наименованием
		mock.ExpectQuery(`SELECT cart_id,id,product,price FROM cart_items WHERE id=\$1`).
			WithArgs(itemID).
			WillReturnRows(sqlmock.NewRows([]string{"cart_id", "id", "product", "price"}).
				AddRow(cartID, itemID, "", 100.00))

		item, err := repo.UpdateCartItem(context.Background(), itemID, "Shoes", 5000.50)

		require.Error(t, err)
		assert.Nil(t, item)
		assert.ErrorIs(t, err, errs.ErrEmptyProduct)

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

		// 1. SELECT проходит
		mock.ExpectQuery(`SELECT cart_id,id,product,price FROM cart_items WHERE id=\$1`).
			WithArgs(itemID).
			WillReturnRows(sqlmock.NewRows([]string{"cart_id", "id", "product", "price"}).
				AddRow(cartID, itemID, "Old Product", 100.00))

		// 2. UPDATE отдает ошибку
		// Аргументы в порядке $1, $2, $3: newProduct, newPrice, itemID
		mock.ExpectQuery(`UPDATE cart_items SET product=\$1, price=\$2 WHERE id=\$3 RETURNING id,cart_id,product,price`).
			WithArgs(newProduct, newPrice, itemID).
			WillReturnError(expectedErr)

		item, err := repo.UpdateCartItem(context.Background(), itemID, newProduct, newPrice)

		require.Error(t, err)
		assert.Nil(t, item)
		assert.ErrorIs(t, err, expectedErr)

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

		itemID := uuid.New()

		// Ожидаем DELETE с успешным удалением 1 строки
		mock.ExpectExec(`DELETE FROM cart_items WHERE id=\$1`).
			WithArgs(itemID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = repo.RemoveCartItem(context.Background(), itemID)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should fail if item does not exist (0 rows affected)", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		nonExistentID := uuid.New()

		// Запрос выполняется без ошибок БД, но удалено 0 строк
		mock.ExpectExec(`DELETE FROM cart_items WHERE id=\$1`).
			WithArgs(nonExistentID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err = repo.RemoveCartItem(context.Background(), nonExistentID)

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

		itemID := uuid.New()
		expectedErr := errors.New("db delete execution error")

		// Ошибка выполнения самого SQL-запроса
		mock.ExpectExec(`DELETE FROM cart_items WHERE id=\$1`).
			WithArgs(itemID).
			WillReturnError(expectedErr)

		err = repo.RemoveCartItem(context.Background(), itemID)

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error - RowsAffected failed", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		itemID := uuid.New()
		expectedErr := errors.New("rows affected error")

		// Запрос прошел, но получение количества измененных строк возвращает ошибку
		mock.ExpectExec(`DELETE FROM cart_items WHERE id=\$1`).
			WithArgs(itemID).
			WillReturnResult(sqlmock.NewErrorResult(expectedErr))

		err = repo.RemoveCartItem(context.Background(), itemID)

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
