package grpcserver

import (
	"context"

	pb "github.com/F3dosik/metalert/internal/proto"
	"github.com/F3dosik/metalert/internal/repository"
	"github.com/F3dosik/metalert/internal/service"
	"github.com/F3dosik/metalert/pkg/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServer
	storage repository.MetricsStorage
}

func NewMetricsServer(s repository.MetricsStorage) *MetricsServer {
	return &MetricsServer{storage: s}
}

func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	metrics := make([]models.Metric, 0, len(req.GetMetrics()))

	for _, m := range req.GetMetrics() {
		metric := models.Metric{ID: m.GetId()}

		switch m.GetType() {
		case pb.Metric_GAUGE:
			metric.MType = "gauge"
			v := models.Gauge(m.GetValue())
			metric.Value = &v
		case pb.Metric_COUNTER:
			metric.MType = "counter"
			d := models.Counter(m.GetDelta())
			metric.Delta = &d
		}

		metrics = append(metrics, metric)
	}

	if err := service.UpdateMetrics(ctx, s.storage, metrics, nil, ""); err != nil {
		return nil, status.Errorf(codes.Internal, "update metrics: %v", err)
	}

	return &pb.UpdateMetricsResponse{}, nil
}
