package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"service-b-antifraud/internal/core/domain"
	"sync"
)

// BackupData representa a estrutura de persistência serializada para disco em formato JSON,
// permitindo restauração do estado de memória após reinicializações do serviço.
type BackupData struct {
	Blocklist    map[string]bool                 `json:"blocklist"`
	Transactions map[string][]domain.Transaction `json:"transactions"`
}

// InMemoryFraudRepository implementa o padrão de repositório com armazenamento em memória,
// utilizando sync.RWMutex para garantir segurança de concorrência em acessos simultâneos,
// e persiste o estado em disco via JSON para recuperação após falhas.
type InMemoryFraudRepository struct {
	mu           sync.RWMutex
	transactions map[string][]domain.Transaction
	blocklist    map[string]bool
	backupFile   string
}

func NewInMemoryFraudRepository(filepath string) *InMemoryFraudRepository {
	repo := &InMemoryFraudRepository{
		transactions: make(map[string][]domain.Transaction),
		blocklist:    make(map[string]bool),
		backupFile:   filepath,
	}
	repo.loadFromDisk()
	return repo
}

// loadFromDisk restaura o estado persistido do repositório a partir do arquivo backup.json.
// Executado durante a inicialização para recuperar dados de transações anteriores e IPs bloqueados.
func (r *InMemoryFraudRepository) loadFromDisk() {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.backupFile)
	if err != nil {
		fmt.Println("⚠️ Nenhum backup encontrado. Iniciando banco vazio.")
		return
	}

	var backup BackupData
	if err := json.Unmarshal(data, &backup); err == nil {
		r.blocklist = backup.Blocklist
		r.transactions = backup.Transactions
		fmt.Println("💾 Banco de dados restaurado com sucesso a partir de:", r.backupFile)
	}
}

// saveToDisk persiste o estado em memória para o arquivo backup.json.
// Deve ser invocado dentro de seções críticas (com lock adquirido) para garantir consistência.
func (r *InMemoryFraudRepository) saveToDisk() {
	backup := BackupData{
		Blocklist:    r.blocklist,
		Transactions: r.transactions,
	}
	data, _ := json.MarshalIndent(backup, "", "  ")
	os.WriteFile(r.backupFile, data, 0644)
}

// SaveTransaction persiste uma transação e sincroniza o estado com o arquivo de backup.
func (r *InMemoryFraudRepository) SaveTransaction(tx domain.Transaction) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.transactions[tx.UserID] = append(r.transactions[tx.UserID], tx)
	r.saveToDisk()
}

// GetHistoryCountByUser retorna o número de transações prévias associadas a um usuário.
func (r *InMemoryFraudRepository) GetHistoryCountByUser(userID string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.transactions[userID])
}

// AddToBlocklist insere um endereço IP na lista de bloqueio após detecção de fraude ou contracarga.
func (r *InMemoryFraudRepository) AddToBlocklist(ipAddress string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.blocklist[ipAddress] = true
	r.saveToDisk()
}

// IsIPBlocked verifica se um endereço IP está presente na lista de bloqueio.
func (r *InMemoryFraudRepository) IsIPBlocked(ipAddress string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.blocklist[ipAddress]
}
