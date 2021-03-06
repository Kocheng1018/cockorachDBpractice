package cockroachdb

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/cockroachdb/cockroach-go/crdb"
	_ "github.com/lib/pq"
)

func transferFunds(tx *sql.Tx, from int, to int, amount int) error {
	// Read the balance.
	var fromBalance int
	if err := tx.QueryRow(
		"SELECT balance FROM accounts WHERE id = $1", from).Scan(&fromBalance); err != nil {
		return err
	}

	if fromBalance < amount {
		return fmt.Errorf("insufficient funds")
	}

	// Perform the transfer.
	if _, err := tx.Exec(
		"UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, from); err != nil {
		return err
	}
	if _, err := tx.Exec(
		"UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, to); err != nil {
		return err
	}
	return nil
}

// 連線
func Connect() (*sql.DB, error) {
	db, err := sql.Open("postgres", "postgres://docker:@localhost:26257/bank?sslmode=disable")
	return db, err
}

func Action() {
	db, err := sql.Open("postgres", "postgres://docker:@localhost:26257/bank?sslmode=disable")
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}

	defer db.Close()

	// Create the "accounts" table.
	// if _, err := db.Exec(
	//     "CREATE TABLE IF NOT EXISTS accounts (id INT PRIMARY KEY, balance INT)"); err != nil {
	//     log.Fatal(err)
	// }

	// Insert two rows into the "accounts" table.
	// if _, err := db.Exec(
	//     "INSERT INTO accounts (id, balance) VALUES (3, 1000), (2, 250)"); err != nil {
	//     log.Fatal(err)
	// }

	// Print out the balances before an account transfer (below).
	PrintBalances(db)

	// Run a transfer in a transaction.
	err = crdb.ExecuteTx(context.Background(), db, nil, func(tx *sql.Tx) error {
		return transferFunds(tx, 1 /* from acct# */, 2 /* to acct# */, 100 /* amount */)
	})
	if err == nil {
		fmt.Println("Success")
	} else {
		log.Fatal("error: ", err)
	}

	// Print out the balances after an account transfer.
	PrintBalances(db)
}
