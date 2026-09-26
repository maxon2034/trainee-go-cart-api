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

		// 1. Проверка существования корзины
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// 2. Подсчет количества элементов
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM cart_items WHERE cart_id=\$1`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		// 3. Вставка товара
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

	t.Run("error - cart not found", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		nonExistentCartID := uuid.New()

		// Возвращаем false из проверки EXISTS
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(nonExistentCartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		item, err := repo.AddCartItem(context.Background(), nonExistentCartID, "Headphones", 80.00)

		require.Error(t, err)
		assert.Nil(t, item)
		assert.ErrorIs(t, err, errs.ErrCartNotFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error - cart exists query failed", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()
		expectedErr := errors.New("db connection timeout")

		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnError(expectedErr)

		item, err := repo.AddCartItem(context.Background(), cartID, "Monitor", 300.00)

		require.Error(t, err)
		assert.Nil(t, item)
		assert.ErrorIs(t, err, expectedErr)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error - cart limit reached", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()

		// 1. Корзина существует
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// 2. В корзине уже 5 элементов
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

		// 1. Корзина существует
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// 2. Ошибка на этапе подсчета
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

		// 1. Корзина существует
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// 2. Подсчет успешный
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM cart_items WHERE cart_id=\$1`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		// 3. Ошибка на INSERT
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

		// 1. Ожидаем проверку существования корзины
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// 2. Ожидаем SELECT по ID товара и ID корзины
		mock.ExpectQuery(`SELECT cart_id,id,product,price FROM cart_items WHERE id=\$1 AND cart_id=\$2`).
			WithArgs(itemID, cartID).
			WillReturnRows(sqlmock.NewRows([]string{"cart_id", "id", "product", "price"}).
				AddRow(cartID, itemID, oldProduct, oldPrice))

		// 3. Ожидаем UPDATE: SET product=$1, price=$2 WHERE id=$3 RETURNING...
		// Аргументы: newProduct, newPrice, itemID
		mock.ExpectQuery(`UPDATE cart_items SET product=\$1, price=\$2 WHERE id=\$3 RETURNING id,cart_id,product,price`).
			WithArgs(newProduct, newPrice, itemID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "cart_id", "product", "price"}).
				AddRow(itemID, cartID, newProduct, newPrice))

		item, err := repo.UpdateCartItem(context.Background(), cartID, itemID, newProduct, newPrice)

		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Equal(t, itemID, item.ID)
		assert.Equal(t, cartID, item.CartID)
		assert.Equal(t, newProduct, item.Product)
		assert.Equal(t, newPrice, item.Price)

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

		// EXISTS возвращает false
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		item, err := repo.UpdateCartItem(context.Background(), cartID, itemID, "Shoes", 5000.50)

		require.Error(t, err)
		assert.Nil(t, item)
		assert.ErrorIs(t, err, errs.ErrCartNotFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should fail if product does not exist (item not found in DB)", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()
		nonExistentID := uuid.New()

		// 1. Корзина существует
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// 2. SELECT возвращает sql.ErrNoRows
		mock.ExpectQuery(`SELECT cart_id,id,product,price FROM cart_items WHERE id=\$1 AND cart_id=\$2`).
			WithArgs(nonExistentID, cartID).
			WillReturnError(sql.ErrNoRows)

		item, err := repo.UpdateCartItem(context.Background(), cartID, nonExistentID, "Shoes", 5000.50)

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

		// 1. Корзина существует
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// 2. Возвращаем из БД запись с отрицательной ценой
		mock.ExpectQuery(`SELECT cart_id,id,product,price FROM cart_items WHERE id=\$1 AND cart_id=\$2`).
			WithArgs(itemID, cartID).
			WillReturnRows(sqlmock.NewRows([]string{"cart_id", "id", "product", "price"}).
				AddRow(cartID, itemID, "Broken Product", -10.00))

		item, err := repo.UpdateCartItem(context.Background(), cartID, itemID, "Shoes", 5000.50)

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

		// 1. Корзина существует
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// 2. Возвращаем из БД запись с пустым наименованием
		mock.ExpectQuery(`SELECT cart_id,id,product,price FROM cart_items WHERE id=\$1 AND cart_id=\$2`).
			WithArgs(itemID, cartID).
			WillReturnRows(sqlmock.NewRows([]string{"cart_id", "id", "product", "price"}).
				AddRow(cartID, itemID, "", 100.00))

		item, err := repo.UpdateCartItem(context.Background(), cartID, itemID, "Shoes", 5000.50)

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

		// 1. Корзина существует
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// 2. SELECT проходит
		mock.ExpectQuery(`SELECT cart_id,id,product,price FROM cart_items WHERE id=\$1 AND cart_id=\$2`).
			WithArgs(itemID, cartID).
			WillReturnRows(sqlmock.NewRows([]string{"cart_id", "id", "product", "price"}).
				AddRow(cartID, itemID, "Old Product", 100.00))

		// 3. UPDATE отдает ошибку
		mock.ExpectQuery(`UPDATE cart_items SET product=\$1, price=\$2 WHERE id=\$3 RETURNING id,cart_id,product,price`).
			WithArgs(newProduct, newPrice, itemID).
			WillReturnError(expectedErr)

		item, err := repo.UpdateCartItem(context.Background(), cartID, itemID, newProduct, newPrice)

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

		cartID := uuid.New()
		itemID := uuid.New()

		// 1. Ожидаем проверку существования корзины
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// 2. Ожидаем DELETE с аргументами itemID ($1) и cartID ($2)
		mock.ExpectExec(`DELETE FROM cart_items WHERE id=\$1 AND cart_id=\$2`).
			WithArgs(itemID, cartID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = repo.RemoveCartItem(context.Background(), cartID, itemID)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should fail if cart does not exist", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		nonExistentCartID := uuid.New()
		itemID := uuid.New()

		// EXISTS возвращает false
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(nonExistentCartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		err = repo.RemoveCartItem(context.Background(), nonExistentCartID, itemID)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrCartNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error - cart exists query failed", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()
		itemID := uuid.New()
		expectedErr := errors.New("db connection timeout")

		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnError(expectedErr)

		err = repo.RemoveCartItem(context.Background(), cartID, itemID)

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
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

		// 1. Корзина существует
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// 2. DELETE отработал, но 0 строк удалено
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

		// 1. Корзина существует
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// 2. Ошибка выполнения SQL-запроса DELETE
		mock.ExpectExec(`DELETE FROM cart_items WHERE id=\$1 AND cart_id=\$2`).
			WithArgs(itemID, cartID).
			WillReturnError(expectedErr)

		err = repo.RemoveCartItem(context.Background(), cartID, itemID)

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

		cartID := uuid.New()
		itemID := uuid.New()
		expectedErr := errors.New("rows affected error")

		// 1. Корзина существует
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// 2. Запрос прошел, но проверка RowsAffected вернет ошибку
		mock.ExpectExec(`DELETE FROM cart_items WHERE id=\$1 AND cart_id=\$2`).
			WithArgs(itemID, cartID).
			WillReturnResult(sqlmock.NewErrorResult(expectedErr))

		err = repo.RemoveCartItem(context.Background(), cartID, itemID)

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCartRepository_CalculateDiscount(t *testing.T) {
	t.Run("success - discount 10 percent (total > 5000)", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()

		// 1. Проверка существования корзины
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// 2. Получение цен товаров (сумма 6000 > 5000, 2 товара <= 3) -> скидка 10%
		mock.ExpectQuery(`SELECT price FROM cart_items WHERE cart_id=\$1`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"price"}).
				AddRow(4000.0).
				AddRow(2000.0))

		res, err := repo.CalculateDiscount(context.Background(), cartID)

		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, cartID, res.CartID)
		assert.Equal(t, 6000.0, res.TotalPrice)
		assert.Equal(t, 0.1, res.DiscountPercent)
		assert.Equal(t, 5400.0, res.FinalPrice)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - discount 5 percent (items count > 3)", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()

		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// Сумма 4000 (<= 5000), но товаров 4 (> 3) -> скидка 5%
		mock.ExpectQuery(`SELECT price FROM cart_items WHERE cart_id=\$1`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"price"}).
				AddRow(1000.0).
				AddRow(1000.0).
				AddRow(1000.0).
				AddRow(1000.0))

		res, err := repo.CalculateDiscount(context.Background(), cartID)

		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, cartID, res.CartID)
		assert.Equal(t, 4000.0, res.TotalPrice)
		assert.Equal(t, 0.05, res.DiscountPercent)
		assert.Equal(t, 3800.0, res.FinalPrice)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - no discount", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()

		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// Сумма 3000 (<= 5000), товаров 2 (<= 3) -> скидка 0%
		mock.ExpectQuery(`SELECT price FROM cart_items WHERE cart_id=\$1`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"price"}).
				AddRow(1500.0).
				AddRow(1500.0))

		res, err := repo.CalculateDiscount(context.Background(), cartID)

		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, cartID, res.CartID)
		assert.Equal(t, 3000.0, res.TotalPrice)
		assert.Equal(t, 0.0, res.DiscountPercent)
		assert.Equal(t, 3000.0, res.FinalPrice)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should fail if cart does not exist", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		nonExistentCartID := uuid.New()

		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(nonExistentCartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		res, err := repo.CalculateDiscount(context.Background(), nonExistentCartID)

		require.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrCartNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should fail if cart has no items (sql.ErrNoRows)", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()

		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		mock.ExpectQuery(`SELECT price FROM cart_items WHERE cart_id=\$1`).
			WithArgs(cartID).
			WillReturnError(sql.ErrNoRows)

		res, err := repo.CalculateDiscount(context.Background(), cartID)

		require.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrEmptyCart)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error - cart exists query failed", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()
		expectedErr := errors.New("db connection timeout")

		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnError(expectedErr)

		res, err := repo.CalculateDiscount(context.Background(), cartID)

		require.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, expectedErr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error - fetch prices query failed", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		repo := repository.New(sqlxDB)

		cartID := uuid.New()
		expectedErr := errors.New("db internal error")

		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM carts WHERE id = \$1\)`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		mock.ExpectQuery(`SELECT price FROM cart_items WHERE cart_id=\$1`).
			WithArgs(cartID).
			WillReturnError(expectedErr)

		res, err := repo.CalculateDiscount(context.Background(), cartID)

		require.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, expectedErr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
