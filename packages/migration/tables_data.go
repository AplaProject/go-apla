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

var tablesDataSQL = `INSERT INTO "1_tables" ("id", "name", "permissions","columns", "conditions", "ecosystem") VALUES
    (next_id('1_tables'), 'contracts',
        '{
            "insert": "ContractConditions(\"DeveloperCondition\")",
            "update": "ContractConditions(\"DeveloperCondition\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{    
            "name": "false",
            "value": "ContractAccess(\"@1EditContract\")",
            "wallet_id": "ContractAccess(\"@1BindWallet\", \"@1UnbindWallet\")",
            "token_id": "ContractAccess(\"@1EditContract\")",
            "conditions": "ContractAccess(\"@1EditContract\")",
            "app_id": "ContractAccess(\"@1ItemChangeAppId\")",
            "ecosystem": "false"
        }',
        'ContractConditions("@1AdminCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'keys',
        '{
            "insert": "true",
            "update": "ContractAccess(\"@1TokensTransfer\",\"@1TokensLockoutMember\",\"@1MultiwalletCreate\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "pub": "false",
            "amount": "ContractAccess(\"@1TokensTransfer\")",
            "maxpay": "ContractConditions(\"@1AdminCondition\")",
            "deleted": "ContractConditions(\"@1AdminCondition\")",
            "blocked": "ContractAccess(\"@1TokensLockoutMember\")",
            "multi": "ContractAccess(\"@1MultiwalletCreate\")",
            "ecosystem": "false"
        }',
        'ContractConditions("@1AdminCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'history',
        '{
            "insert": "ContractAccess(\"@1TokensTransfer\")",
            "update": "ContractConditions(\"@1AdminCondition\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "sender_id": "false",
            "recipient_id": "false",
            "amount":  "false",
            "comment": "false",
            "block_id":  "false",
            "txhash": "false",
            "ecosystem": "false",
            "type": "false"
        }',
        'ContractConditions("@1AdminCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'languages',
        '{
            "insert": "ContractConditions(\"DeveloperCondition\")",
            "update": "ContractConditions(\"DeveloperCondition\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "name": "ContractAccess(\"@1EditLang\")",
            "res": "ContractAccess(\"@1EditLang\")",
            "conditions": "ContractAccess(\"@1EditLang\")",
            "ecosystem": "false"
        }',
        'ContractConditions("@1AdminCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'menu',
        '{
            "insert": "ContractConditions(\"DeveloperCondition\")",
            "update": "ContractConditions(\"DeveloperCondition\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "name": "false",
            "value": "ContractAccess(\"@1EditMenu\",\"@1AppendMenu\")",
            "title": "ContractAccess(\"@1EditMenu\")",
            "conditions": "ContractAccess(\"@1EditMenu\")",
            "ecosystem": "false"
        }',
        'ContractConditions("@1AdminCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'pages',
        '{
            "insert": "ContractConditions(\"DeveloperCondition\")",
            "update": "ContractConditions(\"DeveloperCondition\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "name": "false",
            "value": "ContractAccess(\"@1EditPage\",\"@1AppendPage\")",
            "menu": "ContractAccess(\"@1EditPage\")",
            "validate_count": "ContractAccess(\"@1EditPage\")",
            "validate_mode": "ContractAccess(\"@1EditPage\")",
            "app_id": "ContractAccess(\"@1ItemChangeAppId\")",
            "conditions": "ContractAccess(\"@1EditPage\")",
            "ecosystem": "false"
        }',
        'ContractConditions("@1AdminCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'blocks',
        '{
            "insert": "ContractConditions(\"DeveloperCondition\")",
            "update": "ContractConditions(\"DeveloperCondition\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "name": "false",
            "value": "ContractAccess(\"@1EditBlock\")",
            "conditions": "ContractAccess(\"@1EditBlock\")",
            "app_id": "ContractAccess(\"@1ItemChangeAppId\")",
            "ecosystem": "false"
        }',
        'ContractConditions("@1AdminCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'members',
        '{
            "insert": "ContractAccess(\"@1ProfileEdit\")",
            "update": "ContractAccess(\"@1ProfileEdit\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "image_id": "ContractAccess(\"@1ProfileEdit\")",
            "member_info": "ContractAccess(\"@1ProfileEdit\")",
            "member_name": "false",
            "ecosystem": "false"
        }',
        'ContractConditions("@1AdminCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'roles',
        '{
            "insert": "ContractAccess(\"@1RolesCreate\",\"@1RolesInstall\")",
            "update": "ContractAccess(\"@1RolesAccessManager\",\"@1RolesDelete\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "default_page": "false",
            "creator": "false",
            "deleted": "ContractAccess(\"@1RolesDelete\")",
            "company_id": "false",
            "date_deleted": "ContractAccess(\"@1RolesDelete\")",
            "image_id": "false",
            "role_name": "false",
            "date_created": "false",
            "roles_access": "ContractAccess(\"@1RolesAccessManager\")",
            "role_type": "false",
            "ecosystem": "false"
        }',
        'ContractConditions("@1AdminCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'roles_participants',
        '{
            "insert": "ContractAccess(\"@1RolesAssign\",\"@1VotingDecisionCheck\",\"@1RolesInstall\")",
            "update": "ContractAccess(\"@1RolesUnassign\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "deleted": "ContractAccess(\"@1RolesUnassign\")",
            "date_deleted": "ContractAccess(\"@1RolesUnassign\")",
            "member": "false",
            "role": "false",
            "date_created": "false",
            "appointed": "false",
            "ecosystem": "false"
        }',
        'ContractConditions("@1AdminCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'notifications',
        '{
            "insert": "ContractAccess(\"@1NotificationsSend\", \"@1CheckNodesBan\", \"@1NotificationsBroadcast\")",
            "update": "ContractAccess(\"@1NotificationsSend\", \"@1NotificationsClose\", \"@1NotificationsProcess\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "date_closed": "ContractAccess(\"@1NotificationsClose\")",
            "sender": "false",
            "processing_info": "ContractAccess(\"@1NotificationsClose\",\"@1NotificationsProcess\")",
            "date_start_processing": "ContractAccess(\"@1NotificationsClose\",\"@1NotificationsProcess\")",
            "notification": "false",
            "page_name": "false",
            "page_params": "false",
            "closed": "ContractAccess(\"@1NotificationsClose\")",
            "date_created": "false",
            "recipient": "false",
            "ecosystem": "false"
        }',
        'ContractConditions("@1AdminCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'sections',
        '{
            "insert": "ContractConditions(\"DeveloperCondition\")",
            "update": "ContractConditions(\"DeveloperCondition\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "title": "ContractAccess(\"@1EditSection\")",
            "urlname": "ContractAccess(\"@1EditSection\")",
            "page": "ContractAccess(\"@1EditSection\")",
            "roles_access": "ContractAccess(\"@1SectionRoles\")",
            "status": "ContractAccess(\"@1EditSection\",\"@1NewSection\")",
            "ecosystem": "false"
        }',
        'ContractConditions("@1AdminCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'applications',
        '{
            "insert": "ContractConditions(\"DeveloperCondition\")",
            "update": "ContractConditions(\"DeveloperCondition\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "name": "false",
            "uuid": "false",
            "conditions": "ContractAccess(\"@1EditApplication\")",
            "deleted": "ContractAccess(\"@1DelApplication\")",
            "ecosystem": "false"
        }',
        'ContractConditions("@1AdminCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'binaries',
        '{
            "insert": "ContractAccess(\"@1UploadBinary\")",
            "update": "ContractAccess(\"@1UploadBinary\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "hash": "ContractAccess(\"@1UploadBinary\")",
            "member_id": "false",
            "data": "ContractAccess(\"@1UploadBinary\")",
            "name": "false",
            "app_id": "false",
            "ecosystem": "false"
        }',
        'ContractConditions("@1AdminCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'parameters',
        '{
            "insert": "ContractConditions(\"DeveloperCondition\")",
            "update": "ContractConditions(\"DeveloperCondition\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "name": "false",
            "value": "ContractAccess(\"@1EditParameter\")",
            "conditions": "ContractAccess(\"@1EditParameter\")",
            "ecosystem": "false"
        }',
        'ContractConditions("@1AdminCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'app_params',
        '{
            "insert": "ContractConditions(\"DeveloperCondition\")",
            "update": "ContractConditions(\"DeveloperCondition\")",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "app_id": "ContractAccess(\"@1ItemChangeAppId\")",
            "name": "false",
            "value": "ContractAccess(\"@1EditAppParam\")",
            "conditions": "ContractAccess(\"@1EditAppParam\")",
            "ecosystem": "false"
        }',
        'ContractConditions("@1AdminCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'buffer_data',
        '{
            "insert": "true",
            "update": "true",
            "new_column": "ContractConditions(\"@1AdminCondition\")"
        }',
        '{
            "key": "false",
            "value": "true",
            "member_id": "false",
            "ecosystem": "false"
        }',
        'ContractConditions("@1AdminCondition")', '%[1]d'
    );
`
