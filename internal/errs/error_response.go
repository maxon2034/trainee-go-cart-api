package errs

import "encoding/json"

type ErrorResponse struct {
	Error   string `json:"errs"`
	Message string `json:"message"`
}

func (e ErrorResponse) ToBytes() []byte {
	bytes, err := json.Marshal(e)
	if err != nil {
		return []byte(`{"errs":"INTERNAL_SERVER_ERROR","message":"failed to marshal errs response"}`)
	}
	return bytes
}

func NotFound() []byte {
	return ErrorResponse{
		Error:   "NOT_FOUND",
		Message: "Requested resource was not found",
	}.ToBytes()
}

func BadRequest() []byte {
	return ErrorResponse{
		Error:   "BAD_REQUEST",
		Message: "Requested resource was bad request",
	}.ToBytes()
}

func InternalServerError() []byte {

	return ErrorResponse{
		Error:   "INTERNAL_SERVER_ERROR",
		Message: "An internal server errs occurred. Please try again later.",
	}.ToBytes()
}

func FullCart() []byte {
	return ErrorResponse{
		Error:   "FULL_CART",
		Message: "Unable to add item to cart. Cart is full",
	}.ToBytes()
}

func NegativePrice() []byte {
	return ErrorResponse{
		Error:   "NEGATIVE_PRICE",
		Message: "Unable to add item to cart. Price can't be negative",
	}.ToBytes()
}

func EmptyProduct() []byte {
	return ErrorResponse{
		Error:   "EMPTY_PRODUCT",
		Message: "Product can't be empty",
	}.ToBytes()
}
