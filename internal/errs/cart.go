package errs

import "errors"

var ErrCartNotFound = errors.New("cart not found")
var ErrFullCart = errors.New("unable to insert item into cart: cart is full")
