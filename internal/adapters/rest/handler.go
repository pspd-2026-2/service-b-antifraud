package rest

import (
	"encoding/json"
	"net/http"
	"service-b-antifraud/internal/core/domain"
	"service-b-antifraud/internal/usecase"
)

// RestHandler implementa os endpoints HTTP para análise de risco, adaptando requisições JSON para a lógica de negócio.
type RestHandler struct {
	appService *usecase.RiskService
}

func NewRestHandler(app *usecase.RiskService) *RestHandler {
	return &RestHandler{appService: app}
}

// AnalyzeRiskHandler processa requisições POST no endpoint /api/v1/risk/analyze,
// desserializa a transação JSON e retorna a análise de risco em formato JSON.
func (h *RestHandler) AnalyzeRiskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.Transaction
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := h.appService.EvaluateRisk(req)

	w.Header().Set("Content-Type", "usecase/json")
	json.NewEncoder(w).Encode(result)
}
