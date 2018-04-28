package migration

var pagesDataSQL = `INSERT INTO "%[1]d_pages" (id, name, value, menu, conditions) VALUES
(2, 'app_binary', 'DBFind(buffer_data, src_buffer).Columns("value->app_id,value->app_name,value->menu_name,value->menu_id,value->count_menu").Where("key=''export'' and member_id=#key_id#").Vars(buffer)

If(#buffer_value_app_id# > 0){
    DBFind(applications, src_app).Where("id=#buffer_value_app_id#").Limit(1).Vars("app")

    Div(content-wrapper){
        SetTitle("Binary data": #app_name#)
        AddToolButton(Title: "Upload binary", Page: app_upload_binary, Icon: icon-plus, PageParams: "app_id=#app_id#")

        SetVar(pager_table, binaries).(pager_where, "app_id=#buffer_value_app_id#").(pager_page, app_binary).(pager_limit, 50)
        Include(pager_header)

        SetVar(admin_page, app_binary)
        Include(admin_link)

        DBFind(binaries, src_binparameters).Limit(#pager_limit#).Order(#sort_name#).Offset(#pager_offset#).Where("app_id=#buffer_value_app_id#")

        Form(panel panel-primary){
            Div(panel-body){
                Div(row){
                    ForList(src_binparameters){
                        Div(col-md-#width# col-sm-12){
                            Div(list-group-item){
                                Div(row){
                                    Div(col-md-4){
                                        Span(Class: h5 text-bold, Body: "#id#").Style(margin-right: 10px;)
                                        If(#member_id# == #key_id#){
                                            LinkPage(Class: text-primary h5, Body: #name#, Page: app_upload_binary, PageParams: "id=#id#,app_id=#buffer_value_app_id#")
                                        }.Else{
                                            Span(Class: h5, Body: #name#)
                                        }
                                    }
                                    Div(col-md-8 text-right){
                                        Span(#hash#)
                                    }
                                }
                            }
                        }
                    }
                }
            }
            Div(panel-footer clearfix){
                Include(pager)
            }
        }
    }
}.Else{
    SetTitle("Binary data")
    Div(breadcrumb){
        Span(Class: text-muted, Body: "You did not select the application. Viewing resources is not available")
    }
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(3, 'app_blocks', 'DBFind(buffer_data, src_buffer).Columns("value->app_id,value->app_name,value->menu_name,value->menu_id,value->count_menu").Where("key=''export'' and member_id=#key_id#").Vars(buffer)

If(#buffer_value_app_id# > 0){
    DBFind(applications, src_app).Where("id=#buffer_value_app_id#").Limit(1).Vars("app")

    Div(content-wrapper){
        SetTitle("Blocks": #app_name#)
        AddToolButton(Title: "Create", Page: editor, Icon: icon-plus, PageParams: "create=block,appId=#buffer_value_app_id#")

        SetVar(pager_table, blocks).(pager_where, "app_id=#buffer_value_app_id#").(pager_page, app_blocks).(pager_limit, 50)
        Include(pager_header)

        SetVar(admin_page, app_blocks)
        Include(admin_link)

        DBFind(blocks, src_blocks).Limit(#pager_limit#).Order(#sort_name#).Offset(#pager_offset#).Where("app_id=#buffer_value_app_id#")

        Form(panel panel-primary){
            Div(panel-body){
                Div(row){
                    ForList(src_blocks){
                        Div(col-md-#width# col-sm-12){
                            Div(list-group-item){
                                Div(row){
                                    Div(col-md-4){
                                        Span(Class: h5 text-bold, Body: "#id#").Style(margin-right: 10px;)
                                        Span(Class: h5, Body: "#name#")
                                    }
                                    Div(col-md-8){
                                        Div(pull-right){
                                            Span(LinkPage(Body: Em(Class: fa fa-cogs), Class: text-primary h4, Page: properties_edit, PageParams: "edit_property_id=#id#,type=block")).Style(margin-right: 15px;)
                                            Span(LinkPage(Body: Em(Class: fa fa-edit), Class: text-primary h4, Page: editor, PageParams: "open=block,name=#name#"))
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
            Div(panel-footer clearfix){
                Include(pager)
            }
        }
    }
}.Else{
    SetTitle("Blocks")
    Div(breadcrumb){
        Span(Class: text-muted, Body: "You did not select the application. Viewing resources is not available")
    }
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(4, 'app_contracts', 'DBFind(buffer_data, src_buffer).Columns("value->app_id,value->app_name,value->menu_name,value->menu_id,value->count_menu").Where("key=''export'' and member_id=#key_id#").Vars(buffer)

If(#buffer_value_app_id# > 0){
    DBFind(applications, src_app).Where("id=#buffer_value_app_id#").Limit(1).Vars("app")

    Div(content-wrapper){
        SetTitle("Contracts": #app_name#)
        AddToolButton(Title: "Create", Page: editor, Icon: icon-plus, PageParams: "create=contract,appId=#buffer_value_app_id#")

        SetVar(pager_table, contracts).(pager_where, "app_id=#buffer_value_app_id#").(pager_page, app_contracts).(pager_limit, 50)
        Include(pager_header)

        SetVar(admin_page, app_contracts)
        Include(admin_link)

        DBFind(contracts, src_contracts).Limit(#pager_limit#).Order(#sort_name#).Offset(#pager_offset#).Where("app_id=#buffer_value_app_id#")

        Form(panel panel-primary){
            Div(panel-body){
                Div(row){
                    ForList(src_contracts){
                        Div(col-md-#width# col-sm-12){
                            Div(list-group-item){
                                Div(row){
                                    Div(col-md-4){
                                        Span(Class: h5 text-bold, Body: "#id#").Style(margin-right: 10px;)
                                        Span(Class: h5, Body: "#name#")
                                    }
                                    Div(col-md-8){
                                        Div(pull-right){
                                            Span(LinkPage(Body: Em(Class: fa fa-cogs), Class: text-primary h4, Page: properties_edit, PageParams: "edit_property_id=#id#,type=contract")).Style(margin-right: 15px;)
                                            Span(LinkPage(Body: Em(Class: fa fa-edit), Class: text-primary h4, Page: editor, PageParams: "open=contract,name=#name#"))
                                        }
                                        Div(pull-right){
                                            If(#active#==1){
                                                Span(Class: h5, Body: Em(Class: fa fa-check)).Style(margin-right: 50px;)
                                            }.Else{
                                                Span(Class: h5 text-muted, Body: Em(Class: fa fa-minus)).Style(margin-right: 50px;)
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
            Div(panel-footer clearfix){
                Include(pager)
            }
        }
    }
}.Else{
    SetTitle("Contracts")
    Div(breadcrumb){
        Span(Class: text-muted, Body: "You did not select the application. Viewing resources is not available")
    }
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(5, 'app_edit', 'Div(content-wrapper){
    SetTitle("Application")
    Div(breadcrumb){
		LinkPage("Applications", apps_list)
		Span(/).Style(margin-right: 10px; margin-left: 10px;)
		If(#id# > 0){
			Span(Class: text-muted, Body: "Edit")
		}.Else{
			Span(Class: text-muted, Body: "New")
		}
    }
	
	Form(){
		If(#id# > 0){
			DBFind(applications, src_apps).Columns("id,name,conditions,deleted").Where("id=#id#").Limit(1).Vars("app")
			Div(col-md-12){
				Div(form-group){
					Div(text-left){
						Label("Name")
					}
					Input(Name: name, Disabled: "true", Value: #app_name#)
				}
				Div(form-group){
					Div(text-left){
						Label("Change conditions")
					}
					Input(Name: conditions, Value: #app_conditions#)
				}
				Div(row){
					Div(form-group){
						Div(text-left col-md-6){
							Button(Body: "Save", Class: btn btn-primary, Page: apps_list, Contract: EditApplication, Params: "ApplicationId=#id#,Conditions=Val(conditions)")
						}
						Div(text-right col-md-6){
							If(#app_deleted# == 0){
								Button(Body: "Delete", Class: btn btn-danger, Page: apps_list, Contract: DelApplication, Params: "ApplicationId=#app_id#,Value=1")
							}
						}
					}
				}
			}
		}.Else{
			Div(col-md-12){
				Div(form-group){
					Div(text-left){
						Label("Name")
					}
					Input(Name: name)
				}
				Div(form-group){
					Div(text-left){
						Label("Change conditions")
					}
					Input(Name: conditions)
				}
				Div(form-group){
					Div(text-left){
						Button(Body: "Save", Class: btn btn-primary, Page: apps_list, Contract: NewApplication, Params: "Name=Val(name),Conditions=Val(conditions)")
					}
				}
			}
		}
    }
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(6, 'app_langres', 'DBFind(buffer_data, src_buffer).Columns("value->app_id,value->app_name,value->menu_name,value->menu_id,value->count_menu").Where("key=''export'' and member_id=#key_id#").Vars(buffer)

If(#buffer_value_app_id# > 0){
    DBFind(applications, src_app).Where("id=#buffer_value_app_id#").Limit(1).Vars("app")

    Div(content-wrapper){
        SetTitle("Language resources": #app_name#)
        AddToolButton(Title: "Create", Page: langres_add, Icon: icon-plus, PageParams: "app_id=#app_id#")

        SetVar(pager_table, languages).(pager_where, "app_id=#buffer_value_app_id#").(pager_page, app_langres).(pager_limit, 50)
        Include(pager_header)

        SetVar(admin_page, app_langres)
        Include(admin_link)

        DBFind(languages, src_languages).Limit(#pager_limit#).Order(#sort_name#).Offset(#pager_offset#).Where("app_id=#buffer_value_app_id#")

        Form(panel panel-primary){
            Div(panel-body){
                Div(row){
                    ForList(src_languages){
                        Div(col-md-#width# col-sm-12){
                            Div(list-group-item){
                                Div(row){
                                    Div(col-md-4){
                                        Span(Class: h5 text-bold, Body: "#id#").Style(margin-right: 10px;)
                                        Span(Class: h5, Body: "#name#")
                                    }
                                    Div(col-md-8){
                                        Div(pull-right){
                                            Span(LinkPage(Body: Em(Class: fa fa-edit), Class: text-primary h4, Page: langres_edit, PageParams: "lang_id=#id#"))
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
            Div(panel-footer clearfix){
                Include(pager)
            }
        }
    }
}.Else{
    SetTitle("Language resources")
    Div(breadcrumb){
        Span(Class: text-muted, Body: "You did not select the application. Viewing resources is not available")
    }
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(7, 'app_pages', 'DBFind(buffer_data, src_buffer).Columns("value->app_id,value->app_name,value->menu_name,value->menu_id,value->count_menu").Where("key=''export'' and member_id=#key_id#").Vars(buffer)

If(#buffer_value_app_id# > 0){
    DBFind(applications, src_app).Where("id=#buffer_value_app_id#").Limit(1).Vars("app")

    Div(content-wrapper){
        SetTitle("Pages": #app_name#)
        AddToolButton(Title: "Create", Page: editor, Icon: icon-plus, PageParams: "create=page,appId=#buffer_value_app_id#")

        SetVar(pager_table, pages).(pager_where, "app_id=#buffer_value_app_id#").(pager_page, app_pages).(pager_limit, 50)
        Include(pager_header)

        SetVar(admin_page, app_pages)
        Include(admin_link)

        DBFind(pages, src_pages).Limit(#pager_limit#).Order(#sort_name#).Offset(#pager_offset#).Where("app_id=#buffer_value_app_id#")

        Form(panel panel-primary){
            Div(panel-body){
                Div(row){
                    ForList(src_pages){
                        Div(col-md-#width# col-sm-12){
                            Div(list-group-item){
                                Div(row){
                                    Div(col-md-4){
                                        Span(Class: h5 text-bold, Body: "#id#").Style(margin-right: 10px;)
                                        LinkPage(Page: #name#, Class: text-primary h5, Body: "#name#")
                                    }
                                    Div(col-md-8){
                                        Div(pull-right){
                                            Span(LinkPage(Body: Em(Class: fa fa-cogs), Class: text-primary h4, Page: properties_edit, PageParams: "edit_property_id=#id#,type=page")).Style(margin-right: 15px;)
                                            Span(LinkPage(Body: Em(Class: fa fa-edit), Class: text-primary h4, Page: editor, PageParams: "open=page,name=#name#"))
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
            Div(panel-footer clearfix){
                Include(pager)
            }
        }
    }
}.Else{
    SetTitle("Pages")
    Div(breadcrumb){
        Span(Class: text-muted, Body: "You did not select the application. Viewing resources is not available")
    }
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(8, 'app_params', 'DBFind(buffer_data, src_buffer).Columns("value->app_id,value->app_name,value->menu_name,value->menu_id,value->count_menu").Where("key=''export'' and member_id=#key_id#").Vars(buffer)

If(#buffer_value_app_id# > 0){
    DBFind(applications, src_app).Where("id=#buffer_value_app_id#").Limit(1).Vars("app")

    Div(content-wrapper){
        SetTitle("Application parameters": #app_name#)
        AddToolButton(Title: "Create", Page: app_params_edit, Icon: icon-plus, PageParams: "app_id=#app_id#,create=create")

        SetVar(pager_table, app_params).(pager_where, "app_id=#buffer_value_app_id#").(pager_page, app_params).(pager_limit, 50)
        Include(pager_header)

        SetVar(admin_page, app_params)
        Include(admin_link)

        DBFind(app_params, src_appparameters).Limit(#pager_limit#).Order(#sort_name#).Offset(#pager_offset#).Where("app_id=#buffer_value_app_id#")

        Form(panel panel-primary){
            Div(panel-body){
                Div(row){
                    ForList(src_appparameters){
                        Div(col-md-#width# col-sm-12){
                            Div(list-group-item){
                                Div(row){
                                    Div(col-md-4){
                                        Span(Class: h5 text-bold, Body: "#id#").Style(margin-right: 10px;)
                                        Span(Class: h5, Body: "#name#")
                                    }
                                    Div(col-md-8 text-right){
                                        Span(LinkPage(Body: Em(Class: fa fa-edit), Class: text-primary h4, Page: app_params_edit, PageParams: "id=#id#"))
                                    }
                                }
                            }
                        }
                    }
                }
            }
            Div(panel-footer clearfix){
                Include(pager)
            }
        }
    }
}.Else{
    SetTitle("Application parameters")
    Div(breadcrumb){
        Span(Class: text-muted, Body: "You did not select the application. Viewing resources is not available")
    }
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(9, 'app_params_edit', 'Div(content-wrapper){
    If(#create# == create){
        SetVar(param_name, "New")
    }.Else{
		DBFind(app_params, src_params).Where("id = #id#").Limit(1).Vars("param")
    }
	
	SetTitle("Application parameter")
	Div(Class: breadcrumb){
		LinkPage("Application parameters", app_params)
		Span(/).Style(margin-right: 10px; margin-left: 10px;)
		Span(Class: text-muted, Body: #param_name#)
	}

    Form(){
        Div(form-group){
            Label("Name")
            If(#create# == create){
                Input(Name: name)
            }.Else{
                Input(Name: name, Value: #param_name#, Disabled: "true")
            }
        }
        Div(form-group){
            If(#create# == create){
                Input(Type: textarea, Name: value).Style(height: 500px !important;)
            }.Else{
                Input(Type: textarea, Name: value, Value: "#param_value#").Style(height: 500px !important;)
            }
        }
        Div(form-group){
            Label("Change conditions")
            If(#create# == create){
                Input(Name: conditions)
            }.Else{
                Input(Name: conditions, Value: #param_conditions#)
            }
        }
        Div(form-group){
            If(#create# == create){
                Button(Class: btn btn-primary, Body: "Save", Contract: NewAppParam, Params: "Name=Val(name),Value=Val(value),Conditions=Val(conditions),ApplicationId=#app_id#", Page: app_params)
            }.Else{
                Button(Class: btn btn-primary, Body: "Save", Contract: EditAppParam, Params: "Id=#id#,Value=Val(value),Conditions=Val(conditions)", Page: app_params)
            }
        }
    }
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(10, 'app_tables', 'DBFind(buffer_data, src_buffer).Columns("value->app_id,value->app_name,value->menu_name,value->menu_id,value->count_menu").Where("key=''export'' and member_id=#key_id#").Vars(buffer)

If(#buffer_value_app_id# > 0){
    DBFind(applications, src_app).Where("id=#buffer_value_app_id#").Limit(1).Vars("app")

    Div(content-wrapper){
        SetTitle("Tables": #app_name#)
        AddToolButton(Title: "Create", Page: table_create, Icon: icon-plus, PageParams: "app_id=#app_id#")

        SetVar(pager_table, tables).(pager_where, "app_id=#buffer_value_app_id#").(pager_page, app_tables).(pager_limit, 50)
        Include(pager_header)

        SetVar(admin_page, app_tables)
        Include(admin_link)

        DBFind(tables, src_tables).Limit(#pager_limit#).Order(#sort_name#).Offset(#pager_offset#).Where("app_id=#buffer_value_app_id#")

        Form(panel panel-primary){
            Div(panel-body){
                Div(row){
                    ForList(src_tables){
                        Div(col-md-#width# col-sm-12){
                            Div(list-group-item){
                                Div(row){
                                    Div(col-md-4){
                                        Span(Class: h5 text-bold, Body: "#id#").Style(margin-right: 10px;)
                                        LinkPage(Page: table_view, Class: text-primary h5, Body: "#name#", PageParams: "table_name=#name#")
                                    }
                                    Div(col-md-8){
                                        Div(pull-right){
                                            Span(LinkPage(Body: Em(Class: fa fa-edit), Class: text-primary h4, Page: table_edit, PageParams: "tabl_id=#id#"))
                                        }
                                        Div(pull-right){
                                            DBFind(#name#).Columns("id").Count(countvar)
                                            Span(Class: h5 text-muted, Body: #countvar#).Style(margin-right: 50px;)
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
            Div(panel-footer clearfix){
                Include(pager)
            }
        }
    }
}.Else{
    SetTitle("Tables")
    Div(breadcrumb){
        Span(Class: text-muted, Body: "You did not select the application. Viewing resources is not available")
    }
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(11, 'app_upload_binary', 'Div(content-wrapper){
    SetTitle("Binary data")
    Div(breadcrumb){
        LinkPage("Binary data", app_binary)
        Span(/).Style(margin-right: 10px; margin-left: 10px;)
		If(#id# > 0){
			Span("Edit", text-muted)
			DBFind(binaries).Columns(name).Where(id = #id#).Vars(binary)
		}.Else{
			Span("Upload", text-muted)
		}
    }
	
	Form(){
		Div(form-group){
			Div(text-left){
				Label("Name")
			}
			If(#id# > 0){
				Input(Name: name, Disabled: disabled, Value: #binary_name#)
			}.Else{
				Input(Name: name)
			}
		}
		Div(form-group){
			Div(text-left){
				Label("File")
			}
			Input(Name: databin, Type: file)
		}
		Div(form-group text-left){
			Button(Body: "Upload", Contract: UploadBinary, Class: btn btn-primary, Params: "Name=Val(name),ApplicationId=#app_id#,Data=Val(databin),MemberID=#key_id#", Page: app_binary)
		}
    }
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(12, 'apps_list', 'Div(fullscreen){
    If(#deleted# == deleted){
        SetTitle("Inactive applications")
		Div(breadcrumb){
			LinkPage("Applications", apps_list)
			Span(/).Style(margin-right: 10px; margin-left: 10px;)
			Span(Class: text-muted, Body: "Inactive applications")
		}
	
        DBFind(applications, src_applications).Where("deleted=1").Order("id").Count(countvar).Custom(restore_btn){
            Button(Class: btn btn-link, Page: apps_list, Contract: DelApplication, Params: "ApplicationId=#id#", Body: "Restore")
        }
        If(#countvar# > 0) {
            Table(Source: src_applications, Columns: "ID=id,Name=name,Conditions=conditions,=restore_btn").Style(
                tbody > tr:nth-of-type(odd) {
                    background-color: #fafbfc;
                }
                tbody > tr > td {
                    word-break: break-all;
                    font-weight: 400;
                    font-size: 13px;
                    color: #666;
                    border-top: 1px solid #eee;
                    vertical-align: middle;
                }
                tr > *:first-child {
                    padding-left:20px;
                    width: 80px;
                }
                tr > *:last-child {
                    padding-right:80px;
                    text-align:right;
                    width: 100px;
                }
                thead {
                    background-color: #eee;
                }
            )
        }.Else{
            Div(content-wrapper){
                Span(Class: text-muted, Body: "You don''t have any inactive applications")
            }
        }
    }.Else{
        SetTitle("Applications")
		Div(breadcrumb){
			Span(Class: text-muted, Body: "This section is used to select installed applications")
		}
        AddToolButton(Title: "Inactive apps", Page: apps_list, Icon: icon-close, PageParams:"deleted=deleted")
        AddToolButton(Title: "Create", Page: app_edit, Icon: icon-plus)
	
        DBFind(buffer_data, src_buffer).Columns("value->app_id,value->app_name,value->menu_name,value->menu_id,value->count_menu").Where("key=''export'' and member_id=#key_id#").Vars(buffer)
        DBFind(applications, src_applications).Where("deleted=0").Order("id").Custom(custom_check){
            If(#id#==#buffer_value_app_id#){
                Span(Em(Class: fa fa-check)).Style(margin-left:30px;)
            }.Else{
                Button(Class: btn btn-link, Contract: Export_NewApp, Params: "app_id=#id#", Page: apps_list, Body: "select")
            }
        }.Custom(custom_actions){
            Button(Class: btn btn-link, Body: Em(Class: fa fa-edit), Page: app_edit, PageParams: "id=#id#")
        }

        Table(Source: src_applications, Columns: "ID=id,Name=name,Conditions=conditions,Selected=custom_check,=custom_actions").Style(
            tbody > tr:nth-of-type(odd) {
                background-color: #fafbfc;
            }
            tbody > tr > td {
                word-break: break-all;
                font-weight: 400;
                font-size: 13px;
                color: #666;
                border-top: 1px solid #eee;
                vertical-align: middle;
            }
            tr > *:first-child {
                padding-left:20px;
                width: 80px;
            }
            tr > *:last-child {
                padding-right:15px;
                text-align:right;
                width: 100px;
            }
            thead {
                background-color: #eee;
            }
        )
    }
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(13, 'column_add', 'Div(content-wrapper){
	SetTitle("Tables")
	Div(breadcrumb){
		Div(){
			LinkPage("Tables", app_tables)
			Span(/).Style(margin-right: 10px; margin-left: 10px;)
			LinkPage("Edit table", table_edit, PageParams:"tabl_id=#tabl_id#")
			Span(/).Style(margin-right: 10px; margin-left: 10px;)
			Span("Add column", text-muted)
		}
	}

	Form(panel panel-default){
		Div(panel-body){
			Div(form-group){
				Label("Column")
				Input(Name: ColumnName)
			}
			Div(form-group){
				Data(src_type,"type,name"){
					text,"Text"
					number,"Number"
					varchar,"Varchar"
					datetime,"Date/Time"
					money,"Money"
					double,"Double"
					character,"Character"
					json,"JSON"
				}
				Label("Type")
				Select(Name: Coltype, Source: src_type, NameColumn: name, ValueColumn: type, Value:"text")
			}
			Div(form-group){
				Label("Update")
				Input(Name: ColumnUp)
			}
		}
		Div(panel-footer clearfix){
			Button(Body: "Add column", Contract: NewColumn, Class: btn btn-primary, Page: table_edit, PageParams: "tabl_id=#tabl_id#", Params: "TableName=#next_table_name#,Name=Val(ColumnName),Type=Val(Coltype),Permissions=Val(ColumnUp)")
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(14, 'column_edit', 'Div(content-wrapper){
	SetTitle("Edit column")
	Div(breadcrumb){
		Div(){
			LinkPage("Tables", app_tables)
			Span(/).Style(margin-right: 10px; margin-left: 10px;)
			LinkPage("Edit table", table_edit, PageParams:"tabl_id=#tabl_id#")
			Span(/).Style(margin-right: 10px; margin-left: 10px;)
			Span("Edit column", text-muted)
		}
	}

	DBFind(tables, src_mem).Columns("id,name,columns,conditions").Vars(pre).WhereId(#tabl_id#)
	JsonToSource(src_columns, #pre_columns#)
	Form(panel panel-default){
		Div(panel-body){
			ForList(src_columns){
				If(#key# == #name_column#){
					Div(form-group){
						Label("Column")
						Input(Name: ColumnName, Disabled: "true", Value: #name_column#)
					}
					Div(form-group){
						Label("Type")
						SetVar(col_type, GetColumnType(#pre_name#, #key#))
						If(#col_type# == character){
							SetVar(input_type, "Character")
						}
						If(#col_type# == text){
							SetVar(input_type, "Text")
						}
						If(#col_type# == number){
							SetVar(input_type, "Number")
						}
						If(#col_type# == money){
							SetVar(input_type, "Money")
						}
						If(#col_type# == varchar){
							SetVar(input_type, "Varchar")
						}
						If(#col_type# == datetime){
							SetVar(input_type, "Date/Time")
						}
						If(#col_type# == double){
							SetVar(input_type, "Double")
						}
						If(#col_type# == character){
							SetVar(input_type, "Character")
						}
						If(#col_type# == json){
							SetVar(input_type, "JSON")
						}
						If(#col_type# == bytea){
							SetVar(input_type, "Binary Data")
						}
						If(#col_type# == uuid){
							SetVar(input_type, "UUID")
						}
						Input(Name: Coltype, Disabled: "true", Value: #input_type#)
					}
					Div(form-group){
						Label("Update")
						Input(Name: ColumnUp, Value: #value#)
					}
				}
			}
		}
		Div(panel-footer clearfix){
			Button(Body: "Save", Contract: EditColumn, Class: btn btn-primary, Page: table_edit, PageParams: "tabl_id=#tabl_id#", Params: "TableName=#pre_name#,Name=Val(ColumnName),Type=Val(Coltype),Permissions=Val(ColumnUp)")
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(15, 'export_download', 'Div(fullscreen){
    SetTitle("Export")
    Div(breadcrumb){
        Span(Class: text-muted, Body: "Payload was formed. You can download it now")
    }

    DBFind(binaries, src_binaries).Where("name=''export'' and member_id=#key_id# and app_id=1").Custom(app_name){
        DBFind(Name: buffer_data, Source: src_buffer).Columns("value->app_name").Where("key=''export'' and member_id=#key_id#").Vars(buffer)
        Span(#buffer_value_app_name#)
    }

    Table(Source: src_binaries, "Applications=app_name,=data").Style(
        tbody > tr:nth-of-type(odd) {
            background-color: #fafbfc;
        }
        tbody > tr > td {
            word-break: break-all;
            font-weight: 400;
            font-size: 13px;
            color: #666;
            border-top: 1px solid #eee;
            vertical-align: middle;
        }
        tr > *:first-child {
            padding-left:20px;
            width: 100px;
        }
        tr > *:last-child {
            padding-right:20px;
            text-align:right;
        }
        thead {
            background-color: #eee;
        }
    )
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(16, 'export_resources', 'Div(content-wrapper){
    SetTitle("Export")
    Div(breadcrumb){
        Span(Class: text-muted, Body: "Select the application which do you want to export and proceed to the payload generation process.")
    }

    Include(export_link)
    DBFind(buffer_data, src_buffer).Columns("value->app_id,value->app_name,value->menu_name,value->menu_id,value->count_menu").Where("key=''export'' and member_id=#key_id#").Vars(buffer)

    If(#buffer_value_app_id# > 0){
        If(#res_type#=="pages"){
            DBFind(pages, src).Custom(cbox){
                Input(Name: cbox, Type: checkbox, Value: true, Disabled: 1)
            }.Where("app_id = #buffer_value_app_id#").Order("id")
        }
        If(#res_type#=="blocks"){
            DBFind(blocks, src).Custom(cbox){
                Input(Name: cbox, Type: checkbox, Value: true, Disabled: 1)
            }.Where("app_id = #buffer_value_app_id#").Order("id")
        }
        If(#res_type#=="menu"){
            DBFind(menu, src).Custom(cbox){
                Input(Name: cbox, Type: checkbox, Value: true, Disabled: 1)
            }.Where("id in (#buffer_value_menu_id#)").Order("id")
        }
        If(#res_type#=="parameters"){
            DBFind(app_params, src).Custom(cbox){
                Input(Name: cbox, Type: checkbox, Value: true, Disabled: 1)
            }.Where("app_id = #buffer_value_app_id#").Order("id")
        }
        If(#res_type#=="languages"){
            DBFind(languages, src).Custom(cbox){
                Input(Name: cbox, Type: checkbox, Value: true, Disabled: 1)
            }.Where("app_id = #buffer_value_app_id#").Order("id")
        }
        If(#res_type#=="contracts"){
            DBFind(contracts, src).Custom(cbox){
                Input(Name: cbox, Type: checkbox, Value: true, Disabled: 1)
            }.Where("app_id = #buffer_value_app_id#").Order("id")
        }
        If(#res_type#=="tables"){
            DBFind(tables, src).Custom(cbox){
                Input(Name: cbox, Type: checkbox, Value: true, Disabled: 1)
            }.Where("app_id = #buffer_value_app_id#").Order("id")
        }
    }

    Div(row){
        Div(col-md-9 col-md-offset-0){
            Table(src, "ID=id,Name=name,=cbox").Style(
                tbody > tr:nth-of-type(odd) {
                    background-color: #fafbfc;
                }
                tbody > tr > td {
                    word-break: break-all;
                    padding: 8px 20px !important;
                    font-weight: 400;
                    font-size: 13px;
                    color: #666;
                    border-top: 1px solid #eee;
                    vertical-align: middle;
                }
                tr > *:first-child {
                    padding-left:20px;
                    width: 100px;
                }
                tr > *:last-child {
                    text-align:right;
                    padding-right:20px;
                    width: 50px;
                }
                thead {
                    background-color: #eee;
                }
            )
        }
        Div(col-md-3 col-md-offset-0){
            Include(export_info)
        }
    }
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(17, 'import_app', 'Div(content-wrapper){
    DBFind(buffer_data, src_buffer).Columns("id,value->name,value->data").Where("key=''import'' and member_id=#key_id#").Vars(prefix)
    DBFind(buffer_data, src_buffer).Columns("value->app_name,value->pages,value->pages_count,value->blocks,value->blocks_count,value->menu,value->menu_count,value->parameters,value->parameters_count,value->languages,value->languages_count,value->contracts,value->contracts_count,value->tables,value->tables_count").Where("key=''import_info'' and member_id=#key_id#").Vars(info)

    SetTitle("Import - #info_value_app_name#")
    Data(data_info, "name,count,info"){
        Pages,"#info_value_pages_count#","#info_value_pages#"
        Blocks,"#info_value_blocks_count#","#info_value_blocks#"
        Menu,"#info_value_menu_count#","#info_value_menu#"
        Parameters,"#info_value_parameters_count#","#info_value_parameters#"
        Language resources,"#info_value_languages_count#","#info_value_languages#"
        Contracts,"#info_value_contracts_count#","#info_value_contracts#"
        Tables,"#info_value_tables_count#","#info_value_tables#"
    }
    Div(breadcrumb){
        Span(Class: text-muted, Body: "Your data that you can import")
    }

    Div(panel panel-primary){
        ForList(data_info){
            Div(list-group-item){
                Div(row){
                    Div(col-md-10 mc-sm text-left){
                        Span(Class: text-bold, Body: "#name#")
                    }
                    Div(col-md-2 mc-sm text-right){
                        If(#count# > 0){
                            Span(Class: text-bold, Body: "(#count#)")
                        }.Else{
                            Span(Class: text-muted, Body: "(0)")
                        }
                    }
                }
                Div(row){
                    Div(col-md-12 mc-sm text-left){
                        If(#count# > 0){
                            Span(Class: h6, Body: "#info#")
                        }.Else{
                            Span(Class: text-muted h6, Body: "Nothing selected")
                        }
                    }
                }
            }
        }
        If(#prefix_id# > 0){
            Div(list-group-item text-right){
                Button(Body: "Import", Class: btn btn-primary, Page: apps_list).CompositeContract("Import", "#prefix_value_data#")
            }
        }
    }
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(18, 'import_upload', 'Div(content-wrapper){
    SetTitle("Import")
    Div(breadcrumb){
        Span(Class: text-muted, Body: "Select payload that you want to import")
    }
    Form(panel panel-primary){
        Div(list-group-item){
            Input(Name: input_file, Type: file)
        }
        Div(list-group-item text-right){
            Button(Body: "Load", Class: btn btn-primary, Contract: Import_Upload, Page: import_app)
        }
    }
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(19, 'langres_edit', 'Div(content-wrapper){
	SetTitle("Language resources")
	Div(Class: breadcrumb){
		LinkPage("Language resources", app_langres)
		Span(/).Style(margin-right: 10px; margin-left: 10px;)
		Span(Class: text-muted, Body: "Edit")
	}
	
	Form(panel panel-default){
		Div(panel-body){
			DBFind(languages, src_leng).Vars(pre).WhereId(#lang_id#)
			Div(row){
				Div(col-md-12){
					Label("Name")
					Input(Name: LangName, Disabled: "true", Value: #pre_name#)
				}
			}
			Div(row){
				Div(col-md-1 mt-lg){
					Label(Class: text-muted, Body: "Locale")
				}
				Div(col-md-10 mt-lg){
					Label(Class: text-muted, Body: "Value")
				}
				Div(col-md-1 mt-lg){
					Label(Class: text-muted, Body: "Action")
				}
			}
			SetVar(json,#pre_res#)
			JsonToSource(pv, #json#)
			ForList(Source: pv){
				Div(row){
					Div(col-md-1 mt-sm){
						Input(Name: idshare, Value: #key#)
					}
					Div(col-md-10 mt-sm){
						Input(Name: share, Value: #value#)
					}
					Div(col-md-1 mt-sm){
					}
				}
			}
			If(#del# == 1){
				SetVar(next_count, Calculate( Exp: #count_sec# - 1, Type: int))
			}.Else{
				If(GetVar(count)==""){
					SetVar(count, 0)
					SetVar(next_count, Calculate( Exp: #count#, Type: int))
				}.Else{
					SetVar(next_count, Calculate( Exp: #count_sec# + 1, Type: int))
				}
			}
			Range(params_range, 0, #next_count#)
			ForList(Source: params_range){
				Div(row){
					Div(col-md-1 mt-sm){
						Input(Name:idshare)
					}
					Div(col-md-10 mt-sm){
						Input(Name:share)
					}
					Div(col-md-1 mt-sm){
						If(And(#next_count# == #params_range_index#, #next_count# > 0)){
							Button(Em(Class: fa fa-trash), Class: btn btn-default, PageParams: "lang_id=#lang_id#,count_sec=#next_count#,count=#count#,del=1", Page:langres_edit)
						}
					}
				}
			}
			Div(row){
				Div(col-md-12 mt-lg){
			        LinkPage(Body: "Add localization", Page: langres_edit, PageParams: "lang_id=#lang_id#,count_sec=#next_count#,count=#count#")
                }
            }
		}
		Div(panel-footer){
			Button(Body: "Save", Class: btn btn-primary, Contract: @1EditLang, Params: "Value=Val(share),IdLanguage=Val(idshare),Id=#lang_id#", Page: app_langres)
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(20, 'langres_add', 'Div(content-wrapper){
	SetTitle("Language resources")
	Div(Class: breadcrumb){
		LinkPage("Language resources", app_langres)
		Span(/).Style(margin-right: 10px; margin-left: 10px;)
		Span(Class: text-muted, Body: "Create")
	}

	Form(panel panel-default){
		Div(panel-body){
			Div(row){
				Div(col-md-12){
					Label("Name")
					Input(Name: LangName)
				}
			}
			Div(row){
				Div(col-md-1 mt-lg){
					Label(Class: text-muted, Body: "Locale")
				}
				Div(col-md-10 mt-lg){
					Label(Class: text-muted, Body: "Value")
				}
				Div(col-md-1 mt-lg){
					Label(Class: text-muted, Body: "Action")
				}
			}
			If(#del# == 1){
				SetVar(next_count, Calculate( Exp: #count_sec# - 1, Type: int))
			}.Else{
				If(GetVar(count)==""){
					SetVar(count, 0)
					SetVar(next_count, Calculate( Exp: #count# + 1, Type: int))
				}.Else{
					SetVar(next_count, Calculate( Exp: #count_sec# + 1, Type: int))
				}
			}
			Range(params_range, 0, #next_count#)
			ForList(Source: params_range){
				Div(row){
					Div(col-md-1 mt-sm){
						Input(Name:idshare)
					}
					Div(col-md-10 mt-sm){
						Input(Name:share)
					}
					Div(col-md-1 mt-sm){
						If(And(#next_count# == #params_range_index#, #next_count# > 1)){
							Button(Body: Em(Class: fa fa-trash), Class: btn btn-default, PageParams: "count_sec=#next_count#,count=#count#,del=1,app_id=#app_id#", Page: langres_add)
						}
					}
				}
			}
			Div(row){
				Div(col-md-12 mt-lg){
					LinkPage(Body: "Add localization", Page: langres_add, PageParams:"count_sec=#next_count#,count=#count#,app_id=#app_id#")
				}
			}
		}
		Div(panel-footer){
			Button(Body: "Save", Class: btn btn-primary, Contract:@1NewLang, Page: app_langres, Params: "ApplicationId=#app_id#,Name=Val(LangName),Value=Val(share),IdLanguage=Val(idshare)")
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(21, 'menus_list', 'Div(fullscreen){
    SetTitle("Menu")
    AddToolButton(Title: "Create", Page: editor, Icon: icon-plus, PageParams: "create=menu,appId=0")
	Div(breadcrumb){
		Span(Class: text-muted, Body: "This section is used to manage the menu")
	}

    DBFind(menu, src_menus).Limit(250).Order("id").Custom(action){
        Span(LinkPage(Body: Em(Class: fa fa-cogs), Class: text-primary h4, Page: properties_edit, PageParams: "edit_property_id=#id#,type=menu")).Style(margin-right: 20px;)
        Span(LinkPage(Body: Em(Class: fa fa-edit), Class: text-primary h4, Page: editor, PageParams: "open=menu,name=#name#"))
    }

    Table(src_menus, "ID=id,Name=name,Title=title,Conditions=conditions,=action").Style(
    tbody > tr:nth-of-type(odd) {
        background-color: #fafbfc; 
    }
    tbody > tr > td {
        word-break: break-all;
        font-weight: 400;
        font-size: 13px;
        color: #666;
        border-top: 1px solid #eee;
        vertical-align: middle;
    }
    tr  > *:first-child {
        padding-left:20px;
        width: 80px;
    }
    tr  > *:last-child {
        padding-right:30px;
        text-align:right; 
        width: 100px;
    }
    thead {
        background-color: #eee;
    })
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(22, 'params_edit', 'Div(content-wrapper){
	If(#stylesheet# == stylesheet){
		DBFind(parameters, src_params).Where(name=''#stylesheet#'').Vars("param")
	}.Else{
		If(#id#>0){
			DBFind(parameters, src_params).WhereId(#id#).Vars("param")
		}.Else{
			SetVar(param_name, "New")
		}
	}

	SetTitle("Ecosystem parameters")
    Div(Class: breadcrumb){
        LinkPage("Ecosystem parameters", params_list)
        Span(/).Style(margin-right: 10px; margin-left: 10px;)
        Span(Class: text-muted, Body: #param_name#)
    }
	
	Form(){
		Div(form-group){
			Label("Name")
			If(#param_id#>0){
				Input(Name: name, Value: #param_name#, Disabled: "true")
			}.Else{
				Input(Name: name)
			}
		}
		Div(form-group){
			If(#param_id#>0){
				Input(Type: textarea, Name: value, Value: "#param_value#").Style(height: 500px !important;)
			}.Else{
				Input(Type: textarea, Name: value).Style(height: 500px !important;)
			}
		}
		Div(form-group){
			Label("Change conditions")
			If(#param_id#>0){
				Input(Name: conditions, Value: #param_conditions#)
			}.Else{
				Input(Name: conditions)
			}
		}
		Div(form-group){
			If(#param_id#>0){
				Button(Class: btn btn-primary, Body: "Save", Contract: EditParameter, Params: "Id=#param_id#,Value=Val(value),Conditions=Val(conditions)", Page: params_list)
			}.Else{
				Button(Class: btn btn-primary, Body: "Save", Contract: NewParameter, Params: "Name=Val(name),Value=Val(value),Conditions=Val(conditions)", Page: params_list)
			}
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(23, 'params_list', 'Div(fullscreen){
    SetTitle("Ecosystem parameters")
    AddToolButton(Title: "Manage stylesheet", Page: params_edit, Icon: icon-picture, PageParams:"stylesheet=stylesheet")
    AddToolButton(Title: "Create", Page: params_edit, Icon: icon-plus)
    Div(breadcrumb){
        Span(Class: text-muted, Body: "This section is used to configure stored reusable parameters")
    }

    DBFind(parameters, src_appparameters).Order("id").Custom(custom_actions){
        LinkPage(Body: Em(Class: fa fa-edit), Class: text-primary h4, Page: params_edit, PageParams: "id=#id#")
    }

    Table(src_appparameters, "ID=id,Name=name,Application=app_id,Value=value,Conditions=conditions,=custom_actions").Style(
        tbody > tr:nth-of-type(odd) {
            background-color: #fafbfc;
        }
        tbody > tr > td {
            word-break: break-all;
            font-weight: 400;
            font-size: 13px;
            color: #666;
            border-top: 1px solid #eee;
            vertical-align: middle;
        }
        tr > *:first-child {
            padding-left:20px;
            width: 80px;
        }
        tr > *:last-child {
            padding-right:30px;
            text-align:right;
            width: 100px;
        }
        thead {
            background-color: #eee;
        }
    )
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(24, 'properties_edit', 'Div(Class: content-wrapper){
	SetTitle("Edit properties")
	Div(breadcrumb){
		Div(){
			If(#type# == page){
				LinkPage("Pages", app_pages)
				Span(/).Style(margin-right: 10px; margin-left: 10px;)
				Span("Edit page", text-muted)
				DBFind(Name: pages, Source: src_page).WhereId(#edit_property_id#).Vars(item)
				DBFind(menu, src_menus)
			}
			If(#type# == contract){
				LinkPage("Contracts", app_contracts)
				Span(/).Style(margin-right: 10px; margin-left: 10px;)
				Span("Edit contract", text-muted)
				DBFind(Name: contracts, Source: src_contract).WhereId(#edit_property_id#).Vars(item)
			}
            If(#type# == block){
				LinkPage("Blocks", app_blocks)
				Span(/).Style(margin-right: 10px; margin-left: 10px;)
				Span("Edit block", text-muted)
				DBFind(Name: blocks, Source: src_block).WhereId(#edit_property_id#).Vars(item)
			}
			If(#type# == menu){
				LinkPage("Menu", menus_list)
				Span(/).Style(margin-right: 10px; margin-left: 10px;)
				Span("Edit menu", text-muted)
				DBFind(Name: menu, Source: src_menu).WhereId(#edit_property_id#).Vars(item)
			}
		}
	}
    Form(){
		Div(form-group){
			Label("Name")
			Input(Name: Name, Value: #item_name#, Disabled: "true")
		}
		If(#type# == page){
			Div(form-group){
				Label("Menu")
				Select(Name: Menu, Source: src_menus, NameColumn: name, ValueColumn: name, Value: #item_menu#)
			}
			Div(form-group){
				Label("Change conditions")
				Input(Name: Conditions, Value: #item_conditions#)
			}
			Div(form-group){
				Button(Body: "Save", Class: btn btn-primary, Page: app_pages, Contract: EditPage, Params: "Menu=Val(Menu),Conditions=Val(Conditions),Id=#edit_property_id#")
			}
		}
		If(#type# == contract){
			Div(form-group){
				Label("Change conditions")
				Input(Name: Conditions, Value: #item_conditions#)
			}
			Div(form-group){
				Label("Wallet")
				Div(row){
					Div(col-md-10){
						Input(Name: Wallet,Value: Address(#item_wallet_id#))
					}
					Div(col-md-2){
						If(#item_active# == 0){
							Button(Body: "Bind", Class: btn btn-primary btn-block, Contract: ActivateContract, Params: "Id=#edit_property_id#", Page:app_contracts)
						}.Else{
							Button(Body: "Unbind", Class: btn btn-primary btn-block, Contract: DeactivateContract, Params: "Id=#edit_property_id#", Page:properties_edit, PageParams: "edit_property_id=#edit_property_id#,type=#type#")
						}
					}
				}
			}
			Div(form-group){
				Button(Body: "Save", Class: btn btn-primary, Page: app_contracts, Contract: EditContract, Params: "Conditions=Val(Conditions),WalletId=Val(Wallet),Id=#edit_property_id#")
			}
		}
		If(#type# == block){
			Div(form-group){
				Label("Change conditions")
				Input(Name: Conditions, Value: #item_conditions#)
			}
			Div(form-group){
				Button(Body: "Save", Class: btn btn-primary, Page: app_blocks, Contract: EditBlock, Params: "Conditions=Val(Conditions),Id=#edit_property_id#")
			}
		}
		If(#type# == menu){
			Div(form-group){
				Label("Menu title")
				Input(Name: Title, Value: #item_title#)
			}
			Div(form-group){
				Label("Change conditions")
				Input(Name: Conditions, Value: #item_conditions#)
			}
			Div(form-group){
				Button(Body: "Save", Class: btn btn-primary, Page: menus_list, Contract: EditMenu, Params: "Conditions=Val(Conditions),Id=#edit_property_id#,NameTitle=Val(Title)")
			}
		}
    }
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(25, 'table_create', 'Div(content-wrapper){
	SetTitle("Create table")
	Div(breadcrumb){
		Div(){
			LinkPage("Tables", app_tables)
			Span(/).Style(margin-right: 10px; margin-left: 10px;)
			Span("Create", text-muted)
		}
	}

	Form(){
		Div(panel panel-default){
			Div(panel-body){
				Div(row){
					Div(col-md-12){
						Label("Name")
						Input(Name:TableName)
					}
				}
				Div(row){
					Div(col-md-4 mt-lg){
						Label(Class: text-muted, Body: "Columns")
						Input(Name:disinp, Disabled: true, Value: id)
					}
					Div(col-md-7 mt-lg){
						Label(Class: text-muted, Body: "Type")
						Input(Name: disinp, Disabled: true, Value: Number)
					}
					Div(col-md-1 mt-lg){
						Label(Class: text-muted, Body: "Action")
					}
				}
				Data(src_type,"type,name"){
					text,"Text"
					number,"Number"
					varchar,"Varchar"
					datetime,"Date/Time"
					money,"Money"
					double,"Double"
					character,"Character"
					json,"JSON"
				}
				If(#del# == 1){
					SetVar(next_count, Calculate( Exp: #count_sec# - 1, Type: int))
				}.Else{
					If(GetVar(count)==""){
						SetVar(count, 0)
						SetVar(next_count, Calculate( Exp: #count# + 1, Type: int))
					}.Else{
						SetVar(next_count, Calculate( Exp: #count_sec# + 1, Type: int))
					}
				}
				Range(params_range, 0, #next_count#)
				ForList(Source: params_range){
					Div(row){
						Div(col-md-4 mt-sm){
							Input(Name:idshare)
						}
						Div(col-md-7 mt-sm){
							Select(Name: share, Source: src_type, NameColumn: name, ValueColumn: type,Value:"text")
						}
						Div(col-md-1 mt-sm){
							If(And(#next_count# == #params_range_index#, #next_count# > 1)){
								Button(Body: Em(Class: fa fa-trash), Class: btn btn-default, PageParams: "count_sec=#next_count#,count=#count#,del=1,app_id=#app_id#", Page: table_create)
							}
						}
					}
				}			
			}
			Div(panel-footer){
				Button(Body: "Add column", Class: btn btn-primary, Page: table_create, PageParams: "count_sec=#next_count#,count=#count#,app_id=#app_id#")
			}
		}
		Div(row){
			Div(col-md-6){
				Div(panel panel-default){
					Div(panel-heading, Body: "Write permissions")
					Div(panel-body){
						Div(form-group){
							Label(Insert)
							Input(Name: Insert_con, Value: ContractConditions("MainCondition"))
						}
						Div(form-group){
							Label(Update)
							Input(Name: Update_con, Value: ContractConditions("MainCondition"))
						}
						Div(form-group){
							Label(New column)
							Input(Name: New_column_con, Value: ContractConditions("MainCondition"))
						}
					}
					Div(panel-footer){
						Button(Body: "Save", Class: btn btn-primary, Contract: @1NewTable, Page: app_tables, Params: "Shareholding=Val(share),Id=Val(idshare),ApplicationId=#app_id#")
					}
				}
			}
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(26, 'table_edit', 'Div(content-wrapper){
	SetTitle(Tables)
	Div(breadcrumb){
		Div(){
			LinkPage("Tables", app_tables)
			Span(/).Style(margin-right: 10px; margin-left: 10px;)
			Span("Edit", text-muted)
		}
	}

	Form(){
		Div(panel panel-default){
			Div(panel-body){
				Div(row){
					Div(col-md-3 h4){
						Label("Name")
					}
					Div(col-md-2 h4){
						Label("Type")
					}
					Div(col-md-5 h4){
						Label("Conditions")
					}
					Div(col-md-2 h4 text-right){
					}
				}
				DBFind(tables, src_mem).Columns("id,name,columns,conditions,permissions->insert,permissions->update,permissions->new_column").Vars(pre).WhereId(#tabl_id#)
				JsonToSource(src_columns, #pre_columns#)
				ForList(src_columns){
					Div(list-group-item){
						Div(row){
							Div(col-md-3 h5){
								Span(#key#)
							}
							Div(col-md-2 h5){
								SetVar(col_type,GetColumnType(#pre_name#, #key#))
								If(#col_type# == character){
									Span(Character)
								}
								If(#col_type# == text){
									Span("Text")
								}
								If(#col_type# == number){
									Span("Number")
								}
								If(#col_type# == money){
									Span("Money")
								}
								If(#col_type# == varchar){
									Span("Varchar")
								}
								If(#col_type# == datetime){
									Span("Date/Time")
								}
								If(#col_type# == double){
									Span("Double")
								}
								If(#col_type# == character){
									Span("Character")
								}
								If(#col_type# == json){
									Span("JSON")
								}
								If(#col_type# == bytea){
									Span("Binary Data")
								}
								If(#col_type# == uuid){
									Span("UUID")
								}
							}
							Div(col-md-5 h5){
								Span(#value#)
							}
							Div(col-md-2 text-right){
								Button(Body: "Edit", Class: btn btn-primary, Page: column_edit, PageParams: "name_column=#key#,tabl_id=#tabl_id#")
							}
						}
					}
				}
			}
			Div(panel-footer){
				Button(Body: "Add Column", Class: btn btn-primary, Page: column_add, PageParams: "next_table_name=#pre_name#,tabl_id=#tabl_id#")
			}
		}
		Div(row){
			Div(col-md-6){
				Div(panel panel-default){
					Div(panel-heading){Write permissions}
					Div(panel-body){
						Div(form-group){
							Label("Insert")
							Input(Name: Insert_con, Type: text, Value: #pre_permissions_insert#)
						}
						Div(form-group){
							Label("Update")
							Input(Name: Update_con, Type: text, Value: #pre_permissions_update#)
						}
						Div(form-group){
							Label("New column")
							Input(Name: New_column_con, Type: text, Value: #pre_permissions_new_column#)
						}
					}
					Div(panel-footer){
						Button(Body: "Save", Class: btn btn-primary, Contract: @1EditTable, Page: app_tables, Params: "Name=#pre_name#")
					}
				}
			}
			Div(col-md-6){
				Div(panel panel-default){
					Div(panel-heading){Conditions for changing permissions}
					Div(panel-body){
						Div(form-group){
							Input(Name: Insert_condition, Disabled:"true", Type: text, Value: #pre_conditions#)
						}
					}
				}
			}
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(27, 'table_view', 'Div(content-wrapper){
    SetTitle("Tables")
    Div(breadcrumb){
        LinkPage("Tables", app_tables)
        Span(/).Style(margin-right: 10px; margin-left: 10px;)
        Span(#table_name#, text-muted)
    }
	
	Div(panel panel-default){
		Div(panel-body){
			Div(table-responsive){
				DBFind(#table_name#, src_mem)
				Table(Source: src_mem)
			}
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
(28, 'admin_index', '', 'admin_menu', true),
(29,'notifications',$$DBFind(Name: notifications, Source: notifications_members).Columns("id,page_name,notification->icon,notification->header,notification->body").Where("closed=0 and notification->type='1' and recipient->member_id='#key_id#'")
			ForList(notifications_members){
				Div(Class: list-group-item){
					LinkPage(Page: #page_name#, PageParams: "notific_id=#id#"){
						Div(media-box){
							Div(Class: pull-left){
								Em(Class: fa #notification.icon# fa-1x text-primary)
							}
							Div(media-box-body clearfix){
								Div(Class: m0 text-normal, Body: #notification.header#)
								Div(Class: m0 text-muted h6, Body: #notification.body#)
							}
						}
					}
				}
			}

			DBFind(Name: notifications, Source: notifications_roles).Columns("id,page_name,notification->icon,notification->header,notification->body,recipient->role_id").Where("closed=0 and notification->type='2' and (date_start_processing is null or processing_info->member_id='#key_id#')")
			ForList(notifications_roles){
			    DBFind(Name: roles_participants, Source: src_roles).Columns("id").Where("member->member_id='#key_id#' and role->id='#recipient.role_id#' and deleted=0").Vars(prefix)
			    If(#prefix_id# > 0){
					Div(Class: list-group-item){
						LinkPage(Page: #page_name#, PageParams: "notific_id=#id#"){
							Div(media-box){
								Div(Class: pull-left){
									Em(Class: fa #notification.icon# fa-1x text-primary)
								}
								Div(media-box-body clearfix){
									Div(Class: m0 text-normal, Body: #notification.header#)
									Div(Class: m0 text-muted h6, Body: #notification.body#)
								}
							}
						}
					}
				}
}$$,'default_menu','ContractAccess("@1EditPage")');
`
