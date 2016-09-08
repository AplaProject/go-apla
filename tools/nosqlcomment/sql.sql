DROP TABLE IF EXISTS dlt_transactions;

CREATE TABLE IF NOT EXISTS dlt_transactions (
`id` bigint(20) NOT NULL AUTO_INCREMENT  COMMENT '',
`sender_wallet_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '',
`recipient_wallet_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '',
`recipient_wallet_address` varbinary(512) NOT NULL DEFAULT '' COMMENT '',
`amount` decimal(15,2) NOT NULL DEFAULT '0' COMMENT '',
`commission` decimal(15,2) NOT NULL DEFAULT '0' COMMENT '',
`time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT 'Время, когда транзакцию создал юзер',
`comment` text CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT '',
`block_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT 'Блок, в котором данная транзакция была запечатана. При откате блока все транзакции с таким block_id будут удалены',
PRIMARY KEY (`id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 AUTO_INCREMENT=1 COMMENT='';


DROP TABLE IF EXISTS my_keys;

CREATE TABLE IF NOT EXISTS my_keys (
`id` int(11) NOT NULL AUTO_INCREMENT  COMMENT '',
`add_time` int(11) NOT NULL DEFAULT '0' COMMENT 'для удаления старых my_pending',
`notification` tinyint(1) NOT NULL DEFAULT '0' COMMENT '',
`public_key` varbinary(512) NOT NULL DEFAULT '' COMMENT 'Нужно для поиска в users',
`private_key` varchar(3096) NOT NULL DEFAULT '' COMMENT 'Хранят те, кто не боятся',
`password_hash` varchar(64) NOT NULL DEFAULT '' COMMENT 'Хранят те, кто не боятся',
`status` enum('my_pending','approved') NOT NULL DEFAULT 'my_pending' COMMENT '',
`my_time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT 'Время создания записи',
`time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT 'Время из блока',
`block_id` int(11) NOT NULL DEFAULT '0' COMMENT 'Для откатов и определения крайнего',
PRIMARY KEY (`id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 AUTO_INCREMENT=1 COMMENT='Ключи для авторизации юзера. Используем крайний';


DROP TABLE IF EXISTS my_node_keys;

CREATE TABLE IF NOT EXISTS my_node_keys (
`id` int(11) NOT NULL AUTO_INCREMENT  COMMENT '',
`add_time` int(11) NOT NULL DEFAULT '0' COMMENT 'для удаления старых my_pending',
`public_key` varbinary(512) NOT NULL DEFAULT '' COMMENT '',
`private_key` varchar(3096) NOT NULL DEFAULT '' COMMENT '',
`status` enum('my_pending','approved') NOT NULL DEFAULT 'my_pending' COMMENT '',
`my_time` int(11) NOT NULL DEFAULT '0' COMMENT 'Время создания записи',
`time` bigint(20) NOT NULL DEFAULT '0' COMMENT '',
`block_id` int(11) NOT NULL DEFAULT '0' COMMENT '',
PRIMARY KEY (`id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 AUTO_INCREMENT=1 COMMENT='';


DROP TABLE IF EXISTS transactions_status;

CREATE TABLE IF NOT EXISTS transactions_status (
`hash` binary(16) NOT NULL DEFAULT '' COMMENT '',
`time` int(11) NOT NULL DEFAULT '0' COMMENT '',
`type` int(11) NOT NULL DEFAULT '0' COMMENT '',
`wallet_id` int(11) NOT NULL DEFAULT '0' COMMENT '',
`citizen_id` int(11) NOT NULL DEFAULT '0' COMMENT '',
`block_id` int(11) NOT NULL DEFAULT '0' COMMENT '',
`error` varchar(255) NOT NULL DEFAULT '' COMMENT '',
PRIMARY KEY (`hash`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 COMMENT='Для удобства незарегенных юзеров на пуле. Показываем им статус их тр-ий';


DROP TABLE IF EXISTS confirmations;

CREATE TABLE IF NOT EXISTS confirmations (
`block_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '',
`good` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '',
`bad` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '',
`time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '',
PRIMARY KEY (`block_id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 COMMENT='Результаты сверки имеющегося у нас блока с блоками у случайных нодов';


DROP TABLE IF EXISTS block_chain;

CREATE TABLE IF NOT EXISTS block_chain (
`id` int(11) NOT NULL DEFAULT '0' COMMENT '',
`hash` binary(32) NOT NULL DEFAULT '' COMMENT 'Хэш от полного заголовка блока (new_block_id,prev_block_hash,merkle_root,time,user_id,level). Используется как PREV_BLOCK_HASH',
`data` longblob NOT NULL DEFAULT '' COMMENT '',
`cb_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '',
`wallet_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '',
`time` int(11) NOT NULL DEFAULT '0' COMMENT '',
`tx` text NOT NULL DEFAULT '' COMMENT '',
`cur_0l_miner_id` int(11) NOT NULL DEFAULT '0' COMMENT 'Майнер, который должен был сгенерить блок на 0-м уровне. Для отладки',
`max_miner_id` int(11) NOT NULL DEFAULT '0' COMMENT 'Макс. miner_id на момент, когда был записан этот блок. Для отладки',
PRIMARY KEY (`id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 COMMENT='Главная таблица. Хранит цепочку блоков';


DROP TABLE IF EXISTS currency;

CREATE TABLE IF NOT EXISTS currency (
`id` tinyint(3) unsigned NOT NULL AUTO_INCREMENT  COMMENT '',
`name` char(3) NOT NULL DEFAULT '' COMMENT '',
`full_name` varchar(50) NOT NULL DEFAULT '' COMMENT '',
`max_other_currencies` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '',
`rb_id` int(11) NOT NULL DEFAULT '0' COMMENT '',
PRIMARY KEY (`id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 AUTO_INCREMENT=1 COMMENT='';


DROP TABLE IF EXISTS info_block;

CREATE TABLE IF NOT EXISTS info_block (
`hash` binary(32) NOT NULL DEFAULT '' COMMENT 'Хэш от полного заголовка блока (new_block_id,prev_block_hash,merkle_root,time,user_id,level). Используется как prev_hash',
`block_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '',
`cb_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '',
`wallet_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '',
`time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT 'Время создания блока',
`level` tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT 'На каком уровне был сгенерирован блок',
`current_version` varchar(50) NOT NULL DEFAULT '0.0.1' COMMENT '',
`sent` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'Был ли блок отправлен нодам, указанным в nodes_connections'
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 COMMENT='Текущий блок, данные из которого мы уже занесли к себе';


DROP TABLE IF EXISTS rb_transactions;

CREATE TABLE IF NOT EXISTS rb_transactions (
`hash` binary(16) NOT NULL DEFAULT '' COMMENT '',
`time` int(11) NOT NULL DEFAULT '0' COMMENT '',
PRIMARY KEY (`hash`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 COMMENT='Храним данные за сутки, чтобы избежать дублей.';


DROP TABLE IF EXISTS main_lock;

CREATE TABLE IF NOT EXISTS main_lock (
`lock_time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '',
`script_name` varchar(100) NOT NULL DEFAULT '' COMMENT '',
`info` text NOT NULL DEFAULT '' COMMENT '',
`uniq` tinyint(4) NOT NULL DEFAULT '0' COMMENT '',
UNIQUE KEY (`uniq`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 COMMENT='Полная блокировка на поступление новых блоков/тр-ий';


DROP TABLE IF EXISTS full_nodes;

CREATE TABLE IF NOT EXISTS full_nodes (
`full_node_id` int(11) NOT NULL AUTO_INCREMENT  COMMENT '',
`host` varchar(100) NOT NULL DEFAULT '' COMMENT '',
`wallet_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '',
`state_id` int(11) NOT NULL DEFAULT '0' COMMENT '',
`final_delegate_wallet_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '',
`final_delegate_state_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '',
`rb_id` int(11) NOT NULL DEFAULT '0' COMMENT '',
PRIMARY KEY (`full_node_id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 AUTO_INCREMENT=1 COMMENT='';


DROP TABLE IF EXISTS rb_full_nodes;

CREATE TABLE IF NOT EXISTS rb_full_nodes (
`rb_id` bigint(20) NOT NULL AUTO_INCREMENT  COMMENT '',
`full_nodes_wallet_json` varbinary(1024) NOT NULL DEFAULT '' COMMENT '',
`block_id` int(11) NOT NULL DEFAULT '0' COMMENT 'В каком блоке было занесено. Нужно для удаления старых данных',
`prev_rb_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '',
PRIMARY KEY (`rb_id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 AUTO_INCREMENT=1 COMMENT='';


DROP TABLE IF EXISTS upd_full_nodes;

CREATE TABLE IF NOT EXISTS upd_full_nodes (
`time` int(11) NOT NULL DEFAULT '0' COMMENT '',
`rb_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '',
PRIMARY KEY (`rb_id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 AUTO_INCREMENT=1 COMMENT='';


DROP TABLE IF EXISTS rb_upd_full_nodes;

CREATE TABLE IF NOT EXISTS rb_upd_full_nodes (
`rb_id` bigint(20) NOT NULL AUTO_INCREMENT  COMMENT '',
`time` int(11) NOT NULL DEFAULT '0' COMMENT '',
`block_id` int(11) NOT NULL DEFAULT '0' COMMENT 'В каком блоке было занесено. Нужно для удаления старых данных',
`prev_rb_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '',
PRIMARY KEY (`rb_id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 AUTO_INCREMENT=1 COMMENT='';


DROP TABLE IF EXISTS queue_blocks;

CREATE TABLE IF NOT EXISTS queue_blocks (
`hash` binary(32) NOT NULL DEFAULT '' COMMENT '',
`full_node_id` int(11) NOT NULL DEFAULT '0' COMMENT '',
`block_id` int(11) NOT NULL DEFAULT '0' COMMENT '',
PRIMARY KEY (`hash`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 COMMENT='Блоки, которые мы должны забрать у указанных нодов';


DROP TABLE IF EXISTS queue_tx;

CREATE TABLE IF NOT EXISTS queue_tx (
`hash` binary(16) NOT NULL DEFAULT '' COMMENT 'md5 от тр-ии',
`data` longblob NOT NULL DEFAULT '' COMMENT '',
`_tmp_node_user_id` VARCHAR(255) DEFAULT '' COMMENT '',
PRIMARY KEY (`hash`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 COMMENT='Тр-ии, которые мы должны проверить';


DROP TABLE IF EXISTS transactions;

CREATE TABLE IF NOT EXISTS transactions (
`hash` binary(16) NOT NULL DEFAULT '' COMMENT 'Все хэши из этой таблы шлем тому, у кого хотим получить блок (т.е. недостающие тр-ии для составления блока)',
`data` longblob NOT NULL DEFAULT '' COMMENT 'Само тело тр-ии',
`verified` tinyint(1) NOT NULL DEFAULT '1' COMMENT 'Оставшиеся после прихода нового блока тр-ии отмечаются как "непроверенные" и их нужно проверять по новой',
`used` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'После того как попадют в блок, ставим 1, а те, у которых уже стояло 1 - удаляем',
`high_rate` tinyint(1) NOT NULL DEFAULT '0' COMMENT '1 - админские, 0 - другие',
`for_self_use` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'для new_pct(pct_generator), т.к. эта тр-ия валидна только вместе с блоком, который сгенерил тот, кто сгенерил эту тр-ию',
`type` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'Тип тр-ии. Нужно для недопущения попадения в блок 2-х тр-ий одного типа от одного юзера',
`wallet_id` int(11) NOT NULL DEFAULT '0' COMMENT 'Нужно для недопущения попадения в блок 2-х тр-ий одного типа от одного юзера',
`citizen_id` int(11) NOT NULL DEFAULT '0' COMMENT 'Нужно для недопущения попадения в блок 2-х тр-ий одного типа от одного юзера',
`third_var` int(11) NOT NULL DEFAULT '0' COMMENT 'Для исключения пересения в одном блоке удаления обещанной суммы и запроса на её обмен на DC. И для исключения голосования за один и тот же объект одним и тем же юзеров и одном блоке',
`counter` tinyint(3) NOT NULL DEFAULT '0' COMMENT 'Чтобы избежать зацикливания при проверке тр-ии: verified=1, новый блок, verified=0. При достижении 10-и - удаляем тр-ию ',
`sent` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Была отправлена нодам, указанным в nodes_connections',
PRIMARY KEY (`hash`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 COMMENT='Все незанесенные в блок тр-ии, которые у нас есть';


DROP TABLE IF EXISTS dlt_wallets;

CREATE TABLE IF NOT EXISTS dlt_wallets (
`wallet_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT  COMMENT '',
`address` varbinary(512) NOT NULL DEFAULT '' COMMENT '',
`public_key_0` varbinary(512) NOT NULL DEFAULT '' COMMENT 'Открытый ключ которым проверяются все транзакции от юзера',
`public_key_1` varbinary(512) NOT NULL DEFAULT '' COMMENT '2-й ключ, если есть',
`public_key_2` varbinary(512) NOT NULL DEFAULT '' COMMENT '3-й ключ, если есть',
`node_public_key` varbinary(512) NOT NULL DEFAULT '' COMMENT '',
`amount` decimal(30) NOT NULL DEFAULT '0' COMMENT '',
`host` varchar(50) NOT NULL DEFAULT '' COMMENT '',
`addressVote` varchar(255) NOT NULL DEFAULT '' COMMENT '',
`rb_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '',
PRIMARY KEY (`wallet_id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 AUTO_INCREMENT=1 COMMENT='';


DROP TABLE IF EXISTS dn_citizens;

CREATE TABLE IF NOT EXISTS dn_citizens (
`citizen_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT  COMMENT '',
`public_key_0` varbinary(512) NOT NULL DEFAULT '' COMMENT 'Открытый ключ которым проверяются все транзакции от юзера',
`public_key_1` varbinary(512) NOT NULL DEFAULT '' COMMENT '2-й ключ, если есть',
`public_key_2` varbinary(512) NOT NULL DEFAULT '' COMMENT '3-й ключ, если есть',
`rb_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '',
PRIMARY KEY (`citizen_id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 AUTO_INCREMENT=1 COMMENT='';


DROP TABLE IF EXISTS citizen_fields;

CREATE TABLE IF NOT EXISTS citizen_fields (
`state_code` varchar(2) NOT NULL DEFAULT '' COMMENT '',
`fields` varchar(255) NOT NULL DEFAULT '' COMMENT '',
`rb_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '',
PRIMARY KEY (`citizen_id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 AUTO_INCREMENT=1 COMMENT='';


DROP TABLE IF EXISTS rb_citizens;

CREATE TABLE IF NOT EXISTS rb_citizens (
`rb_id` bigint(20) NOT NULL AUTO_INCREMENT  COMMENT '',
`hash_public_key` binary(20) NOT NULL DEFAULT '' COMMENT 'При авторизации надо понять, какие даные показать юзеру, поэтому он нам шлет хэш, мы по нему ищем, пишем citizen_id в сессию. Юзеру удобно, т.к надо знать только свой приватный ключ',
`public_key_0` varbinary(512) NOT NULL DEFAULT '' COMMENT 'Открытый ключ которым проверяются все транзакции от юзера',
`public_key_1` varbinary(512) NOT NULL DEFAULT '' COMMENT '2-й ключ, если есть',
`public_key_2` varbinary(512) NOT NULL DEFAULT '' COMMENT '3-й ключ, если есть',
`block_id` int(11) NOT NULL DEFAULT '0' COMMENT 'В каком блоке было занесено. Нужно для удаления старых данных',
`prev_rb_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '',
PRIMARY KEY (`rb_id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 AUTO_INCREMENT=1 COMMENT='';


DROP TABLE IF EXISTS rb_dlt_wallets;

CREATE TABLE IF NOT EXISTS rb_dlt_wallets (
`rb_id` bigint(20) NOT NULL AUTO_INCREMENT  COMMENT '',
`hash` varbinary(512) NOT NULL DEFAULT '' COMMENT '',
`public_key_0` varbinary(512) NOT NULL DEFAULT '' COMMENT 'Открытый ключ которым проверяются все транзакции от юзера',
`public_key_1` varbinary(512) NOT NULL DEFAULT '' COMMENT '2-й ключ, если есть',
`public_key_2` varbinary(512) NOT NULL DEFAULT '' COMMENT '3-й ключ, если есть',
`node_public_key` varbinary(512) NOT NULL DEFAULT '' COMMENT '',
`amount` decimal(30) NOT NULL DEFAULT '0' COMMENT '',
`host` varchar(50) NOT NULL DEFAULT '' COMMENT '',
`addressVote` varchar(255) NOT NULL DEFAULT '' COMMENT '',
`block_id` int(11) NOT NULL DEFAULT '0' COMMENT 'В каком блоке было занесено. Нужно для удаления старых данных',
`prev_rb_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '',
PRIMARY KEY (`rb_id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 AUTO_INCREMENT=1 COMMENT='';


DROP TABLE IF EXISTS states;

CREATE TABLE IF NOT EXISTS states (
`state_id` bigint(20) NOT NULL AUTO_INCREMENT  COMMENT '',
`state_code` varchar(2) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT '',
`node_public_key` varbinary(512) NOT NULL DEFAULT '' COMMENT '',
`host` varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT '',
`delegate_wallet_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '',
`delegate_state_id` int(11) NOT NULL DEFAULT '0' COMMENT 'В каком блоке было занесено. Нужно для удаления старых данных',
PRIMARY KEY (`state_id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 AUTO_INCREMENT=1 COMMENT='';


DROP TABLE IF EXISTS install;

CREATE TABLE IF NOT EXISTS install (
`progress` varchar(10) NOT NULL DEFAULT '' COMMENT 'На каком шаге остановились'
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 COMMENT='Используется только в момент установки';


DROP TABLE IF EXISTS wallets_buffer;

CREATE TABLE IF NOT EXISTS wallets_buffer (
`hash` binary(16) NOT NULL DEFAULT '' COMMENT 'Хэш транзакции. Нужно для удаления данных из буфера, после того, как транзакция была обработана в блоке, либо анулирована из-за ошибок при повторной проверке',
`del_block_id` bigint(20) NOT NULL DEFAULT '0' COMMENT 'Т.к. удалять нельзя из-за возможного отката блока, приходится делать delete=1, а через сутки - чистить',
`user_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '',
`currency_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '',
`amount` decimal(15,2) unsigned NOT NULL DEFAULT '0' COMMENT '',
`block_id` bigint(20) NOT NULL DEFAULT '0' COMMENT 'Может быть = 0. Номер блока, в котором была занесена запись. Если блок в процессе фронт. проверки окажется невалдиным, то просто удалим все данные по block_id',
PRIMARY KEY (`hash`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 COMMENT='Суммируем все списания, которые еще не в блоке';


DROP TABLE IF EXISTS config;

CREATE TABLE IF NOT EXISTS config (
`my_block_id` int(11) NOT NULL DEFAULT '0' COMMENT 'Параллельно с info_block пишем и сюда. Нужно при обнулении рабочих таблиц, чтобы знать до какого блока не трогаем таблы my_',
`dlt_wallet_id` int(11) NOT NULL DEFAULT '0' COMMENT '',
`state_id` int(11) NOT NULL DEFAULT '0' COMMENT '',
`citizen_id` int(11) NOT NULL DEFAULT '0' COMMENT '',
`bad_blocks` text NOT NULL DEFAULT '' COMMENT 'Номера и sign плохих блоков. Нужно, чтобы не подцепить более длинную, но глючную цепочку блоков',
`pool_tech_works` tinyint(1) NOT NULL DEFAULT '0' COMMENT '',
`auto_reload` int(11) NOT NULL DEFAULT '0' COMMENT 'Если произойдет сбой и в main_lock будет висеть запись более auto_reload секунд, тогда будет запущен сбор блоков с чистого листа',
`setup_password` varchar(255)  NOT NULL DEFAULT '' COMMENT 'После установки и после сбора блоков, появляется окно, когда кто-угодно может ввести главный ключ',
`sqlite_db_url` varchar(255)  NOT NULL DEFAULT '' COMMENT 'Если не пусто, значит качаем с сервера sqlite базу данных',
`first_load_blockchain_url` varchar(255)  NOT NULL DEFAULT '' COMMENT '',
`first_load_blockchain` enum('nodes','file','null') DEFAULT 'null' COMMENT '',
`current_load_blockchain` enum('nodes','file','null') DEFAULT 'null' COMMENT 'Откуда сейчас собирается база данных',
`http_host` varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT 'адрес, по которому будет висеть панель юзера.  Если это майнер, то адрес должен совпадать с my_table.http_host',
`auto_update` tinyint(1) NOT NULL DEFAULT '0' COMMENT '',
`auto_update_url` varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT '',
`analytics_disabled` tinyint(1) NOT NULL DEFAULT '0' COMMENT '',
`stat_host` varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT ''
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 COMMENT='';


DROP TABLE IF EXISTS stop_daemons;

CREATE TABLE IF NOT EXISTS stop_daemons (
`stop_time` int(11) NOT NULL DEFAULT '0' COMMENT ''
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 COMMENT='Сигнал демонам об остановке';


DROP TABLE IF EXISTS incorrect_tx;

CREATE TABLE IF NOT EXISTS incorrect_tx (
`hash` binary(16) NOT NULL DEFAULT '' COMMENT 'md5 от тр-ии',
`time` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '',
`err` text CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT '',
PRIMARY KEY (`hash`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 COMMENT='';


DROP TABLE IF EXISTS migration_history;

CREATE TABLE IF NOT EXISTS migration_history (
`id` int(11) NOT NULL AUTO_INCREMENT  COMMENT '',
`version` int(11) NOT NULL DEFAULT '0' COMMENT '',
`date_applied` int(11) NOT NULL DEFAULT '0' COMMENT '',
PRIMARY KEY (`id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 AUTO_INCREMENT=1 COMMENT='';


DROP TABLE IF EXISTS dlt_wallets_buffer;

CREATE TABLE IF NOT EXISTS dlt_wallets_buffer (
`hash` binary(16) NOT NULL DEFAULT '' COMMENT 'Хэш транзакции. Нужно для удаления данных из буфера, после того, как транзакция была обработана в блоке, либо анулирована из-за ошибок при повторной проверке',
`del_block_id` bigint(20) NOT NULL DEFAULT '0' COMMENT 'Т.к. удалять нельзя из-за возможного отката блока, приходится делать delete=1, а через сутки - чистить',
`wallet_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '',
`amount` decimal(15,2) unsigned NOT NULL DEFAULT '0' COMMENT '',
`block_id` bigint(20) NOT NULL DEFAULT '0' COMMENT 'Может быть = 0. Номер блока, в котором была занесена запись. Если блок в процессе фронт. проверки окажется невалдиным, то просто удалим все данные по block_id',
PRIMARY KEY (`hash`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 COMMENT='Суммируем все списания, которые еще не в блоке';


DROP TABLE IF EXISTS president;

CREATE TABLE IF NOT EXISTS president (
`id` int(11) NOT NULL AUTO_INCREMENT  COMMENT '',
`state_code` varchar(2) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT '',
`citizen_id` int(11) NOT NULL DEFAULT '0' COMMENT '',
`start_time` bigint(20) NOT NULL DEFAULT '0' COMMENT '',
PRIMARY KEY (`id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 AUTO_INCREMENT=1 COMMENT='';


DROP TABLE IF EXISTS cb_head;

CREATE TABLE IF NOT EXISTS cb_head (
`id` int(11) NOT NULL AUTO_INCREMENT  COMMENT '',
`state_code` varchar(2) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT '',
`citizen_id` int(11) NOT NULL DEFAULT '0' COMMENT '',
PRIMARY KEY (`id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 AUTO_INCREMENT=1 COMMENT='';


DROP TABLE IF EXISTS dn_state_settings;

CREATE TABLE IF NOT EXISTS dn_state_settings (
`parameter` varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT '',
`value` varchar(2) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT '',
`change` varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT '',
PRIMARY KEY (`id`)
) ENGINE=MyISAM  DEFAULT CHARSET=latin1 AUTO_INCREMENT=1 COMMENT='';


