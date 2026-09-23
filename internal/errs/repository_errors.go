package errs

import "errors"

var ErrCartNotFound = errors.New("cart not found")
var ErrFullCart = errors.New("unable to insert item into cart: cart is full")
var ErrEmptyProduct = errors.New("product is empty")
var ErrNegativePrice = errors.New("price is negative")
