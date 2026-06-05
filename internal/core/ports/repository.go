package ports

import "service-b-antifraud/internal/core/domain"

// FraudRepository define o contrato para persistência de transações e gerenciamento de blocklist.
type FraudRepository interface {
	// SaveTransaction persiste uma transação no histórico do usuário.
	SaveTransaction(tx domain.Transaction)
	// GetHistoryCountByUser retorna a quantidade de transações prévias de um usuário.
	GetHistoryCountByUser(userID string) int
	// AddToBlocklist insere um IP na lista negra após detecção de fraude.
	AddToBlocklist(ipAddress string)
	// IsIPBlocked verifica se um IP está bloqueado.
	IsIPBlocked(ipAddress string) bool
}
