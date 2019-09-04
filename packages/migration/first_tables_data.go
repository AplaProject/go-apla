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

var firstTablesDataSQL = `
INSERT INTO "1_tables" ("id", "name", "permissions","columns", "conditions") VALUES
    (next_id('1_tables'), 'delayed_contracts',
        '{
            "insert": "ContractConditions(\"@1FullNodeCondition\")",
            "update": "ContractConditions(\"@1FullNodeCondition\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "contract": "ContractAccess(\"@1EditDelayedContract\")",
            "key_id": "ContractAccess(\"@1EditDelayedContract\")",
            "block_id": "ContractAccess(\"@1CallDelayedContract\",\"@1EditDelayedContract\")",
            "every_block": "ContractAccess(\"@1EditDelayedContract\")",
            "counter": "ContractAccess(\"@1CallDelayedContract\",\"@1EditDelayedContract\")",
            "limit": "ContractAccess(\"@1EditDelayedContract\")",
            "deleted": "ContractAccess(\"@1EditDelayedContract\")",
            "conditions": "ContractAccess(\"@1EditDelayedContract\")"
        }',
        'ContractConditions("@1AdminCondition")'
    ),
    (next_id('1_tables'), 'ecosystems',
        '{
            "insert": "ContractAccess(\"@1NewEcosystem\")",
            "update": "ContractAccess(\"@1EditEcosystemName\",\"@1VotingVesAccept\",\"@1EcManageInfo\",\"@1TeCreate\",\"@1TeChange\",\"@1TeBurn\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "name": "ContractAccess(\"@1EditEcosystemName\")",
            "info": "ContractAccess(\"@1EcManageInfo\")",
            "is_valued": "ContractAccess(\"@1VotingVesAccept\")",
            "emission_amount": "ContractAccess(\"@1TeCreate\",\"@1TeBurn\")",
            "token_title": "ContractAccess(\"@1TeCreate\")",
            "type_emission": "ContractAccess(\"@1TeCreate\",\"@1TeChange\")",
            "type_withdraw": "ContractAccess(\"@1TeCreate\",\"@1TeChange\")"
        }',
        'ContractConditions("@1AdminCondition")'
    ),
    (next_id('1_tables'), 'metrics',
        '{
            "insert": "ContractAccess(\"@1UpdateMetrics\")",
            "update": "ContractAccess(\"@1UpdateMetrics\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "time": "ContractAccess(\"@1UpdateMetrics\")",
            "metric": "ContractAccess(\"@1UpdateMetrics\")",
            "key": "ContractAccess(\"@1UpdateMetrics\")",
            "value": "ContractAccess(\"@1UpdateMetrics\")"
        }',
        'ContractConditions("@1AdminCondition")'
    ),
    (next_id('1_tables'), 'system_parameters',
        '{
            "insert": "false",
            "update": "ContractAccess(\"@1UpdateSysParam\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "value": "ContractAccess(\"@1UpdateSysParam\")",
            "name": "false",
            "conditions": "ContractAccess(\"@1UpdateSysParam\")"
        }',
        'ContractConditions("@1AdminCondition")'
    ),
    (next_id('1_tables'), 'bad_blocks',
        '{
            "insert": "ContractAccess(\"@1NewBadBlock\")",
            "update": "ContractAccess(\"@1NewBadBlock\", \"@1CheckNodesBan\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "block_id": "ContractAccess(\"@1CheckNodesBan\")",
            "producer_node_id": "ContractAccess(\"@1CheckNodesBan\")",
            "consumer_node_id": "ContractAccess(\"@1CheckNodesBan\")",
            "block_time": "ContractAccess(\"@1CheckNodesBan\")",
            "reason": "ContractAccess(\"@1CheckNodesBan\")",
            "deleted": "ContractAccess(\"@1CheckNodesBan\")"
        }',
        'ContractConditions("@1AdminCondition")'
    ),
    (next_id('1_tables'), 'node_ban_logs',
        '{
            "insert": "ContractAccess(\"@1CheckNodesBan\")",
            "update": "ContractAccess(\"@1CheckNodesBan\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "node_id": "ContractAccess(\"@1CheckNodesBan\")",
            "banned_at": "ContractAccess(\"@1CheckNodesBan\")",
            "ban_time": "ContractAccess(\"@1CheckNodesBan\")",
            "reason": "ContractAccess(\"@1CheckNodesBan\")"
        }',
        'ContractConditions("@1AdminCondition")'
    ),
    (next_id('1_tables'), 'time_zones',
        '{
            "insert": "false",
            "update": "false",
            "new_column": "false"
        }',
        '{
            "name": "false",
            "offset": "false"
        }',
        'ContractConditions("@1AdminCondition")'
    );
`
