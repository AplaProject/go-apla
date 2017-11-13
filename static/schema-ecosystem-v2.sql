DROP TABLE IF EXISTS "%[1]d_keys"; CREATE TABLE "%[1]d_keys" (
"id" bigint  NOT NULL DEFAULT '0',
"pub" bytea  NOT NULL DEFAULT '',
"amount" decimal(30) NOT NULL DEFAULT '0',
"rb_id" bigint NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "%[1]d_keys" ADD CONSTRAINT "%[1]d_keys_pkey" PRIMARY KEY (id);

DROP TABLE IF EXISTS "%[1]d_history"; CREATE TABLE "%[1]d_history" (
"id" bigint NOT NULL  DEFAULT '0',
"sender_id" bigint NOT NULL DEFAULT '0',
"recipient_id" bigint NOT NULL DEFAULT '0',
"amount" decimal(30) NOT NULL DEFAULT '0',
"comment" text NOT NULL DEFAULT '',
"block_id" int  NOT NULL DEFAULT '0',
"txhash" bytea  NOT NULL DEFAULT '',
"rb_id" int  NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "%[1]d_history" ADD CONSTRAINT "%[1]d_history_pkey" PRIMARY KEY (id);
CREATE INDEX "%[1]d_history_index_sender" ON "%[1]d_history" (sender_id);
CREATE INDEX "%[1]d_history_index_recipient" ON "%[1]d_history" (recipient_id);
CREATE INDEX "%[1]d_history_index_block" ON "%[1]d_history" (block_id, txhash);


DROP TABLE IF EXISTS "%[1]d_languages"; CREATE TABLE "%[1]d_languages" (
  "id" bigint  NOT NULL DEFAULT '0',
  "name" character varying(100) NOT NULL DEFAULT '',
  "res" text NOT NULL DEFAULT '',
  "conditions" text NOT NULL DEFAULT '',
  "rb_id" bigint NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "%[1]d_languages" ADD CONSTRAINT "%[1]d_languages_pkey" PRIMARY KEY (id);
CREATE INDEX "%[1]d_languages_index_name" ON "%[1]d_languages" (name);

DROP TABLE IF EXISTS "%[1]d_menu"; CREATE TABLE "%[1]d_menu" (
    "id" bigint  NOT NULL DEFAULT '0',
    "name" character varying(255) UNIQUE NOT NULL DEFAULT '',
    "title" character varying(255) NOT NULL DEFAULT '',
    "value" text NOT NULL DEFAULT '',
    "conditions" text NOT NULL DEFAULT '',
    "rb_id" bigint NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "%[1]d_menu" ADD CONSTRAINT "%[1]d_menu_pkey" PRIMARY KEY (id);
CREATE INDEX "%[1]d_menu_index_name" ON "%[1]d_menu" (name);

DROP TABLE IF EXISTS "%[1]d_pages"; CREATE TABLE "%[1]d_pages" (
    "id" bigint  NOT NULL DEFAULT '0',
    "name" character varying(255) UNIQUE NOT NULL DEFAULT '',
    "value" text NOT NULL DEFAULT '',
    "menu" character varying(255) NOT NULL DEFAULT '',
    "conditions" text NOT NULL DEFAULT '',
    "rb_id" bigint NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "%[1]d_pages" ADD CONSTRAINT "%[1]d_pages_pkey" PRIMARY KEY (id);
CREATE INDEX "%[1]d_pages_index_name" ON "%[1]d_pages" (name);

DROP TABLE IF EXISTS "%[1]d_blocks"; CREATE TABLE "%[1]d_blocks" (
    "id" bigint  NOT NULL DEFAULT '0',
    "name" character varying(255) UNIQUE NOT NULL DEFAULT '',
    "value" text NOT NULL DEFAULT '',
    "conditions" text NOT NULL DEFAULT '',
    "rb_id" bigint NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "%[1]d_blocks" ADD CONSTRAINT "%[1]d_blocks_pkey" PRIMARY KEY (id);
CREATE INDEX "%[1]d_blocks_index_name" ON "%[1]d_blocks" (name);

DROP TABLE IF EXISTS "%[1]d_signatures"; CREATE TABLE "%[1]d_signatures" (
    "id" bigint  NOT NULL DEFAULT '0',
    "name" character varying(100) NOT NULL DEFAULT '',
    "value" jsonb,
    "conditions" text NOT NULL DEFAULT '',
    "rb_id" bigint NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "%[1]d_signatures" ADD CONSTRAINT "%[1]d_signatures_pkey" PRIMARY KEY (name);

CREATE TABLE "%[1]d_contracts" (
"id" bigint NOT NULL  DEFAULT '0',
"value" text  NOT NULL DEFAULT '',
"wallet_id" bigint NOT NULL DEFAULT '0',
"token_id" bigint NOT NULL DEFAULT '1',
"active" character(1) NOT NULL DEFAULT '0',
"conditions" text  NOT NULL DEFAULT '',
"rb_id" bigint NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "%[1]d_contracts" ADD CONSTRAINT "%[1]d_contracts_pkey" PRIMARY KEY (id);

INSERT INTO "%[1]d_contracts" ("id", "value", "wallet_id","active", "conditions") VALUES 
('1','contract MainCondition {
  conditions {
    if(StateVal("founder_account")!=$key_id)
    {
      warning "Sorry, you don`t have access to this action."
    }
  }
}', '%[2]d', '0', 'ContractConditions(`MainCondition`)');

DROP TABLE IF EXISTS "%[1]d_parameters";
CREATE TABLE "%[1]d_parameters" (
"id" bigint NOT NULL  DEFAULT '0',
"name" varchar(255) UNIQUE NOT NULL DEFAULT '',
"value" text NOT NULL DEFAULT '',
"conditions" text  NOT NULL DEFAULT '',
"rb_id" bigint  NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "%[1]d_parameters" ADD CONSTRAINT "%[1]d_parameters_pkey" PRIMARY KEY ("id");
CREATE INDEX "%[1]d_parameters_index_name" ON "%[1]d_parameters" (name);

INSERT INTO "%[1]d_parameters" ("id","name", "value", "conditions") VALUES 
('1','founder_account', '%[2]d', 'ContractConditions(`MainCondition`)'),
('2','full_node_wallet_id', '%[2]d', 'ContractConditions(`MainCondition`)'),
('3','host', '', 'ContractConditions(`MainCondition`)'),
('4','new_table', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('5','new_column', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('6','changing_tables', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('7','changing_language', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('8','changing_signature', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('9','changing_page', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('10','changing_menu', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('11','changing_contracts', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('12','ecosystem_name', '%[3]s', 'ContractConditions(`MainCondition`)'),
('13','max_sum', '1000000', 'ContractConditions(`MainCondition`)'),
('14','citizenship_cost', '1', 'ContractConditions(`MainCondition`)'),
('15','money_digit', '2', 'ContractConditions(`MainCondition`)'),
('16','stylesheet', 'body { 
  /* You can define your custom styles here or create custom CSS rules */
}', 'ContractConditions(`MainCondition`)');

CREATE TABLE "%[1]d_tables" (
"id" bigint NOT NULL  DEFAULT '0',
"name" varchar(100) UNIQUE NOT NULL DEFAULT '',
"permissions" jsonb,
"columns" jsonb,
"conditions" text  NOT NULL DEFAULT '',
"rb_id" bigint NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "%[1]d_tables" ADD CONSTRAINT "%[1]d_tables_pkey" PRIMARY KEY ("id");
CREATE INDEX "%[1]d_tables_index_name" ON "%[1]d_tables" (name);

INSERT INTO "%[1]d_tables" ("id", "name", "permissions","columns", "conditions") VALUES ('1', 'contracts', 
        '{"insert": "ContractAccess(\"@1NewContract\")", "update": "ContractAccess(\"@1EditContract\")", 
          "new_column": "ContractAccess(\"@1NewColumn\")"}',
        '{"value": "ContractAccess(\"@1EditContract\", \"@1ActivateContract\")",
          "wallet_id": "ContractAccess(\"@1EditContract\", \"@1ActivateContract\")",
          "token_id": "ContractAccess(\"@1EditContract\", \"@1ActivateContract\")",
          "active": "ContractAccess(\"@1EditContract\", \"@1ActivateContract\")",
          "conditions": "ContractAccess(\"@1EditContract\", \"@1ActivateContract\")"}', 'ContractAccess("@1EditTable")'),
        ('2', 'keys', 
        '{"insert": "ContractAccess(\"@1MoneyTransfer\", \"@1NewEcosystem\")", "update": "ContractAccess(\"@1MoneyTransfer\")", 
          "new_column": "ContractAccess(\"@1NewColumn\")"}',
        '{"pub": "ContractAccess(\"@1MoneyTransfer\")",
          "amount": "ContractAccess(\"@1MoneyTransfer\")"}', 'ContractAccess("@1EditTable")'),
        ('3', 'history', 
        '{"insert": "ContractAccess(\"@1MoneyTransfer\")", "update": "false", 
          "new_column": "false"}',
        '{"sender_id": "ContractAccess(\"@1MoneyTransfer\")",
          "recipient_id": "ContractAccess(\"@1MoneyTransfer\")",
          "amount":  "ContractAccess(\"@1MoneyTransfer\")",
          "comment": "ContractAccess(\"@1MoneyTransfer\")",
          "block_id":  "ContractAccess(\"@1MoneyTransfer\")",
          "txhash": "ContractAccess(\"@1MoneyTransfer\")"}', 'ContractAccess("@1EditTable")'),        
        ('4', 'languages', 
        '{"insert": "ContractAccess(\"@1NewLang\")", "update": "ContractAccess(\"@1EditLang\")", 
          "new_column": "ContractAccess(\"@1NewColumn\")"}',
        '{ "name": "ContractAccess(\"@1EditLang\")",
          "res": "ContractAccess(\"@1EditLang\")",
          "conditions": "ContractAccess(\"@1EditLang\")"}', 'ContractAccess("@1EditTable")'),
        ('5', 'menu', 
        '{"insert": "ContractAccess(\"@1NewMenu\", \"@1NewEcosystem\")", "update": "ContractAccess(\"@1EditMenu\")", 
          "new_column": "ContractAccess(\"@1NewColumn\")"}',
        '{"name": "ContractAccess(\"@1EditMenu\")",
    "value": "ContractAccess(\"@1EditMenu\")",
    "conditions": "ContractAccess(\"@1EditMenu\")"
        }', 'ContractAccess("@1EditTable")'),
        ('6', 'pages', 
        '{"insert": "ContractAccess(\"@1NewPage\", \"@1NewEcosystem\")", "update": "ContractAccess(\"@1EditPage\")", 
          "new_column": "ContractAccess(\"@1NewColumn\")"}',
        '{"name": "ContractAccess(\"@1EditPage\")",
    "value": "ContractAccess(\"@1EditPage\")",
    "menu": "ContractAccess(\"@1EditPage\")",
    "conditions": "ContractAccess(\"@1EditPage\")"
        }', 'ContractAccess("@1EditTable")'),
        ('7', 'blocks', 
        '{"insert": "ContractAccess(\"@1NewBlock\")", "update": "ContractAccess(\"@1EditBlock\")", 
          "new_column": "ContractAccess(\"@1NewColumn\")"}',
        '{"name": "ContractAccess(\"@1EditBlock\")",
    "value": "ContractAccess(\"@1EditBlock\")",
    "conditions": "ContractAccess(\"@1EditBlock\")"
        }', 'ContractAccess("@1EditTable")'),
        ('8', 'signatures', 
        '{"insert": "ContractAccess(\"@1NewSign\")", "update": "ContractAccess(\"@1EditSign\")", 
          "new_column": "ContractAccess(\"@1NewColumn\")"}',
        '{"name": "ContractAccess(\"@1EditSign\")",
    "value": "ContractAccess(\"@1EditSign\")",
    "conditions": "ContractAccess(\"@1EditSign\")"
        }', 'ContractAccess("@1EditTable")');

