package functrace

import (
	"context"
	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	"github.com/toheart/goanalysis/pkg/functrace/ent"
)

/*
*
@file:
@author: levi.Tang
@time: 2024/11/4 17:23
@description:
*
*/
type Data struct {
	db *ent.Client
}

// NewData .
func NewData() (*Data, func(), error) {
	client, err := ent.Open(dialect.SQLite, "./trace.db?_fk=1")
	if err != nil {
		log.Error("failed opening connection to sqlite: %v", err)
		return nil, nil, err
	}
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Error("failed creating schema resources: %v", err)
		return nil, nil, err
	}

	d := &Data{
		db: client,
	}
	return d, func() {
		log.Info("message", "closing the data resources")
		if err := d.db.Close(); err != nil {
			log.Error("%s", err)
		}

	}, nil
}
