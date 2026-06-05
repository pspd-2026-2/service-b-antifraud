package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	grpc_adapter "service-b-antifraud/internal/adapters/grpc"
	"service-b-antifraud/internal/adapters/repository"
	rest_adapter "service-b-antifraud/internal/adapters/rest"
	"service-b-antifraud/internal/config" // Importando o novo pacote
	"service-b-antifraud/internal/usecase"
	"service-b-antifraud/proto/pb"

	"google.golang.org/grpc"
)

func main() {
	cfg := config.LoadConfig()

	repo := repository.NewInMemoryFraudRepository(cfg.BackupFile)
	appService := usecase.NewRiskService(repo)

	// Inicia servidor REST em goroutine, permitindo que o servidor gRPC bloqueie a thread principal
	go func() {
		restHandler := rest_adapter.NewRestHandler(appService)
		http.HandleFunc("/api/v1/risk/analyze", restHandler.AnalyzeRiskHandler)
		fmt.Printf("🚀 REST Server rodando na porta %s...\n", cfg.RestPort)
		log.Fatal(http.ListenAndServe(":"+cfg.RestPort, nil))
	}()

	lis, err := net.Listen("tcp", ":"+cfg.GrpcPort)
	if err != nil {
		log.Fatalf("Falha ao escutar porta gRPC: %v", err)
	}

	grpcServer := grpc.NewServer()
	grpcHandler := grpc_adapter.NewAntiFraudGrpcHandler(appService)
	pb.RegisterAntiFraudServiceServer(grpcServer, grpcHandler)

	fmt.Printf("🚀 gRPC Server rodando na porta %s...\n", cfg.GrpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Falha ao servir gRPC: %v", err)
	}
}
