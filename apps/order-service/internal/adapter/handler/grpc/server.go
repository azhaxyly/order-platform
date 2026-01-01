package grpc

import (
	"context"

	pb "order-platform/api/order"
	"order-platform/apps/order-service/internal/domain"
	"order-platform/apps/order-service/internal/service"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderGrpcServer struct {
	pb.UnimplementedOrderServiceServer
	service *service.OrderService
}

func NewOrderGrpcServer(service *service.OrderService) *OrderGrpcServer {
	return &OrderGrpcServer{service: service}
}

func (s *OrderGrpcServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	items := make([]domain.OrderItem, len(req.Items))
	for i, item := range req.Items {
		prodID, err := uuid.Parse(item.ProductId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid product_id at index %d", i)
		}
		items[i] = domain.OrderItem{
			ProductID: prodID,
			Quantity:  item.Quantity,
			Price:     item.PriceCents,
		}
	}

	order, err := s.service.CreateOrder(ctx, userID, items)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create order: %v", err)
	}

	return &pb.CreateOrderResponse{
		OrderId: order.ID.String(),
		Status:  string(order.Status),
	}, nil
}
