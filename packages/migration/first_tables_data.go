// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package migration

var firstTablesDataSQL = `
INSERT INTO "1_tables" ("id", "name", "permissions","columns", "conditions") VALUES
    (next_id('1_tables'), 'delayed_contracts',
        '{
            "insert": "ContractConditions(\"@1AdminCondition\")",
            "update": "ContractConditions(\"@1AdminCondition\")",
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
            "update": "ContractAccess(\"@1EditEcosystemName\",\"@1VotingDecisionCheck\",\"@1EcManageInfo\",\"@1TeCreate\",\"@1TeChange\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "name": "ContractAccess(\"@1EditEcosystemName\")",
            "info": "ContractAccess(\"@1EcManageInfo\")",
            "is_valued": "ContractAccess(\"@1VotingDecisionCheck\")",
            "emission_amount": "ContractAccess(\"@1TeCreate\")",
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
            "id": "ContractAccess(\"@1CheckNodesBan\")",
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
            "id": "ContractAccess(\"@1CheckNodesBan\")",
            "node_id": "ContractAccess(\"@1CheckNodesBan\")",
            "banned_at": "ContractAccess(\"@1CheckNodesBan\")",
            "ban_time": "ContractAccess(\"@1CheckNodesBan\")",
            "reason": "ContractAccess(\"@1CheckNodesBan\")"
        }',
        'ContractConditions("@1AdminCondition")'
    );
`
