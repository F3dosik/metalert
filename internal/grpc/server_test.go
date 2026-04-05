package grpcserver_test

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	grpcserver "github.com/F3dosik/metalert/internal/grpc"
	pb "github.com/F3dosik/metalert/internal/proto"
	"github.com/F3dosik/metalert/internal/repository"
)

const bufSize = 1024 * 1024

// newTestServer запускает gRPC-сервер на in-memory listener (bufconn)
// и возвращает клиент и функцию остановки.
func newTestServer(t *testing.T, trustedSubnet *net.IPNet) (pb.MetricsClient, func()) {
	t.Helper()

	lis := bufconn.Listen(bufSize)
	storage := repository.NewMemMetricsStorage()

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(grpcserver.SubnetInterceptor(trustedSubnet)),
	)
	pb.RegisterMetricsServer(srv, grpcserver.NewMetricsServer(storage))

	go srv.Serve(lis)

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc.NewClient: %v", err)
	}

	stop := func() {
		conn.Close()
		srv.Stop()
	}

	return pb.NewMetricsClient(conn), stop
}

// ctxWithIP добавляет x-real-ip в исходящие метаданные запроса.
func ctxWithIP(ip string) context.Context {
	md := metadata.Pairs("x-real-ip", ip)
	return metadata.NewOutgoingContext(context.Background(), md)
}

func TestUpdateMetrics_Gauge(t *testing.T) {
	client, stop := newTestServer(t, nil)
	defer stop()

	mtype := pb.Metric_GAUGE
	val := 3.14
	req := pb.UpdateMetricsRequest_builder{
		Metrics: []*pb.Metric{
			pb.Metric_builder{
				Id:    strPtr("cpu"),
				Type:  &mtype,
				Value: &val,
			}.Build(),
		},
	}.Build()

	_, err := client.UpdateMetrics(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateMetrics gauge: %v", err)
	}
}

func TestUpdateMetrics_Counter(t *testing.T) {
	client, stop := newTestServer(t, nil)
	defer stop()

	mtype := pb.Metric_COUNTER
	delta := int64(42)
	req := pb.UpdateMetricsRequest_builder{
		Metrics: []*pb.Metric{
			pb.Metric_builder{
				Id:    strPtr("requests"),
				Type:  &mtype,
				Delta: &delta,
			}.Build(),
		},
	}.Build()

	_, err := client.UpdateMetrics(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateMetrics counter: %v", err)
	}
}

func TestUpdateMetrics_Batch(t *testing.T) {
	client, stop := newTestServer(t, nil)
	defer stop()

	gaugeType := pb.Metric_GAUGE
	counterType := pb.Metric_COUNTER
	v1, v2 := 1.1, 2.2
	d1, d2 := int64(10), int64(20)

	req := pb.UpdateMetricsRequest_builder{
		Metrics: []*pb.Metric{
			pb.Metric_builder{Id: strPtr("m1"), Type: &gaugeType, Value: &v1}.Build(),
			pb.Metric_builder{Id: strPtr("m2"), Type: &gaugeType, Value: &v2}.Build(),
			pb.Metric_builder{Id: strPtr("c1"), Type: &counterType, Delta: &d1}.Build(),
			pb.Metric_builder{Id: strPtr("c2"), Type: &counterType, Delta: &d2}.Build(),
		},
	}.Build()

	_, err := client.UpdateMetrics(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateMetrics batch: %v", err)
	}
}

func TestUpdateMetrics_EmptyRequest(t *testing.T) {
	client, stop := newTestServer(t, nil)
	defer stop()

	req := pb.UpdateMetricsRequest_builder{}.Build()
	_, err := client.UpdateMetrics(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateMetrics empty: %v", err)
	}
}

func TestUpdateMetrics_SubnetAllowed(t *testing.T) {
	_, subnet, _ := net.ParseCIDR("192.168.1.0/24")
	client, stop := newTestServer(t, subnet)
	defer stop()

	mtype := pb.Metric_GAUGE
	val := 1.0
	req := pb.UpdateMetricsRequest_builder{
		Metrics: []*pb.Metric{
			pb.Metric_builder{Id: strPtr("x"), Type: &mtype, Value: &val}.Build(),
		},
	}.Build()

	_, err := client.UpdateMetrics(ctxWithIP("192.168.1.10"), req)
	if err != nil {
		t.Fatalf("expected success for trusted IP: %v", err)
	}
}

func TestUpdateMetrics_SubnetDenied(t *testing.T) {
	_, subnet, _ := net.ParseCIDR("192.168.1.0/24")
	client, stop := newTestServer(t, subnet)
	defer stop()

	mtype := pb.Metric_GAUGE
	val := 1.0
	req := pb.UpdateMetricsRequest_builder{
		Metrics: []*pb.Metric{
			pb.Metric_builder{Id: strPtr("x"), Type: &mtype, Value: &val}.Build(),
		},
	}.Build()

	_, err := client.UpdateMetrics(ctxWithIP("10.0.0.1"), req)
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied for untrusted IP, got: %v", err)
	}
}

func TestUpdateMetrics_SubnetMissingIP(t *testing.T) {
	_, subnet, _ := net.ParseCIDR("192.168.1.0/24")
	client, stop := newTestServer(t, subnet)
	defer stop()

	mtype := pb.Metric_GAUGE
	val := 1.0
	req := pb.UpdateMetricsRequest_builder{
		Metrics: []*pb.Metric{
			pb.Metric_builder{Id: strPtr("x"), Type: &mtype, Value: &val}.Build(),
		},
	}.Build()

	// Запрос без метаданных.
	_, err := client.UpdateMetrics(context.Background(), req)
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied without IP metadata, got: %v", err)
	}
}

func strPtr(s string) *string { return &s }
