package updates

var M115 = `
update "1_pages" set "conditions" = 'ContractConditions("@1DeveloperCondition")' 
		where "name" in ('default_page', 'admin_index', 'developer_index', 'notifications', 'import_app',
		'import_upload') and "conditions" = 'ContractAccess("@1EditPage")';
		
update "1_blocks" set "conditions" = 'ContractConditions("@1DeveloperCondition")' 
		where "name" in ('admin_link', 'export_info', 'export_link', 'pager', 'pager_header') and
		  "conditions" = 'ContractConditions("MainCondition")';
`
