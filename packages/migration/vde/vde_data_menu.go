package vde

var menuDataSQL = `
INSERT INTO "%[1]d_menu" ("id","name","title","value","conditions") VALUES('2','admin_menu','Admin menu','MenuItem(
    Icon: "icon-screen-desktop",
    Page: "interface",
    Vde: "true",
    Title: "Interface"
)
MenuItem(
    Icon: "icon-docs",
    Page: "tables",
    Vde: "true",
    Title: "Tables"
)
MenuItem(
    Icon: "icon-briefcase",
    Page: "contracts",
    Vde: "true",
    Title: "Smart Contracts"
)
MenuItem(
    Icon: "icon-settings",
    Page: "parameters",
    Vde: "true",
    Title: "Ecosystem parameters"
)
MenuItem(
    Icon: "icon-globe",
    Page: "languages",
    Vde: "true",
    Title: "Language resources"
)
MenuItem(
    Icon: "icon-cloud-upload",
    Page: "import",
    Vde: "true",
    Title: "Import"
)
MenuItem(
    Icon: "icon-cloud-download",
    Page: "export",
    Vde: "true",
    Title: "Export"
)','true');`
