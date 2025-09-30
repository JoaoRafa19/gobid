package main

import (
	"github.com/spf13/cobra"
)

// statusCmd representa o comando status
var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"s"},
	Short:   "Verifica o status das migrações (tern status)",
	Long:    `Exibe o status atual de todas as migrações, mostrando quais foram aplicadas e quando.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Chama a função auxiliar com a ação "status"
		executeTernCommand("status")
	},
}
