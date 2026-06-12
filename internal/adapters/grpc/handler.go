package grpc_adapter

import (
	"context"
	"io"
	"log"
	"service-b-antifraud/internal/core/domain"
	"service-b-antifraud/internal/usecase"
	"service-b-antifraud/proto/pb"
	"time"
)

// AntiFraudGrpcHandler implementa os serviços gRPC definidos no proto, adaptando a lógica de negócio
// para o protocolo RPC: suporta Unary RPC para análise individual e Client Streaming para atualização em lote.
type AntiFraudGrpcHandler struct {
	pb.UnimplementedAntiFraudServiceServer
	appService *usecase.RiskService
}

// NewAntiFraudGrpcHandler instancia um novo handler gRPC com a dependência de RiskService injetada.
func NewAntiFraudGrpcHandler(app *usecase.RiskService) *AntiFraudGrpcHandler {
	return &AntiFraudGrpcHandler{appService: app}
}

func riskLevel(status string) string {
	switch status {
	case "blocked":
		return "HIGH"
	case "review":
		return "MEDIUM"
	default:
		return "LOW"
	}
}

// Unary
func (h *AntiFraudGrpcHandler) AnalyzeRiskScore(ctx context.Context, req *pb.RiskRequest) (*pb.RiskResponse, error) {
	tx := domain.Transaction{
		TransactionID: req.TransactionId,
		UserID:        req.UserId,
		Amount:        req.Amount,
		IPAddress:     req.IpAddress,
	}

	start := time.Now()
	result := h.appService.EvaluateRisk(tx)

	log.Printf("[ANALYZE] gRPC AnalyzeRiskScore transactionId=%s status=%s risk=%s riskScore=%d duration=%s",
		result.TransactionID, result.Status, riskLevel(result.Status), result.RiskScore, time.Since(start))

	return &pb.RiskResponse{
		TransactionId: result.TransactionID,
		Status:        result.Status,
		RiskScore:     int32(result.RiskScore),
		Reason:        result.Reason,
		SecurityHash:  result.SecurityHash,
	}, nil
}

func (h *AntiFraudGrpcHandler) BatchChargeback(stream pb.AntiFraudService_BatchChargebackServer) error {
	var processed int32 = 0

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			// Streaming finalizado pelo cliente
			return stream.SendAndClose(&pb.ChargebackSummary{
				TotalProcessed: processed,
				Message:        "Batch applied to in-memory blocklist.",
			})
		}
		if err != nil {
			return err
		}

		// LOGICA REAL: O estorno contém o IP que fraudou.
		h.appService.BlockIP(req.IpAddress)
		processed++
	}
}
