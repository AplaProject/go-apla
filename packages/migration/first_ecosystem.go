package migration

// SchemaFirstEcosystem contains SQL queries for creating first ecosystem
var firstEcosystemSchema = `
DROP TABLE IF EXISTS "1_ecosystems";
CREATE TABLE "1_ecosystems" (
		"id" bigint NOT NULL DEFAULT '0',
		"name"	varchar(255) NOT NULL DEFAULT '',
		"is_valued" bigint NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "1_ecosystems" ADD CONSTRAINT "1_ecosystems_pkey" PRIMARY KEY ("id");


DROP TABLE IF EXISTS "1_system_parameters";
	CREATE TABLE "1_system_parameters" (
	"id" bigint NOT NULL DEFAULT '0',
	"name" varchar(255)  NOT NULL DEFAULT '',
	"value" text NOT NULL DEFAULT '',
	"conditions" text  NOT NULL DEFAULT ''
	);
	ALTER TABLE ONLY "1_system_parameters" ADD CONSTRAINT "1_system_parameters_pkey" PRIMARY KEY (id);
	CREATE INDEX "1_system_parameters_index_name" ON "1_system_parameters" (name);
	
	
	DROP TABLE IF EXISTS "1_delayed_contracts";
	CREATE TABLE "1_delayed_contracts" (
		"id" int NOT NULL default 0,
		"contract" varchar(255) NOT NULL DEFAULT '',
		"key_id" bigint NOT NULL DEFAULT '0',
		"block_id" bigint NOT NULL DEFAULT '0',
		"every_block" bigint NOT NULL DEFAULT '0',
		"counter" bigint NOT NULL DEFAULT '0',
		"limit" bigint NOT NULL DEFAULT '0',
		"deleted" bigint NOT NULL DEFAULT '0',
		"conditions" text NOT NULL DEFAULT ''
	);
	ALTER TABLE ONLY "1_delayed_contracts" ADD CONSTRAINT "1_delayed_contracts_pkey" PRIMARY KEY ("id");
	CREATE INDEX "1_delayed_contracts_index_block_id" ON "1_delayed_contracts" ("block_id");

	DROP TABLE IF EXISTS "1_metrics";
	CREATE TABLE "1_metrics" (
		"id" int NOT NULL default 0,
		"time" bigint NOT NULL DEFAULT '0',
		"metric" varchar(255) NOT NULL,
		"key" varchar(255) NOT NULL,
		"value" bigint NOT NULL
	);
	ALTER TABLE ONLY "1_metrics" ADD CONSTRAINT "1_metrics_pkey" PRIMARY KEY (id);
	CREATE INDEX "1_metrics_unique_index" ON "1_metrics" (metric, time, "key");

	DROP TABLE IF EXISTS "1_bad_blocks"; CREATE TABLE "1_bad_blocks" (
		"id" bigint NOT NULL DEFAULT '0',
		"producer_node_id" bigint NOT NULL,
		"block_id" bigint NOT NULL,
		"consumer_node_id" bigint NOT NULL,
		"block_time" timestamp NOT NULL,
		"reason" TEXT NOT NULL DEFAULT '',
		"deleted" bigint NOT NULL DEFAULT '0'
	);
	ALTER TABLE ONLY "1_bad_blocks" ADD CONSTRAINT "1_bad_blocks_pkey" PRIMARY KEY ("id");

	DROP TABLE IF EXISTS "1_node_ban_logs"; CREATE TABLE "1_node_ban_logs" (
		"id" bigint NOT NULL DEFAULT '0',
		"node_id" bigint NOT NULL,
		"banned_at" timestamp NOT NULL,
		"ban_time" bigint NOT NULL,
		"reason" TEXT NOT NULL DEFAULT ''
	);
	ALTER TABLE ONLY "1_node_ban_logs" ADD CONSTRAINT "1_node_ban_logs_pkey" PRIMARY KEY ("id");
`
var firstEcosystemCommon = `DROP TABLE IF EXISTS "1_keys"; CREATE TABLE "1_keys" (
	"id" bigint  NOT NULL DEFAULT '0',
	"pub" bytea  NOT NULL DEFAULT '',
	"amount" decimal(30) NOT NULL DEFAULT '0' CHECK (amount >= 0),
	"maxpay" decimal(30) NOT NULL DEFAULT '0' CHECK (maxpay >= 0),
	"multi" bigint NOT NULL DEFAULT '0',
	"deleted" bigint NOT NULL DEFAULT '0',
	"blocked" bigint NOT NULL DEFAULT '0',
	"ecosystem" bigint NOT NULL DEFAULT '1'
	);
	ALTER TABLE ONLY "1_keys" ADD CONSTRAINT "1_keys_pkey" PRIMARY KEY (ecosystem,id);

	DROP TABLE IF EXISTS "1_menu";
	CREATE TABLE "1_menu" (
		"id" bigint  NOT NULL DEFAULT '0',
		"name" character varying(255) NOT NULL DEFAULT '',
		"title" character varying(255) NOT NULL DEFAULT '',
		"value" text NOT NULL DEFAULT '',
		"conditions" text NOT NULL DEFAULT '',
		"ecosystem" bigint NOT NULL DEFAULT '1',
		UNIQUE (ecosystem, name)
	);
	ALTER TABLE ONLY "1_menu" ADD CONSTRAINT "1_menu_pkey" PRIMARY KEY (id);
	CREATE INDEX "1_menu_index_name" ON "1_menu" (ecosystem,name);

	DROP TABLE IF EXISTS "1_pages"; 
	CREATE TABLE "1_pages" (
		"id" bigint  NOT NULL DEFAULT '0',
		"name" character varying(255) NOT NULL DEFAULT '',
		"value" text NOT NULL DEFAULT '',
		"menu" character varying(255) NOT NULL DEFAULT '',
		"validate_count" bigint NOT NULL DEFAULT '1',
		"conditions" text NOT NULL DEFAULT '',
		"app_id" bigint NOT NULL DEFAULT '1',
		"validate_mode" character(1) NOT NULL DEFAULT '0',
		"ecosystem" bigint NOT NULL DEFAULT '1',
		UNIQUE (ecosystem, name)
	);
	ALTER TABLE ONLY "1_pages" ADD CONSTRAINT "1_pages_pkey" PRIMARY KEY (id);
	CREATE INDEX "1_pages_index_name" ON "1_pages" (ecosystem,name);

		
	DROP TABLE IF EXISTS "1_blocks"; CREATE TABLE "1_blocks" (
		"id" bigint  NOT NULL DEFAULT '0',
		"name" character varying(255) NOT NULL DEFAULT '',
		"value" text NOT NULL DEFAULT '',
		"conditions" text NOT NULL DEFAULT '',
		"app_id" bigint NOT NULL DEFAULT '1',
		"ecosystem" bigint NOT NULL DEFAULT '1',
		UNIQUE (ecosystem, name)
	);
	ALTER TABLE ONLY "1_blocks" ADD CONSTRAINT "1_blocks_pkey" PRIMARY KEY (id);
	CREATE INDEX "1_blocks_index_name" ON "1_blocks" (ecosystem,name);

	DROP TABLE IF EXISTS "1_languages"; CREATE TABLE "1_languages" (
		"id" bigint  NOT NULL DEFAULT '0',
		"name" character varying(100) NOT NULL DEFAULT '',
		"res" text NOT NULL DEFAULT '',
		"conditions" text NOT NULL DEFAULT '',
		"ecosystem" bigint NOT NULL DEFAULT '1'
	  );
	  ALTER TABLE ONLY "1_languages" ADD CONSTRAINT "1_languages_pkey" PRIMARY KEY (id);
	  CREATE INDEX "1_languages_index_name" ON "1_languages" (ecosystem, name);

	  CREATE TABLE "1_contracts" (
		"id" bigint NOT NULL  DEFAULT '0',
		"name" text NOT NULL DEFAULT '',
		"value" text  NOT NULL DEFAULT '',
		"wallet_id" bigint NOT NULL DEFAULT '0',
		"token_id" bigint NOT NULL DEFAULT '1',
		"active" character(1) NOT NULL DEFAULT '0',
		"conditions" text  NOT NULL DEFAULT '',
		"app_id" bigint NOT NULL DEFAULT '1',
		"ecosystem" bigint NOT NULL DEFAULT '1',
		UNIQUE(ecosystem,name)
		);
		ALTER TABLE ONLY "1_contracts" ADD CONSTRAINT "1_contracts_pkey" PRIMARY KEY (id);
		CREATE INDEX "1_contracts_index_ecosystem" ON "1_contracts" (ecosystem);

	DROP TABLE IF EXISTS "1_tables";
	CREATE TABLE "1_tables" (
	"id" bigint NOT NULL  DEFAULT '0',
	"name" varchar(100) NOT NULL DEFAULT '',
	"permissions" jsonb,
	"columns" jsonb,
	"conditions" text  NOT NULL DEFAULT '',
	"app_id" bigint NOT NULL DEFAULT '1',
	"ecosystem" bigint NOT NULL DEFAULT '1',
	UNIQUE(ecosystem,name)
    );
	ALTER TABLE ONLY "1_tables" ADD CONSTRAINT "1_tables_pkey" PRIMARY KEY ("id");
	CREATE INDEX "1_tables_index_name" ON "1_tables" (ecosystem, name);

	DROP TABLE IF EXISTS "1_parameters";
	CREATE TABLE "1_parameters" (
	"id" bigint NOT NULL  DEFAULT '0',
	"name" varchar(255) NOT NULL DEFAULT '',
	"value" text NOT NULL DEFAULT '',
	"conditions" text  NOT NULL DEFAULT '',
	"ecosystem" bigint NOT NULL DEFAULT '1',
	UNIQUE(ecosystem,name)
	);
	ALTER TABLE ONLY "1_parameters" ADD CONSTRAINT "1_parameters_pkey" PRIMARY KEY ("id");
	CREATE INDEX "1_parameters_index_name" ON "1_parameters" (ecosystem,name);

	DROP TABLE IF EXISTS "1_history"; CREATE TABLE "1_history" (
		"id" bigint NOT NULL  DEFAULT '0',
		"sender_id" bigint NOT NULL DEFAULT '0',
		"recipient_id" bigint NOT NULL DEFAULT '0',
		"amount" decimal(30) NOT NULL DEFAULT '0',
		"comment" text NOT NULL DEFAULT '',
		"block_id" bigint  NOT NULL DEFAULT '0',
		"txhash" bytea  NOT NULL DEFAULT '',
		"created_at" timestamp DEFAULT NOW(),
		"ecosystem" bigint NOT NULL DEFAULT '1'
		);
		ALTER TABLE ONLY "1_history" ADD CONSTRAINT "1_history_pkey" PRIMARY KEY (id);
		CREATE INDEX "1_history_index_sender" ON "1_history" (ecosystem, sender_id);
		CREATE INDEX "1_history_index_recipient" ON "1_history" (ecosystem, recipient_id);
		CREATE INDEX "1_history_index_block" ON "1_history" (block_id, txhash);
		
	DROP TABLE IF EXISTS "1_sections"; CREATE TABLE "1_sections" (
			"id" bigint  NOT NULL DEFAULT '0',
			"title" varchar(255)  NOT NULL DEFAULT '',
			"urlname" varchar(255) NOT NULL DEFAULT '',
			"page" varchar(255) NOT NULL DEFAULT '',
			"roles_access" jsonb,
			"status" bigint NOT NULL DEFAULT '0',
			"ecosystem" bigint NOT NULL DEFAULT '1'
			);
		  ALTER TABLE ONLY "1_sections" ADD CONSTRAINT "1_sections_pkey" PRIMARY KEY (id);
		  CREATE INDEX "1_sections_index_ecosystem" ON "1_sections" (ecosystem);
	

	DROP TABLE IF EXISTS "1_members";
		CREATE TABLE "1_members" (
			"id" bigint NOT NULL DEFAULT '0',
			"member_name"	varchar(255) NOT NULL DEFAULT '',
			"image_id"	bigint NOT NULL DEFAULT '0',
			"member_info"   jsonb,
			"ecosystem" bigint NOT NULL DEFAULT '1'
		);
		ALTER TABLE ONLY "1_members" ADD CONSTRAINT "1_members_pkey" PRIMARY KEY (ecosystem,id);
		CREATE INDEX "1_members_index_ecosystem" ON "1_members" (ecosystem);


	DROP TABLE IF EXISTS "1_roles";
		CREATE TABLE "1_roles" (
			"id" 	bigint NOT NULL DEFAULT '0',
			"default_page"	varchar(255) NOT NULL DEFAULT '',
			"role_name"	varchar(255) NOT NULL DEFAULT '',
			"deleted"    bigint NOT NULL DEFAULT '0',
			"role_type" bigint NOT NULL DEFAULT '0',
			"creator" jsonb NOT NULL DEFAULT '{}',
			"date_created" timestamp,
			"date_deleted" timestamp,
			"company_id" bigint NOT NULL DEFAULT '0',
			"roles_access" jsonb, 
			"image_id" bigint NOT NULL DEFAULT '0',
			"ecosystem" bigint NOT NULL DEFAULT '1'
		);
		ALTER TABLE ONLY "1_roles" ADD CONSTRAINT "1_roles_pkey" PRIMARY KEY ("id");
		CREATE INDEX "1_roles_index_deleted" ON "1_roles" (ecosystem, deleted);
		CREATE INDEX "1_roles_index_type" ON "1_roles" (ecosystem, role_type);


		DROP TABLE IF EXISTS "1_roles_participants";
		CREATE TABLE "1_roles_participants" (
			"id" bigint NOT NULL DEFAULT '0',
			"role" jsonb,
			"member" jsonb,
			"appointed" jsonb,
			"date_created" timestamp,
			"date_deleted" timestamp,
			"deleted" bigint NOT NULL DEFAULT '0',
			"ecosystem" bigint NOT NULL DEFAULT '1'
		);
		ALTER TABLE ONLY "1_roles_participants" ADD CONSTRAINT "1_roles_participants_pkey" PRIMARY KEY ("id");
		CREATE INDEX "1_roles_participants_ecosystem" ON "1_roles_participants" (ecosystem);

		DROP TABLE IF EXISTS "1_notifications";
		CREATE TABLE "1_notifications" (
			"id"    bigint NOT NULL DEFAULT '0',
			"recipient" jsonb,
			"sender" jsonb,
			"notification" jsonb,
			"page_params"	jsonb,
			"processing_info" jsonb,
			"page_name"	varchar(255) NOT NULL DEFAULT '',
			"date_created"	timestamp,
			"date_start_processing" timestamp,
			"date_closed" timestamp,
			"closed" bigint NOT NULL DEFAULT '0',
			"ecosystem" bigint NOT NULL DEFAULT '1'
		);
		ALTER TABLE ONLY "1_notifications" ADD CONSTRAINT "1_notifications_pkey" PRIMARY KEY ("id");
		CREATE INDEX "1_notifications_ecosystem" ON "1_notifications" (ecosystem);

		DROP TABLE IF EXISTS "1_applications";
		CREATE TABLE "1_applications" (
			"id" bigint NOT NULL DEFAULT '0',
			"name" varchar(255) NOT NULL DEFAULT '',
			"uuid" uuid NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
			"conditions" text NOT NULL DEFAULT '',
			"deleted" bigint NOT NULL DEFAULT '0',
			"ecosystem" bigint NOT NULL DEFAULT '1'
		);
		ALTER TABLE ONLY "1_applications" ADD CONSTRAINT "1_application_pkey" PRIMARY KEY ("id");
		CREATE INDEX "1_applications_ecosystem" ON "1_applications" (ecosystem);

		DROP TABLE IF EXISTS "1_binaries";
		CREATE TABLE "1_binaries" (
			"id" bigint NOT NULL DEFAULT '0',
			"app_id" bigint NOT NULL DEFAULT '1',
			"member_id" bigint NOT NULL DEFAULT '0',
			"name" varchar(255) NOT NULL DEFAULT '',
			"data" bytea NOT NULL DEFAULT '',
			"hash" varchar(64) NOT NULL DEFAULT '',
			"mime_type" varchar(255) NOT NULL DEFAULT '',
			"ecosystem" bigint NOT NULL DEFAULT '1'
		);
		ALTER TABLE ONLY "1_binaries" ADD CONSTRAINT "1_binaries_pkey" PRIMARY KEY (id);
		CREATE UNIQUE INDEX "1_binaries_index_app_id_member_id_name" ON "1_binaries" (ecosystem,app_id, member_id, name);
				
		DROP TABLE IF EXISTS "1_app_params";
		CREATE TABLE "1_app_params" (
		"id" bigint NOT NULL  DEFAULT '0',
		"app_id" bigint NOT NULL  DEFAULT '0',
		"name" varchar(255) NOT NULL DEFAULT '',
		"value" text NOT NULL DEFAULT '',
		"conditions" text  NOT NULL DEFAULT '',
		"ecosystem" bigint NOT NULL DEFAULT '1',
		UNIQUE(ecosystem,name)
		);
		ALTER TABLE ONLY "1_app_params" ADD CONSTRAINT "1_app_params_pkey" PRIMARY KEY ("id");
		CREATE INDEX "1_app_params_index_name" ON "1_app_params" (ecosystem,name);
		CREATE INDEX "1_app_params_index_app" ON "1_app_params" (ecosystem,app_id);
		
		DROP TABLE IF EXISTS "1_buffer_data";
		CREATE TABLE "1_buffer_data" (
			"id" bigint NOT NULL DEFAULT '0',
			"key" varchar(255) NOT NULL DEFAULT '',
			"value" jsonb,
			"member_id" bigint NOT NULL DEFAULT '0',
			"ecosystem" bigint NOT NULL DEFAULT '1'
		);
		ALTER TABLE ONLY "1_buffer_data" ADD CONSTRAINT "1_buffer_data_pkey" PRIMARY KEY ("id");
		CREATE INDEX "1_buffer_data_ecosystem" ON "1_buffer_data" (ecosystem);


	DROP TABLE IF EXISTS "1_roles";
		CREATE TABLE "1_roles" (
			"id" 	bigint NOT NULL DEFAULT '0',
			"default_page"	varchar(255) NOT NULL DEFAULT '',
			"role_name"	varchar(255) NOT NULL DEFAULT '',
			"deleted"    bigint NOT NULL DEFAULT '0',
			"role_type" bigint NOT NULL DEFAULT '0',
			"creator" jsonb NOT NULL DEFAULT '{}',
			"date_created" timestamp,
			"date_deleted" timestamp,
			"company_id" bigint NOT NULL DEFAULT '0',
			"roles_access" jsonb, 
			"image_id" bigint NOT NULL DEFAULT '0',
			"ecosystem" bigint NOT NULL DEFAULT '1'
		);
		ALTER TABLE ONLY "1_roles" ADD CONSTRAINT "1_roles_pkey" PRIMARY KEY ("id");
		CREATE INDEX "1_roles_index_deleted" ON "1_roles" (ecosystem, deleted);
		CREATE INDEX "1_roles_index_type" ON "1_roles" (ecosystem, role_type);


		DROP TABLE IF EXISTS "1_roles_participants";
		CREATE TABLE "1_roles_participants" (
			"id" bigint NOT NULL DEFAULT '0',
			"role" jsonb,
			"member" jsonb,
			"appointed" jsonb,
			"date_created" timestamp,
			"date_deleted" timestamp,
			"deleted" bigint NOT NULL DEFAULT '0',
			"ecosystem" bigint NOT NULL DEFAULT '1'
		);
		ALTER TABLE ONLY "1_roles_participants" ADD CONSTRAINT "1_roles_participants_pkey" PRIMARY KEY ("id");
		CREATE INDEX "1_roles_participants_ecosystem" ON "1_roles_participants" (ecosystem);

		DROP TABLE IF EXISTS "1_notifications";
		CREATE TABLE "1_notifications" (
			"id"    bigint NOT NULL DEFAULT '0',
			"recipient" jsonb,
			"sender" jsonb,
			"notification" jsonb,
			"page_params"	jsonb,
			"processing_info" jsonb,
			"page_name"	varchar(255) NOT NULL DEFAULT '',
			"date_created"	timestamp,
			"date_start_processing" timestamp,
			"date_closed" timestamp,
			"closed" bigint NOT NULL DEFAULT '0',
			"ecosystem" bigint NOT NULL DEFAULT '1'
		);
		ALTER TABLE ONLY "1_notifications" ADD CONSTRAINT "1_notifications_pkey" PRIMARY KEY ("id");
		CREATE INDEX "1_notifications_ecosystem" ON "1_notifications" (ecosystem);


`
