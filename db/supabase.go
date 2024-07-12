package db

import (
	"fmt"
	"log"
	"os"

	"github.com/supabase-community/supabase-go"
)

func NewSupabase() *supabase.Client {
	client, err := supabase.NewClient(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_KEY"), nil)
	if err != nil {
		log.Fatal(fmt.Errorf("error connecting to supabase: %w", err))
	}

	return client
}
