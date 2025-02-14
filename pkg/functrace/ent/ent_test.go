package ent

import (
	"context"
	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"testing"
)

/**
@file:
@author: levi.Tang
@time: 2024/11/4 15:33
@description:
**/

func TestSqlite3(t *testing.T) {
	client, err := Open(dialect.SQLite, "./trace.db?_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	_, err = client.Trace.Create().SetFuncName("test").SetIndent(1).SetGid(1).SetParams("").Save(context.Background())
	if err != nil {
		t.Error(err)
	}
}
