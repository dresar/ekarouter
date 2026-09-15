package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "data/ekarouter.db")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// 1. Update route items pointing to decommissioned groq models
	_, err = db.Exec(`
		UPDATE route_items
		SET model_id = 'groq/openai/gpt-oss-120b'
		WHERE model_id = 'groq/llama-3.3-70b-versatile' OR model_id = 'groq/llama-3.1-8b-instant'`)
	if err != nil {
		fmt.Printf("Route items update err: %v\n", err)
	} else {
		fmt.Println("Updated route items to groq/openai/gpt-oss-120b")
	}

	// 2. Remove decommissioned groq models from models table
	_, _ = db.Exec(`
		DELETE FROM models
		WHERE provider_id = 'groq'
		  AND external_name IN ('llama-3.3-70b-versatile', 'llama-3.1-8b-instant', 'deepseek-r1-distill-llama-70b', 'qwen-qwq-32b', 'mistral-saba-24b')`)

	// 3. Insert live 2026 groq models
	type NewModel struct {
		ExtName string
		Disp    string
		Enabled int
	}
	newModels := []NewModel{
		{"openai/gpt-oss-120b", "GPT-OSS 120B (Flagship)", 1},
		{"openai/gpt-oss-20b", "GPT-OSS 20B (Fast)", 1},
		{"qwen/qwen3.8-27b", "Qwen 3.8 27B", 1},
		{"groq/compound", "Groq Compound (Agentic)", 1},
		{"groq/compound-mini", "Groq Compound Mini", 1},
		{"allam-2-7b", "Allam 2 7B", 1},
		{"openai/gpt-oss-safeguard-20b", "GPT-OSS Safeguard 20B", 0},
		{"canopylabs/orpheus-v1-english", "Orpheus v1 English (Audio)", 0},
		{"canopylabs/orpheus-arabic-saudi", "Orpheus Arabic Saudi (Audio)", 0},
		{"whisper-large-v3-turbo", "Whisper Large v3 Turbo", 0},
		{"whisper-large-v3", "Whisper Large v3", 0},
		{"meta-llama/llama-prompt-guard-2-86m", "Llama Prompt Guard 2 86M", 0},
		{"meta-llama/llama-prompt-guard-2-22m", "Llama Prompt Guard 2 22M", 0},
	}

	for _, nm := range newModels {
		fullID := "groq/" + nm.ExtName
		_, err = db.Exec(`
			INSERT INTO models (id, provider_id, external_name, display_name, enabled)
			VALUES (?, 'groq', ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				display_name = excluded.display_name,
				enabled = excluded.enabled`,
			fullID, nm.ExtName, nm.Disp, nm.Enabled)
		if err != nil {
			fmt.Printf("Error inserting %s: %v\n", fullID, err)
		}
	}
	fmt.Println("Successfully seeded 2026 Groq models into database!")

	// 4. Verify
	mRows, err := db.Query("SELECT id, external_name, display_name, enabled FROM models WHERE provider_id = 'groq'")
	if err == nil {
		fmt.Println("\n=== ACTIVE GROQ MODELS IN DB ===")
		for mRows.Next() {
			var mid, extName, dispName string
			var enabled int
			_ = mRows.Scan(&mid, &extName, &dispName, &enabled)
			fmt.Printf("id: %-36s | ext: %-26s | enabled: %d\n", mid, extName, enabled)
		}
		mRows.Close()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
