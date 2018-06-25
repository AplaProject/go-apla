package migration

var pagesDataSQL = `INSERT INTO "%[1]d_pages" (id, name, value, menu, conditions) VALUES
	(2, 'admin_dashboard', 'SetVar(this_page,admin_dashboard)
If(GetVar(block)){
	Div(breadcrumb){
		LinkPage(Body:Dashboard,Page:#this_page#)
		Span(/).Style(margin-right: 10px; margin-left: 10px;)
		Span(Class: text-muted, Body: Block: #block#)
	}
	Include(Name:#block#)
}.Else{
	SetTitle(Dashboard)
	DBFind(buffer_data).Columns("value->app_id").Where("key=''export'' and member_id=#key_id#").Vars(buffer)

	Data(tables, "Table,Cols,Page"){
		contracts,"id,app_id,name,active",editor
		pages,"id,app_id,name",editor
		blocks,"id,app_id,name",editor
		tables,"id,app_id,name",table_create
		app_params,"id,app_id,name,value",app_params_edit
		binaries,"id,app_id,name",app_upload_binary
		languages,"id,app_id,name",langres_add
	}
	DBFind(applications,src_apps).Order(id).Count(apps_count)
	SetVar(active_btn,"btn btn-info").(create_icon,fa fa-plus-square).(cols,3)
	If(GetVar(appid)){
		SetVar(where,"app_id=#appid#")
	}.Else{
		If(#buffer_value_app_id#>0){
			SetVar(appid,#buffer_value_app_id#).(where,"app_id=#appid#")
		}.Else{
			SetVar(where,"id>0").(appid,)
		}
	}
	Div(container){
		If(#apps_count#>8){
			Div(row){
				Form(col-sm-4 input-group form-group){
					Div(input-group-addon){
						App ID: #appid#
					}
					Select(Name:appid, Source:src_apps, NameColumn:name, ValueColumn:id, Value:#appid#)
					Div(input-group-btn){
						Button(Page: #this_page#, Class: #active_btn# fa fa-check, PageParams: "appid=Val(appid)")
						Button(Page: #this_page#, Class: btn btn-default fa fa-refresh, PageParams: "appid=#buffer_value_app_id#")
					}
				}
			}
		}.Else{
			Div(row){
				Div(col-sm-12 btn-group){
					ForList(src_apps){
						If(#id#==1){
							If(#appid#==#buffer_value_app_id#){
								Button(Class: btn btn-default disabled fa fa-refresh)
							}.Else{
								Button(Page: #this_page#, Class: btn btn-default fa fa-refresh, PageParams: "appid=#buffer_value_app_id#")
							}
						}
						If(#appid#==#id#){
							Button(Class: #active_btn# disabled, Body:"#id#:#name#")
						}.Else{
							Button(Page: #this_page#, Class: btn btn-default, PageParams: "appid=#id#", Body:"#id#:#name#")
						}
					}
				}
			}
		}
		Div(form-group){
			ForList(tables){
				DBFind(#Table#, src_table).Limit(250).Columns(#Cols#).Order("name").Where(#where#)
				Div(row){
					Div(h3){
						LangRes(#Table#)
					}
				}
				Div(row list-group-item){
					Div(cols){
						SetVar(value,)
						ForList(src_table){
							Div(clearfix){
								If(#Table#==contracts){
									LinkPage(Page: #Page#, PageParams: "open=contract,name=#name#"){#name#}
								}
								If(#Table#==pages){
									LinkPage(Page: #Page#, PageParams: "open=page,name=#name#"){#name#}
								}
								If(#Table#==blocks){
									LinkPage(Page: #Page#, PageParams: "open=block,name=#name#"){#name#}
								}
								If(#Table#==tables){
									LinkPage(Page: table_edit, PageParams: "tabl_id=#id#"){#name#}
								}
								If(#Table#==app_params){
									LinkPage(Page: #Page#, PageParams: "id=#id#"){#name#}
								}
								If(#Table#==binaries){
									LinkPage(Page: #Page#, PageParams: "id=#id#,application_id=#appid#"){#name#}
								}
								If(#Table#==languages){
									LinkPage(Class: text-primary h4, Page: langres_edit, PageParams: "lang_id=#id#"){#name#}
								}
								If(` + "`" + `#value#` + "`" + `!=""){
									:Div(text-muted){` + "`" + `#value#` + "`" + `}.Style(max-height:1.5em;overflow:hidden;)
								}
								Div(pull-right){
									If(#Table#==contracts){
										If(#active#==1){
											Span(actived,text-success mr-lg)
										}
										LinkPage(Class: text-muted fa fa-cogs, Page: properties_edit, PageParams: "edit_property_id=#id#,type=contract")
									}
									If(#Table#==pages){
										LinkPage(Class: text-muted fa fa-eye, Page: #name#)
										LinkPage(Class: text-muted fa fa-cogs, Page: properties_edit, PageParams: "edit_property_id=#id#,type=page")
									}
									If(#Table#==blocks){
										LinkPage(Class: text-muted fa fa-eye, Page: #this_page#, PageParams:"block=#name#")
										LinkPage(Class: text-muted fa fa-cogs, Page: properties_edit, PageParams: "edit_property_id=#id#,type=block")
									}
									If(#Table#==tables){
										LinkPage(Class: text-muted fa fa-eye, Page: table_view, PageParams: "tabl_id=#id#,table_name=#name#")
									}
								}
							}
						}
					}
					Div(row col-sm-12 mt-lg text-right){
						If(#Table#==contracts){
							LinkPage(Page: #Page#, PageParams: "create=contract,appId=#appid#"){
								Em(Class: #create_icon#) CREATE Em(Class: #create_icon#)
							}
						}.ElseIf(#Table#==pages){
							LinkPage(Page: #Page#, PageParams: "create=page,appId=#appid#"){
								Em(Class: #create_icon#) CREATE Em(Class: #create_icon#)
							}
						}.ElseIf(#Table#==blocks){
							LinkPage(Page: #Page#, PageParams: "create=block,appId=#appid#"){
								Em(Class: #create_icon#) CREATE Em(Class: #create_icon#)
							}
						}.ElseIf(#Table#==tables){
							LinkPage(Page: #Page#, PageParams: "application_id=#appid#"){
								Em(Class: #create_icon#) CREATE Em(Class: #create_icon#)
							}
						}.ElseIf(#Table#==app_params){
							LinkPage(Page: #Page#, PageParams: "application_id=#appid#,create=create"){
								Em(Class: #create_icon#) CREATE Em(Class: #create_icon#)
							}
						}.ElseIf(#Table#==binaries){
							LinkPage(Page: #Page#, PageParams: "application_id=#appid#"){
								Em(Class: #create_icon#) CREATE Em(Class: #create_icon#)
							}
						}.ElseIf(#Table#==languages){
							LinkPage(Page: #Page#, PageParams: "application_id=#appid#"){
								Em(Class: #create_icon#) CREATE Em(Class: #create_icon#)
							}
						}
					}
				}
			}
		}
	}.Style(
		.pull-right a {
			margin-right:10px;
		}
		.cols {
			-moz-column-count: #cols#;
			-webkit-column-count: #cols#;
			column-count: #cols#;
		}
	)
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
	(3, 'app_binary', 'DBFind(buffer_data, src_buffer).Columns("value->app_id").Where("key=''export'' and member_id=#key_id#").Vars(buffer)
If(#buffer_value_app_id# > 0){
	DBFind(applications, src_app).Where("id=#buffer_value_app_id#").Vars("application")

	Div(content-wrapper){
		SetTitle("Binary data": #application_name#)
		AddToolButton(Title: "Upload binary", Page: app_upload_binary, Icon: icon-plus, PageParams: "application_id=#application_id#")

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
											LinkPage(Class: text-primary h5, Body: #name#, Page: app_upload_binary, PageParams: "id=#id#,application_id=#buffer_value_app_id#")
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
	(4, 'app_blocks', 'DBFind(buffer_data, src_buffer).Columns("value->app_id").Where("key=''export'' and member_id=#key_id#").Vars(buffer)
If(#buffer_value_app_id# > 0){
	DBFind(applications, src_app).Where("id=#buffer_value_app_id#").Vars("application")

	Div(content-wrapper){
		SetTitle("Blocks": #application_name#)
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
	(5, 'app_contracts', 'DBFind(buffer_data, src_buffer).Columns("value->app_id").Where("key=''export'' and member_id=#key_id#").Vars(buffer)
If(#buffer_value_app_id# > 0){
	DBFind(applications, src_app).Where("id=#buffer_value_app_id#").Vars("application")

	Div(content-wrapper){
		SetTitle("Contracts": #application_name#)
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
	(6, 'app_edit', 'Div(content-wrapper){
	SetTitle("Applications")
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
			DBFind(applications, src_apps).Columns("id,name,conditions,deleted").Where("id=#id#").Vars("application")
			Div(form-group){
				Label("Name")
				Input(Name: Name, Disabled: "true", Value: #application_name#)
			}
			Div(form-group){
				Label("Change conditions")
				Input(Name: Conditions, Value: #application_conditions#)
			}
			Div(form-group){
				Div(row){
					Div(text-left col-md-6){
						Button(Body: "Save", Class: btn btn-primary, Page: apps_list, Contract: @1EditApplication, Params: "ApplicationId=#id#")
					}
					Div(text-right col-md-6){
						If(#application_deleted# == 0){
							Button(Body: "Delete", Class: btn btn-danger, Page: apps_list, Contract: @1DelApplication, Params: "ApplicationId=#application_id#,Value=1")
						}
					}
				}
			}
		}.Else{
			Div(form-group){
				Label("Name")
				Input(Name: Name)
			}
			Div(form-group){
				Label("Change conditions")
				Input(Name: Conditions)
			}
			Div(form-group){
				Div(text-left){
					Button(Body: "Save", Class: btn btn-primary, Page: apps_list, Contract: @1NewApplication)
				}
			}
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
	(7, 'app_langres', 'DBFind(buffer_data, src_buffer).Columns("value->app_id").Where("key=''export'' and member_id=#key_id#").Vars(buffer)
If(#buffer_value_app_id# > 0){
	DBFind(applications, src_app).Where("id=#buffer_value_app_id#").Vars("application")

	Div(content-wrapper){
		SetTitle("Language resources": #application_name#)
		AddToolButton(Title: "Create", Page: langres_add, Icon: icon-plus, PageParams: "application_id=#application_id#")

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
							Div(list-group-item clearfix){
								Span(Class: mr-sm text-bold, Body: "#id#")
								#name#
								LinkPage(Class:fa fa-edit pull-right, Page: langres_edit, PageParams: "lang_id=#id#")
							}
						}
					}
				}
			}
			Div(panel-footer){
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
	(8, 'app_pages', 'DBFind(buffer_data, src_buffer).Columns("value->app_id").Where("key=''export'' and member_id=#key_id#").Vars(buffer)
If(#buffer_value_app_id# > 0){
	DBFind(applications, src_app).Where("id=#buffer_value_app_id#").Vars("application")

	Div(content-wrapper){
		SetTitle("Pages": #application_name#)
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
	(9, 'app_params', 'DBFind(buffer_data, src_buffer).Columns("value->app_id").Where("key=''export'' and member_id=#key_id#").Vars(buffer)
If(#buffer_value_app_id# > 0){
	DBFind(applications, src_app).Where("id=#buffer_value_app_id#").Vars("application")

	Div(content-wrapper){
		SetTitle("Application parameters": #application_name#)
		AddToolButton(Title: "Create", Page: app_params_edit, Icon: icon-plus, PageParams: "application_id=#application_id#,create=create")

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
	(10, 'app_params_edit', 'Div(content-wrapper){
	If(#create# == create){
		SetVar(param_name, "New")
	}.Else{
		DBFind(app_params, src_params).Where("id=#id#").Vars("param")
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
				Button(Class: btn btn-primary, Body: "Save", Contract: @1NewAppParam, Params: "Name=Val(name),Value=Val(value),Conditions=Val(conditions),ApplicationId=#application_id#", Page: app_params)
			}.Else{
				Button(Class: btn btn-primary, Body: "Save", Contract: @1EditAppParam, Params: "Id=#id#,Value=Val(value),Conditions=Val(conditions)", Page: app_params)
			}
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
	(11, 'app_tables', 'DBFind(buffer_data, src_buffer).Columns("value->app_id").Where("key=''export'' and member_id=#key_id#").Vars(buffer)
If(#buffer_value_app_id# > 0){
	DBFind(applications, src_app).Where("id=#buffer_value_app_id#").Vars("application")

	Div(content-wrapper){
		SetTitle("Tables": #application_name#)
		AddToolButton(Title: "Create", Page: table_create, Icon: icon-plus, PageParams: "application_id=#application_id#")

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
										LinkPage(Page: table_view, Class: text-primary h5, Body: "#name#", PageParams: "tabl_id=#id#")
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
	(12, 'app_upload_binary', 'Div(content-wrapper){
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
			Button(Body: "Upload", Contract: @1UploadBinary, Class: btn btn-primary, Params: "Name=Val(name),ApplicationId=#application_id#,Data=Val(databin),MemberID=#key_id#", Page: app_binary)
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
	(13, 'apps_list', 'Div(fullscreen){
	If(#deleted# == deleted){
		SetTitle("Inactive applications")
		Div(breadcrumb){
			LinkPage("Applications", apps_list)
			Span(/).Style(margin-right: 10px; margin-left: 10px;)
			Span(Class: text-muted, Body: "Inactive applications")
		}
	
		DBFind(applications, src_applications).Where("deleted=1").Order("id").Count(countvar).Custom(restore_btn){
			Button(Class: btn btn-link, Page: apps_list, Contract: @1DelApplication, Params: "ApplicationId=#id#", Body: "Restore")
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
				Button(Class: btn btn-link, Contract: @1ExportNewApp, Params: "ApplicationId=#id#", Page: apps_list, Body: "select")
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
	(14, 'column_add', 'Div(content-wrapper){
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
			Button(Body: "Add column", Contract: @1NewColumn, Class: btn btn-primary, Page: table_edit, PageParams: "tabl_id=#tabl_id#", Params: "TableName=#next_table_name#,Name=Val(ColumnName),Type=Val(Coltype),Permissions=Val(ColumnUp)")
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
	(15, 'column_edit', 'Div(content-wrapper){
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
			Button(Body: "Save", Contract: @1EditColumn, Class: btn btn-primary, Page: table_edit, PageParams: "tabl_id=#tabl_id#", Params: "TableName=#pre_name#,Name=Val(ColumnName),Type=Val(Coltype),Permissions=Val(ColumnUp)")
		}
	}
}
', 'admin_menu', 'ContractAccess("@1EditPage")'),
	(16, 'confirmations', 'Div(fullscreen){
	SetTitle(Confirmations)
	AddToolButton(Title: "Create", Page: confirmations_new, Icon: icon-plus)
	Div(breadcrumb){
		Span(Class: text-muted, Body: "This section is used to manage contracts with confirmation")
	}

	DBFind(signatures, src_sign).Limit(250).Order("id").Columns("id,name,value->params,value->title,conditions").Custom(custom_title){
		Span(#value.title#)
	}.Custom(custom_params){
		Span(#value.params#)
	}.Custom(action){
		Span(LinkPage(Body: Em(Class: fa fa-edit), Class: text-primary h4, Page: confirmations_edit, PageParams: "sign_id=#id#"))
	}

	Table(Source:src_sign, Columns:"Contract=name,Title=custom_title,Params=custom_params,Conditions=conditions,=action").Style(
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
	(17, 'confirmations_edit', 'Div(content-wrapper){
	SetTitle("Confirmations")
	Div(Class: breadcrumb){
		LinkPage("Confirmations", confirmations)
		Span(/).Style(margin-right: 10px; margin-left: 10px;)
		Span(Class: text-muted, Body: "Edit")
	}

	Form(){
		DBFind(signatures, src_signatures).Columns("name,conditions,value->title,value->params").Vars(pre).WhereId(#sign_id#)
		Div(form-group){
			Label("Contract name")
			Input(Name: Name, Value: #pre_name#, Disabled: 1)
		}		
		Div(form-group){
			Label("Title of confirmation")
			Input(Name: Title, Value: #pre_value_title#)
		}		
		Div(form-group){
			Label("Parameters")
			Input(Name: Parameter, Value: #pre_value_params#)
		}		
		Div(form-group){
			Label("Conditions")
			Input(Name: Conditions, Value: #pre_conditions#)
		}
		Div(form-group){
			Button(Body: "Save", Class: btn btn-primary, Contract: @1EditSignJoint, Page: confirmations, Params: "Id=#sign_id#")
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
	(18, 'confirmations_new', 'Div(content-wrapper){
	SetTitle("Confirmations")
	Div(Class: breadcrumb){
		LinkPage("Confirmations", confirmations)
		Span(/).Style(margin-right: 10px; margin-left: 10px;)
		Span(Class: text-muted, Body: "Create")
	}

	Form(panel panel-default){
		Div(panel-body){
			Div(form-group){
				Label("Contract name")
				Input(Name: Name, Placeholder: "Name")
			}		
			Div(form-group){
				Label("Title of confirmation")
				Input(Name: Title, Placeholder: "Title")
			}
			Div(form-group){
				Label("Conditions")
				Input(Name: Conditions, Placeholder: "Conditions")
			}
			Div(row){
				Div(col-md-4){
					Label(Class: text-bold, Body: "Parameter")
				}
				Div(col-md-7){
					Label(Class: text-bold, Body: "Value")
				}
				Div(col-md-1){
					Label(Class: text-bold, Body: "Action")
				}
			}
			If(GetVar(cs)==""){
				SetVar(cs, Calculate( Exp: 0, Type: int))
			}
			If(#del# == 1){
				SetVar(cs, Calculate( Exp: #cs# - 1, Type: int))
			}.Else{
				SetVar(cs, Calculate( Exp: #cs# + 1, Type: int))
			}
			Range(params_range, 0, #cs#)
			ForList(Source: params_range){
				Div(row){
					Div(col-md-4 mt-sm){
						Input(Name:ParamArr)
					}
					Div(col-md-7 mt-sm){
						Input(Name:ValueArr)
					}
					Div(col-md-1 mt-sm){
						If(And(#cs#==#params_range_index#,#cs#>1)){
							Button(Body: Em(Class: fa fa-trash), Class: btn btn-default, PageParams: "cs=#cs#,del=1,application_id=#application_id#", Page: confirmations_new)
						}
					}
				}
			}
			Div(row){
				Div(col-md-12 mt-lg){
					LinkPage(Body: "Add parameter", Page: confirmations_new, PageParams:"cs=#cs#,application_id=#application_id#")
				}
			}
		}
		Div(panel-footer){
			Button(Body: "Save", Class: btn btn-primary, Contract: @1NewSignJoint, Page: confirmations)
		}
	}	
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
	(19, 'export_download', 'Div(fullscreen){
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
	(20, 'export_resources', 'Div(content-wrapper){
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
	(21, 'import_app', 'Div(content-wrapper){
	DBFind(buffer_data, src_buffer).Columns("id,value->name,value->data").Where("key=''import'' and member_id=#key_id#").Vars(hash00001)
	DBFind(buffer_data, src_buffer).Columns("value->app_name,value->pages,value->pages_count,value->blocks,value->blocks_count,value->menu,value->menu_count,value->parameters,value->parameters_count,value->languages,value->languages_count,value->contracts,value->contracts_count,value->tables,value->tables_count").Where("key=''import_info'' and member_id=#key_id#").Vars(hash00002)

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
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
	(22, 'import_upload', 'Div(content-wrapper){
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
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
	(23, 'langres_add', 'If(GetVar(application_id)){}.Else{
	DBFind(buffer_data).Columns("value->app_id").Where("key=''export'' and member_id=#key_id#").Vars(buffer)
	If(#buffer_value_app_id#>0){
		SetVar(application_id,#buffer_value_app_id#)
	}.Else{
		SetVar(application_id,1)
	}
}
If(GetVar(name)){}.Else{
	SetVar(name,)
}
Div(content-wrapper){
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
					Input(Name:Name, Value:#name#)
				}
			}
			Div(row text-muted){
				Div(col-md-1 mt-lg){
					Label(){Locale}
				}
				Div(col-md-10 mt-lg){
					Label(){Value}
				}
				Div(col-md-1 mt-lg){
					Label(){Action}
				}
			}
			If(GetVar(cs)==""){
				SetVar(cs,0)
			}
			If(#del# == 1){
				SetVar(cs,Calculate(#cs# - 1))
			}.Else{
				SetVar(cs,Calculate(#cs# + 1))
			}
			Range(params_range, 0, #cs#)
			ForList(Source: params_range){
				Div(row mt-sm){
					Div(col-md-1){
						Input(Name:LocaleArr)
					}.Style(input {padding: 6px;text-align:center;})
					Div(col-md-10){
						Input(Name:ValueArr)
					}
					Div(col-md-1){
						If(And(#cs#==#params_range_index#,#cs#>1)){
							Button(Class:fa fa-trash btn btn-default, PageParams: "cs=#cs#,del=1,application_id=#application_id#", Page: langres_add)
						}
					}
				}
			}
			Div(row){
				Div(col-md-12 mt-lg){
					LinkPage(Body: "Add localization", Page: langres_add, PageParams:"cs=#cs#,application_id=#application_id#")
				}
			}
		}
		Div(panel-footer){
			Button(Body: "Save", Class: btn btn-primary, Contract: @1NewLangJoint, Page: app_langres, Params: "ApplicationId=#application_id#")
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
	(24, 'langres_edit', 'Div(content-wrapper){
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

			JsonToSource(pv, #pre_res#)
			ForList(Source: pv, Index:s_ind){
				SetVar(max_sec, #s_ind#)
			}
			If(GetVar(cs)==""){
				SetVar(cs, #max_sec#)
			}
			If(Or(#del_flag#==1,#del_data#>0)){
				SetVar(cs, Calculate(Exp:#cs#-1, Type: int))
			}
			
			SetVar(next_sec, Calculate(Exp:#cs#+1, Type: int))
			SetVar(data_sec, Calculate(Exp:#cs#-#max_sec#, Type: int))

			ForList(Source: pv, Index:s_ind){
				If(#s_ind#>#cs#){
				}.Else{
					Div(row){
						Div(col-md-1 mt-sm){
							Input(Name: LocaleArr, Value: #key#)
						}.Style(input {padding: 6px;text-align:center;})
						Div(col-md-10 mt-sm){
							Input(Name: ValueArr, Value: #value#)
						}
						Div(col-md-1 mt-sm){
							If(And(#s_ind#>1,#s_ind#==#cs#)){
								Button(Body: Em(Class: fa fa-trash), Class: btn btn-default, PageParams: "lang_id=#lang_id#,cs=#cs#,del_data=#s_ind#", Page: langres_edit)
							}
						}
					}
				}
			}
			Range(params_range, #max_sec#, #cs#)
			ForList(Source: params_range, Index:s_ind){
				Div(row){
					Div(col-md-1 mt-sm){
						Input(Name:LocaleArr)
					}.Style(input {padding: 6px;text-align:center;})
					Div(col-md-10 mt-sm){
						Input(Name:ValueArr)
					}
					Div(col-md-1 mt-sm){
						If(#s_ind#==#data_sec#){
							Button(Body: Em(Class: fa fa-trash), Class: btn btn-default, PageParams: "lang_id=#lang_id#,cs=#cs#,del_flag=1", Page: langres_edit)
						}
					}
				}
			}
			Div(row){
				Div(col-md-12 mt-lg){
					LinkPage(Body: "Add localization", Page: langres_edit, PageParams: "lang_id=#lang_id#,cs=#next_sec#")
				}
			}
		}
		Div(panel-footer){
			Button(Body: "Save", Class: btn btn-primary, Contract: @1EditLangJoint, Params: "Id=#lang_id#", Page: app_langres)
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
	(25, 'menus_list', 'Div(fullscreen){
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
	(26, 'params_edit', 'Div(content-wrapper){
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
				Button(Class: btn btn-primary, Body: "Save", Contract: @1EditParameter, Params: "Id=#param_id#,Value=Val(value),Conditions=Val(conditions)", Page: params_list)
			}.Else{
				Button(Class: btn btn-primary, Body: "Save", Contract: @1NewParameter, Params: "Name=Val(name),Value=Val(value),Conditions=Val(conditions)", Page: params_list)
			}
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
	(27, 'params_list', 'Div(fullscreen){
	SetTitle("Ecosystem parameters")
	AddToolButton(Title: "Manage stylesheet", Page: params_edit, Icon: icon-picture, PageParams:"stylesheet=stylesheet")
	AddToolButton(Title: "Create", Page: params_edit, Icon: icon-plus)
	Div(breadcrumb){
		Span(Class: text-muted, Body: "This section is used to configure stored reusable parameters")
	}

	DBFind(parameters, src_appparameters).Order("id").Custom(custom_actions){
		LinkPage(Body: Em(Class: fa fa-edit), Class: text-primary h4, Page: params_edit, PageParams: "id=#id#")
	}

	Table(src_appparameters, "ID=id,Name=name,Value=value,Conditions=conditions,=custom_actions").Style(
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
	(28, 'properties_edit', 'Div(Class: content-wrapper){
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
				Button(Body: "Save", Class: btn btn-primary, Page: app_pages, Contract: @1EditPage, Params: "Menu=Val(Menu),Conditions=Val(Conditions),Id=#edit_property_id#")
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
						SetVar(address_item_wallet_id, Address(#item_wallet_id#))
						Input(Name: Wallet,Value: #address_item_wallet_id#)
					}
					Div(col-md-2){
						If(#item_active# == 0){
							Button(Body: "Bind", Class: btn btn-primary btn-block, Contract: @1ActivateContract, Params: "Id=#edit_property_id#", Page:app_contracts)
						}.Else{
							Button(Body: "Unbind", Class: btn btn-primary btn-block, Contract: @1DeactivateContract, Params: "Id=#edit_property_id#", Page:properties_edit, PageParams: "edit_property_id=#edit_property_id#,type=#type#")
						}
					}
				}
			}
			Div(form-group){
				Button(Body: "Save", Class: btn btn-primary, Page: app_contracts, Contract: @1EditContract, Params: "Conditions=Val(Conditions),WalletId=Val(Wallet),Id=#edit_property_id#")
			}
		}
		If(#type# == block){
			Div(form-group){
				Label("Change conditions")
				Input(Name: Conditions, Value: #item_conditions#)
			}
			Div(form-group){
				Button(Body: "Save", Class: btn btn-primary, Page: app_blocks, Contract: @1EditBlock, Params: "Conditions=Val(Conditions),Id=#edit_property_id#")
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
				Button(Body: "Save", Class: btn btn-primary, Page: menus_list, Contract: @1EditMenu, Params: "Conditions=Val(Conditions),Id=#edit_property_id#,NameTitle=Val(Title)")
			}
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
	(29, 'table_create', 'Div(content-wrapper){
	SetTitle("Create table")
	Div(breadcrumb){
		Div(){
			LinkPage("Tables", app_tables)
			Span(/).Style(margin-right: 10px; margin-left: 10px;)
			Span("Create", text-muted)
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
	Form(){
		Div(panel panel-default){
			Div(panel-body){
				Div(row){
					Div(col-md-12){
						Label("Name")
						Input(Name:Name)
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
				If(GetVar(cs)==""){
					SetVar(cs, Calculate( Exp: 0, Type: int))
				}
				If(#del# == 1){
					SetVar(cs, Calculate( Exp: #cs# - 1, Type: int))
				}.Else{
					SetVar(cs, Calculate( Exp: #cs# + 1, Type: int))
				}
				Range(params_range, 0, #cs#)
				ForList(Source: params_range){
					Div(row){
						Div(col-md-4 mt-sm){
							Input(Name:ColumnsArr)
						}
						Div(col-md-7 mt-sm){
							Select(Name: TypesArr, Source: src_type, NameColumn: name, ValueColumn: type)
						}
						Div(col-md-1 mt-sm){
							If(And(#cs#==#params_range_index#, #cs# > 1)){
								Button(Body: Em(Class: fa fa-trash), Class: btn btn-default, PageParams: "cs=#cs#,del=1,application_id=#application_id#", Page: table_create)
							}
						}
					}
				}			
			}
			Div(panel-footer){
				Button(Body: "Add column", Class: btn btn-primary, Page: table_create, PageParams: "cs=#cs#,application_id=#application_id#")
			}
		}
		Div(row){
			Div(col-md-6){
				Div(panel panel-default){
					Div(panel-heading, Body: "Write permissions")
					Div(panel-body){
						Div(form-group){
							Label(Insert)
							Input(Name: InsertPerm, Value: ContractConditions("MainCondition"))
						}
						Div(form-group){
							Label(Update)
							Input(Name: UpdatePerm, Value: ContractConditions("MainCondition"))
						}
						Div(form-group){
							Label(New column)
							Input(Name: NewColumnPerm, Value: ContractConditions("MainCondition"))
						}
					}
					Div(panel-footer){
						Button(Body: "Save", Class: btn btn-primary, Contract: @1NewTableJoint, Page: app_tables, Params: "ApplicationId=#application_id#")
					}
				}
			}
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
	(30, 'table_edit', 'Div(content-wrapper){
	DBFind(tables, src_mem).Columns("id,name,columns,conditions,permissions->insert,permissions->update,permissions->new_column").Vars(pre).WhereId(#tabl_id#)
	
	SetTitle("Tables")
	Div(breadcrumb){
		Div(){
			LinkPage("Tables", app_tables)
			Span(/).Style(margin-right: 10px; margin-left: 10px;)
			LinkPage(#pre_name#, table_view,, "tabl_id=#tabl_id#")
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
				JsonToSource(src_columns, #pre_columns#)
				ForList(src_columns){
					Div(list-group-item){
						Div(row){
							Div(col-md-3 h5){
								Span(#key#)
							}
							Div(col-md-2 h5){
								SetVar(col_type,GetColumnType(#pre_name#, #key#))
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
							Input(Name: InsertPerm, Type: text, Value: #pre_permissions_insert#)
						}
						Div(form-group){
							Label("Update")
							Input(Name: UpdatePerm, Type: text, Value: #pre_permissions_update#)
						}
						Div(form-group){
							Label("New column")
							Input(Name: NewColumnPerm, Type: text, Value: #pre_permissions_new_column#)
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
							Input(Name: Insert_condition, Disabled: true, Type: text, Value: #pre_conditions#)
						}
					}
				}
			}
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
	(31, 'table_view', 'Div(content-wrapper){
	DBFind(tables).WhereId(#tabl_id#).Columns("id,name").Vars(pre)

	SetTitle("Tables")
	Div(breadcrumb){
		LinkPage("Tables", app_tables)
		Span(/).Style(margin-right: 10px; margin-left: 10px;)
		Span(#pre_name#, text-muted)
		Span(/).Style(margin-right: 10px; margin-left: 10px;)
		LinkPage(Body:"Edit", Page: table_edit, PageParams: "tabl_id=#tabl_id#")
	}

	DBFind(#pre_name#).Count(count)
	If(#page#>0){
		SetVar(prev_page,Calculate(#page#-1)
	}.Else{
		SetVar(page,0).(prev_page,0)
	}
	SetVar(per_page,25).(off,Calculate(#page#*#per_page#)).(last_page,Calculate(#count#/#per_page#)).(next_page,#last_page#)
	If(#count#>Calculate(#off#+#per_page#)){
		SetVar(next_page,Calculate(#page#+1)
	}
	Div(button-group){
		If(#page#>0){
			Button(Body:"1", Class:btn btn-default, Page:table_view, PageParams: "tabl_id=#tabl_id#,page=0")
		}.Else{
			Button(Body:"1", Class:btn btn-default disabled)
		}
		If(#page#>1){
			Button(Body:Calculate(#prev_page#+1), Class:btn btn-default, Page:table_view, PageParams: "tabl_id=#tabl_id#,page=#prev_page#")
		}
		If(And(#page#>0,#page#<#last_page#)){
			Button(Body:Calculate(#page#+1), Class:btn btn-default disabled)
		}
		If(#next_page#<#last_page#){
			Button(Body:Calculate(#next_page#+1), Class:btn btn-default, Page:table_view, PageParams: "tabl_id=#tabl_id#,page=#next_page#")
		}
		If(#page#<#last_page#){
			Button(Body:Calculate(#last_page#+1), Class:btn btn-default, Page:table_view, PageParams: "tabl_id=#tabl_id#,page=#last_page#")
		}.ElseIf(#last_page#>0){
			Button(Body:Calculate(#last_page#+1), Class:btn btn-default disabled)
		}
	}
	Div(panel panel-default){
		Div(panel-body){
			Div(table-responsive){
				DBFind(#pre_name#, src_mem).Offset(#off#).Order(id)
				Table(src_mem)
			}
		}
	}
}', 'admin_menu', 'ContractAccess("@1EditPage")'),
	(32, 'admin_index', '', 'admin_menu', true),
	(33, 'notifications', $$DBFind(Name: notifications, Source: notifications_members).Columns("id,page_name,notification->icon,notification->header,notification->body").Where("closed=0 and notification->type='1' and recipient->member_id='#key_id#'")
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
