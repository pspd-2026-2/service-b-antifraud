package ports

import "service-b-antifraud/internal/core/domain"

// RiskUseCase define o contrato de negócio para análise de risco de transações,
type RiskUseCase interface {
	// EvaluateRisk realiza a análise de fraude em uma transação.
	EvaluateRisk(tx domain.Transaction) domain.RiskResult
}
