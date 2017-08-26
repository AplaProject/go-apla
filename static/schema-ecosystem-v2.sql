DROP TABLE IF EXISTS "%d_keys"; CREATE TABLE "%[1]d_keys" (
"id" bigint  NOT NULL DEFAULT '0',
"pub" bytea  NOT NULL DEFAULT '',
"amount" decimal(30) NOT NULL DEFAULT '0',
"rb_id" bigint NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "%[1]d_keys" ADD CONSTRAINT "%[1]d_keys_pkey" PRIMARY KEY (id);

DROP TABLE IF EXISTS "%d_languages"; CREATE TABLE "%[1]d_languages" (
  "name" character varying(100) NOT NULL DEFAULT '',
  "res" jsonb,
  "conditions" text NOT NULL DEFAULT '',
  "rb_id" bigint NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "%[1]d_languages" ADD CONSTRAINT "%[1]d_languages_pkey" PRIMARY KEY (name);

DROP TABLE IF EXISTS "%d_menu"; CREATE TABLE "%[1]d_menu" (
    "name" character varying(255) NOT NULL DEFAULT '',
    "value" text NOT NULL DEFAULT '',
    "conditions" text NOT NULL DEFAULT '',
    "rb_id" bigint NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "%[1]d_menu" ADD CONSTRAINT "%[1]d_menu_pkey" PRIMARY KEY (name);

DROP TABLE IF EXISTS "%d_pages"; CREATE TABLE "%[1]d_pages" (
    "name" character varying(255) NOT NULL DEFAULT '',
    "value" text NOT NULL DEFAULT '',
    "menu" character varying(255) NOT NULL DEFAULT '',
    "conditions" text NOT NULL DEFAULT '',
    "rb_id" bigint NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "%[1]d_pages" ADD CONSTRAINT "%[1]d_pages_pkey" PRIMARY KEY (name);

DROP TABLE IF EXISTS "%d_signatures"; CREATE TABLE "%[1]d_signatures" (
    "name" character varying(100) NOT NULL DEFAULT '',
    "value" jsonb,
    "conditions" text NOT NULL DEFAULT '',
    "rb_id" bigint NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "%[1]d_signatures" ADD CONSTRAINT "%[1]d_signatures_pkey" PRIMARY KEY (name);

DROP SEQUENCE IF EXISTS "%[1]d_contracts_id_seq" CASCADE;
CREATE SEQUENCE "%[1]d_contracts_id_seq" START WITH 1;
CREATE TABLE "%[1]d_contracts" (
"id" bigint NOT NULL  default nextval('%[1]d_contracts_id_seq'),
"value" text  NOT NULL DEFAULT '',
"wallet_id" bigint NOT NULL DEFAULT '0',
"active" character(1) NOT NULL DEFAULT '0',
"conditions" text  NOT NULL DEFAULT '',
"rb_id" bigint NOT NULL DEFAULT '0'
);
ALTER SEQUENCE "%[1]d_contracts_id_seq" owned by "%[1]d_contracts".id;
ALTER TABLE ONLY "%[1]d_contracts" ADD CONSTRAINT "%[1]d_contracts_pkey" PRIMARY KEY (id);

DROP TABLE IF EXISTS "%[1]d_parameters";
CREATE TABLE "%[1]d_parameters" (
"name" varchar(255)  NOT NULL DEFAULT '',
"value" text NOT NULL DEFAULT '',
"conditions" text  NOT NULL DEFAULT '',
"rb_id" bigint  NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "%[1]d_parameters" ADD CONSTRAINT "%[1]d_parameters_pkey" PRIMARY KEY ("name");

CREATE TABLE "%[1]d_tables" (
"name" varchar(100)  NOT NULL DEFAULT '',
"permissions" jsonb,
"columns" jsonb,
"conditions" text  NOT NULL DEFAULT '',
"rb_id" bigint NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "%[1]d_tables" ADD CONSTRAINT "%[1]d_tables_pkey" PRIMARY KEY (name);

INSERT INTO "%[1]d_tables" ("name", "permissions","columns", "conditions") VALUES ('%[1]d_contracts', 
        '{"insert": "ContractAccess(\"@1NewContract\")", "update": "ContractAccess(\"@1EditContract\")", 
          "new_column": "ContractAccess(\"@1NewColumn\")"}',
        '{}', 'ContractAccess(\"@1EditTable\")'),
        ('%[1]d_keys', 
        '{"insert": "ContractAccess(\"@1DLTTransfer\")", "update": "ContractAccess(\"@1DLTTransfer\")", 
          "new_column": "ContractAccess(\"@1NewColumn\")"}',
        '{}', 'ContractAccess(\"@1EditTable\")'),
        ('%[1]d_languages', 
        '{"insert": "ContractAccess(\"@1NewLang\")", "update": "ContractAccess(\"@1EditLang\")", 
          "new_column": "ContractAccess(\"@1NewColumn\")"}',
        '{}', 'ContractAccess(\"@1EditTable\")'),
        ('%[1]d_menu', 
        '{"insert": "ContractAccess(\"@1NewMenu\")", "update": "ContractAccess(\"@1EditMenu\")", 
          "new_column": "ContractAccess(\"@1NewColumn\")"}',
        '{}', 'ContractAccess(\"@1EditTable\")'),
        ('%[1]d_pages', 
        '{"insert": "ContractAccess(\"@1NewPage\")", "update": "ContractAccess(\"@1EditPage\")", 
          "new_column": "ContractAccess(\"@1NewColumn\")"}',
        '{}', 'ContractAccess(\"@1EditTable\")'),
        ('%[1]d_signatures', 
        '{"insert": "ContractAccess(\"@1NewSign\")", "update": "ContractAccess(\"@1EditSign\")", 
          "new_column": "ContractAccess(\"@1NewColumn\")"}',
        '{}', 'ContractAccess(\"@1EditTable\")');

