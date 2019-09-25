// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package migration

//go:generate go run ./gen/contracts.go

var (
	migrationInitial = `
	{{headseq "migration_history"}}
		t.Column("id", "int", {"default_raw": "nextval('migration_history_id_seq')"})
		t.Column("version", "string", {"default": "", "size":255})
		t.Column("date_applied", "int", {})
	{{footer "seq" "primary"}}
`
	migrationInitialTables = `
	{{head "transactions_status"}}
		t.Column("hash", "bytea", {"default": ""})
		t.Column("time", "int", {"default": "0"})
		t.Column("type", "int", {"default": "0"})
		t.Column("ecosystem", "int", {"default": "1"})
		t.Column("wallet_id", "bigint", {"default": "0"})
		t.Column("block_id", "int", {"default": "0"})
		t.Column("error", "string", {"default": "", "size":255})
	{{footer "primary(hash)"}}

	{{head "confirmations"}}
		t.Column("block_id", "bigint", {"default": "0"})
		t.Column("good", "int", {"default": "0"})
		t.Column("bad", "int", {"default": "0"})
		t.Column("time", "int", {"default": "0"})
	{{footer "primary(block_id)"}}

	{{head "block_chain"}}
		t.Column("id", "int", {"default": "0"})
		t.Column("hash", "bytea", {"default": ""})
		t.Column("rollbacks_hash", "bytea", {"default": ""})
		t.Column("data", "bytea", {"default": ""})
		t.Column("ecosystem_id", "int", {"default": "0"})
		t.Column("key_id", "bigint", {"default": "0"})
		t.Column("node_position", "bigint", {"default": "0"})
		t.Column("time", "int", {"default": "0"})
		t.Column("tx", "int", {"default": "0"})
	{{footer "primary"}}

	{{head "log_transactions"}}
		t.Column("hash", "bytea", {"default": ""})
		t.Column("block", "int", {"default": "0"})
	{{footer "primary(hash)"}}

	{{head "queue_tx"}}
		t.Column("hash", "bytea", {"default": ""})
		t.Column("data", "bytea", {"default": ""})
		t.Column("from_gate", "int", {"default": "0"})
	{{footer "primary(hash)"}}

	{{head "info_block"}}
		t.Column("hash", "bytea", {"default": ""})
		t.Column("rollbacks_hash", "bytea", {"default": ""})
		t.Column("block_id", "int", {"default": "0"})
		t.Column("node_position", "int", {"default": "0"})
		t.Column("ecosystem_id", "bigint", {"default": "0"})
		t.Column("key_id", "bigint", {"default": "0"})
		t.Column("time", "int", {"default": "0"})
		t.Column("current_version", "string", {"default": "0.0.1", "size": 50})
		t.Column("sent", "smallint", {"default": "0"})
	{{footer}}

	{{head "queue_blocks"}}
		t.Column("hash", "bytea", {"default": ""})
		t.Column("full_node_id", "bigint", {"default": "0"})
		t.Column("block_id", "int", {"default": "0"})
	{{footer "primary(hash)"}}

	{{head "transactions"}}
		t.Column("hash", "bytea", {"default": ""})
		t.Column("data", "bytea", {"default": ""})
		t.Column("used", "smallint", {"default": "0"})
		t.Column("high_rate", "smallint", {"default": "0"})
		t.Column("type", "smallint", {"default": "0"})
		t.Column("key_id", "bigint", {"default": "0"})
		t.Column("counter", "smallint", {"default": "0"})
		t.Column("sent", "smallint", {"default": "0"})
		t.Column("attempt", "smallint", {"default": "0"})
		t.Column("verified", "smallint", {"default": "1"})
	{{footer "primary(hash)"}}

	{{headseq "rollback_tx"}}
		t.Column("id", "bigint", {"default_raw": "nextval('rollback_tx_id_seq')"})
		t.Column("block_id", "bigint", {"default": "0"})
		t.Column("tx_hash", "bytea", {"default": ""})
		t.Column("table_name", "string", {"default": "", "size":255})
		t.Column("table_id", "string", {"default": "", "size":255})
		t.Column("data", "text", {"default": ""})
	{{footer "seq" "primary" "index(table_name, table_id)"}}

	{{head "install"}}
		t.Column("progress", "string", {"default": "", "size":10})
	{{footer}}

	sql("DROP TYPE IF EXISTS \"my_node_keys_enum_status\" CASCADE;")
	sql("CREATE TYPE \"my_node_keys_enum_status\" AS ENUM ('my_pending','approved');")

	{{headseq "my_node_keys"}}
		t.Column("id", "int", {"default_raw": "nextval('my_node_keys_id_seq')"})
		t.Column("add_time", "int", {"default": "0"})
		t.Column("public_key", "bytea", {"default": ""})
		t.Column("private_key", "string", {"default": "", "size":3096})
		t.Column("status", "my_node_keys_enum_status", {"default": "my_pending"})
		t.Column("my_time", "int", {"default": "0"})
		t.Column("time", "bigint", {"default": "0"})
		t.Column("block_id", "int", {"default": "0"})
	{{footer "seq" "primary"}}

	{{head "stop_daemons"}}
		t.Column("stop_time", "int", {"default": "0"})
	{{footer}}
`

	migrationInitialSchema = `
		CREATE OR REPLACE FUNCTION next_id(table_name TEXT, OUT result INT) AS
		$$
		BEGIN
			EXECUTE FORMAT('SELECT COUNT(*) + 1 FROM "%s"', table_name)
			INTO result;
			RETURN;
		END
		$$
		LANGUAGE plpgsql;`
)
