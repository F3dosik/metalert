package grpcserver

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// fakeHandler всегда возвращает nil — используется как конечный обработчик в перехватчике.
func fakeHandler(_ context.Context, req any) (any, error) {
	return req, nil
}

func TestSubnetInterceptor(t *testing.T) {
	_, allowedNet, _ := net.ParseCIDR("192.168.1.0/24")

	interceptor := SubnetInterceptor(allowedNet)
	info := &grpc.UnaryServerInfo{FullMethod: "/metrics.Metrics/UpdateMetrics"}

	tests := []struct {
		name     string
		ip       string
		wantCode codes.Code
	}{
		{
			name:     "IP в доверенной подсети",
			ip:       "192.168.1.42",
			wantCode: codes.OK,
		},
		{
			name:     "IP вне доверенной подсети",
			ip:       "10.0.0.1",
			wantCode: codes.PermissionDenied,
		},
		{
			name:     "некорректный IP",
			ip:       "not-an-ip",
			wantCode: codes.PermissionDenied,
		},
		{
			name:     "пустой x-real-ip",
			ip:       "",
			wantCode: codes.PermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := metadata.MD{}
			if tt.ip != "" {
				md.Set("x-real-ip", tt.ip)
			}
			ctx := metadata.NewIncomingContext(context.Background(), md)

			_, err := interceptor(ctx, nil, info, fakeHandler)

			got := status.Code(err)
			if got != tt.wantCode {
				t.Errorf("got code %v, want %v", got, tt.wantCode)
			}
		})
	}
}

func TestSubnetInterceptor_NoSubnet(t *testing.T) {
	// Если подсеть не задана — пропускаем любой запрос без метаданных.
	interceptor := SubnetInterceptor(nil)
	info := &grpc.UnaryServerInfo{}

	ctx := context.Background()
	_, err := interceptor(ctx, nil, info, fakeHandler)
	if err != nil {
		t.Errorf("expected no error when subnet is nil, got: %v", err)
	}
}

func TestSubnetInterceptor_NoMetadata(t *testing.T) {
	_, subnet, _ := net.ParseCIDR("10.0.0.0/8")
	interceptor := SubnetInterceptor(subnet)
	info := &grpc.UnaryServerInfo{}

	// Контекст без metadata.
	_, err := interceptor(context.Background(), nil, info, fakeHandler)
	if status.Code(err) != codes.PermissionDenied {
		t.Errorf("expected PermissionDenied without metadata, got: %v", err)
	}
}
