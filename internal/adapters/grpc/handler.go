package grpc_adapter

import (
	"context"
	"io"
	"service-b-antifraud/internal/core/domain"
	"service-b-antifraud/internal/usecase"
	"service-b-antifraud/proto/pb"
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

// Unary
func (h *AntiFraudGrpcHandler) AnalyzeRiskScore(ctx context.Context, req *pb.RiskRequest) (*pb.RiskResponse, error) {
	tx := domain.Transaction{
		TransactionID: req.TransactionId,
		UserID:        req.UserId,
		Amount:        req.Amount,
		IPAddress:     req.IpAddress,
	}

	result := h.appService.EvaluateRisk(tx)

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
		// Vamos supor que o req contenha o IP ou você o recupere.
		// Para simplificar o exemplo, vamos bloquear um IP baseado na transação.
		// (Ajuste o seu protobuf para o ChargebackRequest receber o IP se necessário)
		h.appService.BlockIP(req.IpAddress)
		processed++
	}
}
