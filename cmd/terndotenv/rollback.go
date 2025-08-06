package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/tern/v2/migrate"
	"github.com/spf13/cobra"
	"log"
	"os"
)

// rollbackCmd representa o comando rollback
var rollbackCmd = &cobra.Command{
	Use:   "rollback [opções]",
	Short: "Reverte a última migração (tern rollback)",
	Long: `Reverte a migração mais recente aplicada ao banco de dados.
Você pode passar argumentos adicionais que serão repassados para o tern.
Exemplo: dbtool rollback`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		conn, err := pgx.Connect(ctx, fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
			os.Getenv("GOBID_DATABASE_USER"),
			os.Getenv("GOBID_DATABASE_PASSWORD"),
			os.Getenv("GOBID_DATABASE_NAME"),
			os.Getenv("GOBID_DATABASE_HOST"),
			os.Getenv("GOBID_DATABASE_PORT"),
		))

		if err != nil {
			log.Printf("Could not connect to database: %v", err)
			os.Exit(1)
		}

		m, err := migrate.NewMigrator(ctx, conn, "schema_version")
		if err != nil {
			log.Printf("Não foi possível criar o migrador: %v", err)
			os.Exit(1)

		}

		migrationsFS := os.DirFS(migrationsPath)

		if err := m.LoadMigrations(migrationsFS); err != nil {
			log.Printf("Não foi possível carregar os arquivos de migração: %v", err)
			os.Exit(1)

		}
		version, err := m.GetCurrentVersion(ctx)
		if err != nil {
			os.Exit(1)
		}

		if err := m.MigrateTo(ctx, version-1); err != nil {
			log.Println(err)
			os.Exit(1)
		}
	},
}
