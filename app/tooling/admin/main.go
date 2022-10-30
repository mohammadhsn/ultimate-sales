package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/mohammadhsn/ultimate-service/business/data/schema"
	"github.com/mohammadhsn/ultimate-service/business/sys/database"
)

var dbCfg = database.Config{
	User:        "postgres",
	Password:    "postgres",
	Host:        "0.0.0.0",
	Name:        "postgres",
	MaxIdleCons: 0,
	MaxOpenCons: 0,
	DisableTLS:  true,
}

func main() {
	err := migrate()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func seed() error {
	db, err := database.Open(dbCfg)

	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := schema.Seed(ctx, db); err != nil {
		return fmt.Errorf("seed database: %w", err)
	}

	fmt.Println("seed complete")

	return nil
}

func migrate() error {
	db, err := database.Open(dbCfg)

	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := schema.Migrate(ctx, db); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	fmt.Println("migrations complete")

	return seed()
}

// genKey creates an x509 private/public key for auth tokens.
func genKey() error {
	// Generate a new private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Create a file for the private key information in PEM form.
	privateFile, err := os.Create("private.pem")
	if err != nil {
		return fmt.Errorf("creating private file: %w", err)
	}
	defer privateFile.Close()

	// Construct a PEM block for the private key.
	privateBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Write the private key to the private key file.
	if err := pem.Encode(privateFile, &privateBlock); err != nil {
		return fmt.Errorf("encoding to private file: %w", err)
	}

	// Marshal the public key from the private key to PKIX.
	ans1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshaling public key: %w", err)
	}

	publicFile, err := os.Create("public.pem")

	if err != nil {
		return fmt.Errorf("creating public file: %w", err)
	}
	defer publicFile.Close()

	publicBlock := pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: ans1Bytes,
	}

	if err := pem.Encode(publicFile, &publicBlock); err != nil {
		return fmt.Errorf("encoding to public file %w", err)
	}

	fmt.Println("private and public key files generated")

	return nil
}
