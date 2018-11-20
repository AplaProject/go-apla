package updates

var M114 = `
update "1_tables" set "columns" = "columns" || jsonb '{"created_at": "false"}' where name='history';
`
