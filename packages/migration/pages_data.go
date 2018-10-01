package migration

var pagesDataSQL = `INSERT INTO "1_pages" (id, name, value, menu, conditions, ecosystem) VALUES
	(next_id('1_pages'), 'admin_index', '', 'admin_menu','ContractAccess("@1EditPage")', '%[1]d'),
	(next_id('1_pages'), 'notifications', '', 'default_menu','ContractAccess("@1EditPage")', '%[1]d'),
	(next_id('1_pages'), 'import_app', 'Div(content-wrapper){
	DBFind(buffer_data, src_buffer).Columns("id,value->name,value->data").Where({key:import,member_id:#key_id#}).Vars(hash00001)
	DBFind(buffer_data, src_buffer).Columns("value->app_name,value->pages,value->pages_count,value->blocks,value->blocks_count,value->menu,value->menu_count,value->parameters,value->parameters_count,value->languages,value->languages_count,value->contracts,value->contracts_count,value->tables,value->tables_count").Where({key:import_info,member_id:#key_id#}).Vars(hash00002)

	SetTitle("Import - #hash00002_value_app_name#")
	Data(data_info, "hash00003_name,hash00003_count,hash00003_info"){
		Pages,"#hash00002_value_pages_count#","#hash00002_value_pages#"
		Blocks,"#hash00002_value_blocks_count#","#hash00002_value_blocks#"
		Menu,"#hash00002_value_menu_count#","#hash00002_value_menu#"
		Parameters,"#hash00002_value_parameters_count#","#hash00002_value_parameters#"
		Language resources,"#hash00002_value_languages_count#","#hash00002_value_languages#"
		Contracts,"#hash00002_value_contracts_count#","#hash00002_value_contracts#"
		Tables,"#hash00002_value_tables_count#","#hash00002_value_tables#"
	}
	Div(breadcrumb){
		Span(Class: text-muted, Body: "Your data that you can import")
	}

	Div(panel panel-primary){
		ForList(data_info){
			Div(list-group-item){
				Div(row){
					Div(col-md-10 mc-sm text-left){
						Span(Class: text-bold, Body: "#hash00003_name#")
					}
					Div(col-md-2 mc-sm text-right){
						If(#hash00003_count# > 0){
							Span(Class: text-bold, Body: "(#hash00003_count#)")
						}.Else{
							Span(Class: text-muted, Body: "(0)")
						}
					}
				}
				Div(row){
					Div(col-md-12 mc-sm text-left){
						If(#hash00003_count# > 0){
							Span(Class: h6, Body: "#hash00003_info#")
						}.Else{
							Span(Class: text-muted h6, Body: "Nothing selected")
						}
					}
				}
			}
		}
		If(#hash00001_id# > 0){
			Div(list-group-item text-right){
				Button(Body: "Import", Class: btn btn-primary, Page: apps_list).CompositeContract(@1Import, "#hash00001_value_data#")
			}
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")', '%[1]d'),
	(next_id('1_pages'), 'import_upload', 'Div(content-wrapper){
	SetTitle("Import")
	Div(breadcrumb){
		Span(Class: text-muted, Body: "Select payload that you want to import")
	}
	Form(panel panel-primary){
		Div(list-group-item){
			Input(Name: input_file, Type: file)
		}
		Div(list-group-item text-right){
			Button(Body: "Load", Class: btn btn-primary, Contract: @1ImportUpload, Page: import_app)
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")', '%[1]d');
`
