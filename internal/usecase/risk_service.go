package usecase

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"service-b-antifraud/internal/core/domain"
	"service-b-antifraud/internal/core/ports"
	"strings"
)

// RiskService encapsula a lógica de negócio para análise e prevenção de fraudes em transações.
type RiskService struct {
	repo ports.FraudRepository
}

// NewRiskService instancia um novo RiskService com a dependência do repositório injetada.
func NewRiskService(repo ports.FraudRepository) *RiskService {
	return &RiskService{repo: repo}
}

// EvaluateRisk analisa uma transação aplicando regras de negócio, histórico de usuário e validação criptográfica
// para determinar se deve ser aprovada, bloqueada ou submetida à revisão manual.
func (s *RiskService) EvaluateRisk(tx domain.Transaction) domain.RiskResult {
	score := 0
	var reasonParts []string

	// Consulta o repositório para verificar IPs bloqueados por chargebacks anteriores
	if s.repo.IsIPBlocked(tx.IPAddress) {
		score += 100
		reasonParts = append(reasonParts, "IP is explicitly blocked by previous chargebacks")
	}

	// Análise de velocidade: múltiplas transações em curto período indicam potencial fraude
	historyCount := s.repo.GetHistoryCountByUser(tx.UserID)
	if historyCount > 5 {
		score += 40
		reasonParts = append(reasonParts, "High velocity: Too many transactions recently")
	}

	// Regra de negócio: transações de alto valor elevam o risco
	if tx.Amount > 5000 {
		score += 30
		reasonParts = append(reasonParts, "Unusually high transaction amount")
	}

	// Gera assinatura digital via Key Stretching: iteração intensiva de SHA256 (50.000 rounds)
	// cria gargalo intencional de CPU, simulando Proof of Work para validação criptográfica.
	// Este mecanismo fornece: (1) evidência de computação para auditoria, (2) teste de multiplexação gRPC,
	// (3) fator de dificuldade contra ataques de força bruta.
	baseString := fmt.Sprintf("%s-%s-%f", tx.TransactionID, tx.UserID, tx.Amount)
	currentHash := sha256.Sum256([]byte(baseString))

	for i := 0; i < 50000; i++ {
		currentHash = sha256.Sum256(currentHash[:])
	}

	securityToken := hex.EncodeToString(currentHash[:])

	// Classifica o resultado conforme a pontuação acumulada
	status := "approved"
	finalReason := "Low risk"

	if score > 80 {
		status = "blocked"
		finalReason = "Fraud prevented: " + strings.Join(reasonParts, ", ")
	} else if score > 40 {
		status = "review"
		finalReason = "Manual review required: " + strings.Join(reasonParts, ", ")
	} else if len(reasonParts) > 0 {
		finalReason = "Approved with warnings: " + strings.Join(reasonParts, ", ")
	}

	// Persiste a transação para atualizar o histórico do usuário em futuras análises
	s.repo.SaveTransaction(tx)

	return domain.RiskResult{
		TransactionID: tx.TransactionID,
		Status:        status,
		RiskScore:     score,
		Reason:        finalReason,
		SecurityHash:  securityToken,
	}
}

// BlockIP insere um endereço IP na lista negra após confirmação de fraude.
// Validações subsequentes rejeitarão transações originadas deste IP.
func (s *RiskService) BlockIP(ipAddress string) {
	if ipAddress != "" {
		s.repo.AddToBlocklist(ipAddress)
	}
}
