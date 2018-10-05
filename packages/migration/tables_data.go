package migration

var tablesDataSQL = `INSERT INTO "1_tables" ("id", "name", "permissions","columns", "conditions", "ecosystem") VALUES
    (next_id('1_tables'), 'contracts',
        '{
            "insert": "ContractAccess(\"@1NewContract\")",
            "update": "ContractAccess(\"@1ActivateContract\", \"@1DeactivateContract\",\"@1EditContract\", \"@1ItemChangeAppId\")",
            "new_column": "ContractConditions(\"MainCondition\")"
        }',
        '{    
            "name": "false",
            "value": "ContractAccess(\"@1EditContract\")",
            "wallet_id": "ContractAccess(\"@1EditContract\")",
            "token_id": "ContractAccess(\"@1EditContract\")",
            "active": "ContractAccess(\"@1ActivateContract\", \"@1DeactivateContract\",\"@1EditContract\")",
            "conditions": "ContractAccess(\"@1EditContract\")",
            "app_id": "ContractAccess(\"@1ItemChangeAppId\")",
            "ecosystem": "false"
        }',
        'ContractAccess("@1EditTable")', '%[1]d'
    ),
    (next_id('1_tables'), 'keys',
        '{
            "insert": "true",
            "update": "true",
            "new_column": "ContractConditions(\"MainCondition\")"
        }',
        '{
            "pub": "ContractConditions(\"MainCondition\")",
            "amount": "ContractConditions(\"MainCondition\")",
            "maxpay": "ContractConditions(\"MainCondition\")",
            "deleted": "ContractConditions(\"MainCondition\")",
            "blocked": "ContractConditions(\"MainCondition\")",
            "multi": "ContractAccess(\"@1MultiwalletCreate\")",
            "ecosystem": "false"
        }',
        'ContractAccess("@1EditTable")', '%[1]d'
    ),
    (next_id('1_tables'), 'history',
        '{
            "insert": "ContractConditions(\"@1NodeOwnerCondition\")",
            "update": "ContractConditions(\"MainCondition\")",
            "new_column": "ContractConditions(\"MainCondition\")"
        }',
        '{
            "sender_id": "ContractConditions(\"MainCondition\")",
            "recipient_id": "ContractConditions(\"MainCondition\")",
            "amount":  "ContractConditions(\"MainCondition\")",
            "comment": "ContractConditions(\"MainCondition\")",
            "block_id":  "ContractConditions(\"MainCondition\")",
            "txhash": "ContractConditions(\"MainCondition\")",
            "ecosystem": "false"
        }',
        'ContractAccess("@1EditTable")', '%[1]d'
    ),
    (next_id('1_tables'), 'languages',
        '{
            "insert": "ContractConditions(\"MainCondition\")",
            "update": "ContractConditions(\"MainCondition\")",
            "new_column": "ContractConditions(\"MainCondition\")"
        }',
        '{
            "name": "ContractAccess(\"@1EditLang\")",
            "res": "ContractAccess(\"@1EditLang\")",
            "conditions": "ContractAccess(\"@1EditLang\")",
            "ecosystem": "false"
        }',
        'ContractAccess("@1EditTable")', '%[1]d'
    ),
    (next_id('1_tables'), 'menu',
        '{
            "insert": "ContractConditions(\"MainCondition\")",
            "update": "ContractConditions(\"MainCondition\")",
            "new_column": "ContractConditions(\"MainCondition\")"
        }',
        '{
            "name": "ContractConditions(\"MainCondition\")",
            "value": "ContractAccess(\"@1EditMenu\",\"@1AppendMenu\")",
            "conditions": "ContractAccess(\"@1EditMenu\")",
            "ecosystem": "false"
        }',
        'ContractAccess("@1EditTable")', '%[1]d'
    ),
    (next_id('1_tables'), 'pages',
        '{
            "insert": "ContractConditions(\"MainCondition\")",
            "update": "ContractConditions(\"MainCondition\")",
            "new_column": "ContractConditions(\"MainCondition\")"
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
        'ContractAccess("@1EditTable")', '%[1]d'
    ),
    (next_id('1_tables'), 'blocks',
        '{
            "insert": "ContractConditions(\"MainCondition\")",
            "update": "ContractConditions(\"MainCondition\")",
            "new_column": "ContractConditions(\"MainCondition\")"
        }',
        '{
            "name": "false",
            "value": "ContractAccess(\"@1EditBlock\")",
            "conditions": "ContractAccess(\"@1EditBlock\")",
            "app_id": "ContractAccess(\"@1ItemChangeAppId\")",
            "ecosystem": "false"
        }',
        'ContractAccess("@1EditTable")', '%[1]d'
    ),
    (next_id('1_tables'), 'members',
        '{
            "insert":"ContractAccess(\"@1ProfileEdit\")",
            "update":"true",
            "new_column":"ContractConditions(\"MainCondition\")"
        }',
        '{
            "image_id":"ContractAccess(\"@1ProfileEdit\")",
            "member_info":"ContractAccess(\"@1ProfileEdit\")",
            "member_name":"false",
            "ecosystem": "false"
        }',
        'ContractConditions("MainCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'roles',
        '{
            "insert":"ContractAccess(\"@1RolesCreate\",\"@1RolesInstall\")",
            "update":"true",
            "new_column":"ContractConditions(\"MainCondition\")"
        }',
        '{
            "default_page":"false",
            "creator":"false",
            "deleted":"ContractAccess(\"@1RolesDelete\")",
            "company_id":"false",
            "date_deleted":"ContractAccess(\"@1RolesDelete\")",
            "image_id":"ContractAccess(\"@1RolesCreate\")",
            "role_name":"false",
            "date_created":"false",
            "roles_access":"ContractAccess(\"@1RolesAccessManager\")",
            "role_type":"false",
            "ecosystem": "false"
        }',
        'ContractConditions("MainCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'roles_participants',
        '{
            "insert":"ContractAccess(\"@1RolesAssign\",\"@1VotingDecisionCheck\",\"@1RolesInstall\")",
            "update":"ContractConditions(\"MainCondition\")",
            "new_column":"ContractConditions(\"MainCondition\")"
        }',
        '{
            "deleted":"ContractAccess(\"@1RolesUnassign\")",
            "date_deleted":"ContractAccess(\"@1RolesUnassign\")",
            "member":"false",
            "role":"false",
            "date_created":"false",
            "appointed":"false",
            "ecosystem": "false"
        }',
        'ContractConditions("MainCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'notifications',
        '{
            "insert":"ContractAccess(\"@1NotificationsSend\", \"@1CheckNodesBan\", \"@1NotificationsBroadcast\")",
            "update":"ContractAccess(\"@1NotificationsSend\", \"@1NotificationsClose\", \"@1NotificationsProcess\")",
            "new_column":"ContractConditions(\"MainCondition\")"
        }',
        '{
            "date_closed":"ContractAccess(\"@1NotificationsClose\")",
            "sender":"false",
            "processing_info":"ContractAccess(\"@1NotificationsClose\",\"@1NotificationsProcess\")",
            "date_start_processing":"ContractAccess(\"@1NotificationsClose\",\"@1NotificationsProcess\")",
            "notification":"false",
            "page_name":"false",
            "page_params":"false",
            "closed":"ContractAccess(\"@1NotificationsClose\")",
            "date_created":"false",
            "recipient":"false",
            "ecosystem": "false"
        }',
        'ContractAccess("@1EditTable")', '%[1]d'
    ),
    (next_id('1_tables'), 'sections',
        '{
            "insert": "ContractConditions(\"MainCondition\")",
            "update": "ContractConditions(\"MainCondition\")",
            "new_column": "ContractConditions(\"MainCondition\")"
        }',
        '{
            "title": "ContractAccess(\"@1EditSection\")",
            "urlname": "ContractAccess(\"@1EditSection\")",
            "page": "ContractAccess(\"@1EditSection\")",
            "roles_access": "ContractAccess(\"@1SectionRoles\")",
            "status": "ContractAccess(\"@1EditSection\",\"@1NewSection\")",
            "ecosystem": "false"
        }',
        'ContractConditions("MainCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'applications',
        '{
            "insert": "ContractConditions(\"MainCondition\")",
            "update": "ContractConditions(\"MainCondition\")",
            "new_column": "ContractConditions(\"MainCondition\")"
        }',
        '{
            "name": "false",
            "uuid": "false",
            "conditions": "ContractAccess(\"@1EditApplication\")",
            "deleted": "ContractAccess(\"@1DelApplication\")",
            "ecosystem": "false"
        }',
        'ContractConditions("MainCondition")', '%[1]d'
    ),
    (next_id('1_tables'), 'binaries',
        '{
            "insert":"ContractAccess(\"@1UploadBinary\")",
            "update":"ContractAccess(\"@1UploadBinary\")",
            "new_column":"ContractConditions(\"MainCondition\")"
        }',
        '{
            "hash":"ContractAccess(\"@1UploadBinary\")",
            "member_id":"false",
            "data":"ContractAccess(\"@1UploadBinary\")",
            "name":"false",
            "app_id":"false",
            "ecosystem": "false"
        }',
        'ContractConditions(\"MainCondition\")', '%[1]d'
    ),
    (next_id('1_tables'), 'parameters',
        '{
            "insert": "ContractConditions(\"MainCondition\")",
            "update": "ContractConditions(\"MainCondition\")",
            "new_column": "ContractConditions(\"MainCondition\")"
        }',
        '{
            "name": "false",
            "value": "ContractAccess(\"@1EditParameter\")",
            "conditions": "ContractAccess(\"@1EditParameter\")",
            "ecosystem": "false"
        }',
        'ContractAccess("@1EditTable")', '%[1]d'
    ),
    (next_id('1_tables'), 'app_params',
        '{
            "insert": "ContractConditions(\"MainCondition\")",
            "update": "ContractConditions(\"MainCondition\")",
            "new_column": "ContractConditions(\"MainCondition\")"
        }',
        '{
            "app_id": "ContractAccess(\"@1ItemChangeAppId\")",
            "name": "false",
            "value": "ContractAccess(\"@1EditAppParam\")",
            "conditions": "ContractAccess(\"@1EditAppParam\")",
            "ecosystem": "false"
        }',
        'ContractAccess("@1EditTable")', '%[1]d'
    ),
    (next_id('1_tables'), 'buffer_data',
        '{
            "insert":"true",
            "update":"true",
            "new_column":"ContractConditions(\"MainCondition\")"
        }',
        '{
            "key": "false",
            "value": "true",
            "member_id": "false",
            "ecosystem": "false"
        }',
        'ContractConditions("MainCondition")', '%[1]d'
    );
`