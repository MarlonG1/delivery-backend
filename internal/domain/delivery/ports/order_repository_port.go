package ports

import (
	"context"
	"domain/delivery/models/entities"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *entities.Order) error
	CreateQRData(ctx context.Context, qr *entities.QRCode) error
	GetOrderByID(ctx context.Context, id string) (*entities.Order, error)
	GetOrderByQR(ctx context.Context, qr *entities.QRCode) (*entities.Order, error)
	GetOrderByTrackingNumber(ctx context.Context, trackingNumber string) (*entities.Order, error)
	GetOrdersByUserID(ctx context.Context, userID string) ([]entities.Order, error)
	GetOrders(ctx context.Context) ([]entities.Order, error)
	GetLocationCoordinates(ctx context.Context, orderID string, addressType string) (float64, float64, error)
	UpdateOrder(ctx context.Context, orderID string, order *entities.Order) error
	DeleteOrder(ctx context.Context, id string) error
	ChangeStatus(ctx context.Context, id string, status string) error
	AssignDriverToOrder(ctx context.Context, orderID, driverID string) error
}
