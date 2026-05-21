package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bpcl/etl/internal/config"
	"github.com/bpcl/etl/internal/etl"
	"github.com/bpcl/etl/internal/microsoft"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/robfig/cron/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log.Println("ETL Service Starting...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("ETL Config: FilePath=%s, Enabled=%v, Cron=%s",
		cfg.MicrosoftFilePath, cfg.ETLEnabled, cfg.ETLCron)

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DBUrl)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("Connected to database")

	// Create Azure AD credential using client secret
	credential, err := azidentity.NewClientSecretCredential(
		cfg.MicrosoftTenantID,
		cfg.MicrosoftClientID,
		cfg.MicrosoftClientSecret,
		&azidentity.ClientSecretCredentialOptions{},
	)
	if err != nil {
		log.Fatalf("Failed to create Azure AD credential: %v", err)
	}

	// Create Microsoft Graph client for Excel file access
	microsoftClient, err := microsoft.NewClient(ctx, credential, cfg.MicrosoftFilePath)
	if err != nil {
		log.Fatalf("Failed to create Microsoft Graph client: %v", err)
	}

	log.Println("Connected to Microsoft Graph")

	extractor := etl.NewExtractor(microsoftClient, "TARGETS", 3, 500)
	transformer := etl.NewTransformer(pool, cfg.TargetPeriodMonth)
	loader := etl.NewLoader(pool)

	runETL := func() {
		log.Println("Starting ETL run...")
		start := time.Now()

		rawRows, err := extractor.Extract(ctx)
		if err != nil {
			errMsg := fmt.Sprintf("Extract failed: %v", err)
			log.Println(errMsg)
			_ = loader.LogETLRun(ctx, "error", errMsg, 0)
			return
		}
		log.Printf("Extracted %d rows", len(rawRows))

		records, err := transformer.Transform(ctx, rawRows)
		if err != nil {
			errMsg := fmt.Sprintf("Transform failed: %v", err)
			log.Println(errMsg)
			_ = loader.LogETLRun(ctx, "error", errMsg, 0)
			return
		}
		log.Printf("Transformed %d target records", len(records))

		result, err := loader.LoadTargets(ctx, records)
		if err != nil {
			errMsg := fmt.Sprintf("Load failed: %v", err)
			log.Println(errMsg)
			_ = loader.LogETLRun(ctx, "error", errMsg, 0)
			return
		}

		durationMs := int(time.Since(start).Milliseconds())
		detail := fmt.Sprintf("Inserted=%d, Skipped=%d", result.Inserted, result.Skipped)
		log.Printf("ETL completed: %s (took %dms)", detail, durationMs)

		_ = loader.LogETLRun(ctx, "ok", detail, durationMs)

		if len(result.Errors) > 0 {
			log.Printf("Errors: %v", result.Errors)
		}
	}

	if cfg.ETLEnabled && cfg.ETLCron != "" {
		c := cron.New()
		_, err := c.AddFunc(cfg.ETLCron, runETL)
		if err != nil {
			log.Fatalf("Failed to add cron job: %v", err)
		}
		c.Start()
		log.Printf("Cron scheduled: %s", cfg.ETLCron)
	} else {
		log.Println("Cron disabled - running once")
		runETL()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
}

func RunETLOnce(cfg *config.Config, pool *pgxpool.Pool, extractor *etl.Extractor, transformer *etl.Transformer, loader *etl.Loader) error {
	ctx := context.Background()

	log.Println("Starting manual ETL run...")
	start := time.Now()

	rawRows, err := extractor.Extract(ctx)
	if err != nil {
		return fmt.Errorf("extract failed: %w", err)
	}

	records, err := transformer.Transform(ctx, rawRows)
	if err != nil {
		return fmt.Errorf("transform failed: %w", err)
	}

	result, err := loader.LoadTargets(ctx, records)
	if err != nil {
		return fmt.Errorf("load failed: %w", err)
	}

	durationMs := int(time.Since(start).Milliseconds())
	detail := fmt.Sprintf("Inserted=%d, Skipped=%d", result.Inserted, result.Skipped)

	_ = loader.LogETLRun(ctx, "ok", detail, durationMs)

	return nil
}