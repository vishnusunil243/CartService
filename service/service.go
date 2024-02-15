package service

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/vishnusunil243/CartService/adapters"
	entities "github.com/vishnusunil243/CartService/entity"
	"github.com/vishnusunil243/proto-files/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	Tracer        opentracing.Tracer
	ProductClient pb.ProductServiceClient
)

func RetrieveTracer(tr opentracing.Tracer) {
	Tracer = tr
}

type CartService struct {
	Adapter adapters.AdapterInterface
	pb.UnimplementedCartServiceServer
}

func NewCartService(adapter adapters.AdapterInterface) *CartService {
	return &CartService{
		Adapter: adapter,
	}
}
func (cart *CartService) CreateCart(ctx context.Context, req *pb.UserCartCreate) (*pb.CartResponse, error) {
	span := Tracer.StartSpan("create cart grpc")
	defer span.Finish()
	err := cart.Adapter.CreateCart(int(req.UserId))
	if err != nil {
		return &pb.CartResponse{}, err
	}
	res := &pb.CartResponse{
		UserId: req.UserId,
	}
	return res, nil
}
func (cart *CartService) AddToCart(ctx context.Context, req *pb.AddToCartRequest) (*pb.CartResponse, error) {
	productData, err := ProductClient.GetProduct(context.TODO(), &pb.GetProductById{Id: int32(req.ProductId)})
	if err != nil {
		return &pb.CartResponse{}, fmt.Errorf("error fetching product data")
	}
	if productData.Name == "" {
		return &pb.CartResponse{}, fmt.Errorf("product not found")
	}
	if productData.Quantity < req.Quantity {
		return &pb.CartResponse{}, fmt.Errorf("not enough quantity is available to add in cart please decrease the quantity")
	}
	reqEntity := entities.Cart_items{
		ProductId: uint(req.ProductId),
		Total:     float64(productData.Price),
		Quantity:  int(req.Quantity),
	}
	fmt.Println(reqEntity)
	err = cart.Adapter.AddToCart(reqEntity, int(req.UserId))
	if err != nil {
		return &pb.CartResponse{}, err
	}
	res := &pb.CartResponse{
		UserId:  req.UserId,
		IsEmpty: false,
	}

	return res, nil
}
func (cart *CartService) RemoveFromCart(ctx context.Context, req *pb.RemoveFromCartRequest) (*pb.CartResponse, error) {
	productData, err := ProductClient.GetProduct(context.TODO(), &pb.GetProductById{Id: int32(req.ProductId)})
	if err != nil {
		return &pb.CartResponse{}, fmt.Errorf("error fetching products")
	}
	if productData.Name == "" {
		return &pb.CartResponse{}, fmt.Errorf("no product found with the given id")
	}
	reqEntity := entities.Cart_items{
		ProductId: uint(req.ProductId),
		Total:     float64(productData.Price),
	}
	err = cart.Adapter.RemoveFromCart(reqEntity, int(req.UserId))
	if err != nil {
		return &pb.CartResponse{}, err
	}
	res := &pb.CartResponse{
		UserId:  req.UserId,
		IsEmpty: cart.Adapter.IsEmpty(reqEntity, int(req.UserId)),
	}

	return res, nil

}
func (cart *CartService) GetAllCartItems(req *pb.UserCartCreate, srv pb.CartService_GetAllCartItemsServer) error {
	cartItems, err := cart.Adapter.GetAllCartItems(int(req.UserId))
	if err != nil {
		return err
	}
	for _, item := range cartItems {
		if err := srv.Send(&pb.GetAllCartResponse{
			UserId:    req.UserId,
			ProductId: uint32(item.ProductId),
			Quantity:  int32(item.Quantity),
			Total:     float32(item.Total),
		}); err != nil {
			return err
		}
	}
	return nil
}
func (cart *CartService) TruncateCart(ctx context.Context, req *pb.UserCartCreate) (*emptypb.Empty, error) {
	if err := cart.Adapter.TruncateCart(int(req.UserId)); err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

type HealthChecker struct {
	grpc_health_v1.UnimplementedHealthServer
}

func (s *HealthChecker) Check(ctx context.Context, in *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	fmt.Println("check called")
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (s *HealthChecker) Watch(in *grpc_health_v1.HealthCheckRequest, srv grpc_health_v1.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watching is not supported")
}
