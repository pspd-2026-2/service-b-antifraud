package domain

// Transaction representa os dados de transação financeira independentemente do protocolo de transporte (JSON ou Protobuf).
type Transaction struct {
	TransactionID string  `json:"transactionId"`
	UserID        string  `json:"userId"`
	Amount        float64 `json:"amount"`
	IPAddress     string  `json:"ipAddress"`
}

// RiskResult encapsula a decisão de análise de risco com score, classificação e evidência criptográfica.
// Os valores de Status são: "approved", "blocked" ou "review".
type RiskResult struct {
	TransactionID string `json:"transactionId"`
	Status        string `json:"status"`
	RiskScore     int    `json:"riskScore"`
	Reason        string `json:"reason"`
	SecurityHash  string `json:"securityHash"`
}
