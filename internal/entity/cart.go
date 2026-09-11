package entity

import "errors"

var ErrCartFull = errors.New("cart exceeds maximum items limit")

type Cart struct {
	ID    int        `json:"id"`
	Items []CartItem `json:"items"`
}

func (c *Cart) AddItem(item CartItem) error {
	if len(c.Items) >= 5 {
		return ErrCartFull
	}
	c.Items = append(c.Items, item)
	return nil
}
