package entity

import "errors"

// Domain errors
var (
	ErrPackSize          = errors.New("pack size must be greater than 0")
	ErrPackNotFound      = errors.New("pack not found")
	ErrOrderNotFound     = errors.New("order not found")
	ErrInvalidQuantity   = errors.New("quantity must be greater than 0")
	ErrEmptyOrder        = errors.New("order cannot be empty")
	ErrInvalidAmount     = errors.New("amount must be greater than 0")
	ErrDuplicatePackSize = errors.New("pack size already exists")
)
