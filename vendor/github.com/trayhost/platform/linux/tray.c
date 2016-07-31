#ifndef NATIVE_C
#define NATIVE_C

#include <stdio.h>
#include <unistd.h>
#include <gtk/gtk.h>
#include <gio/gio.h>
#include <gdk-pixbuf/gdk-pixbuf.h>

static const char *menu_title = NULL;
static const char *url = NULL;
static GtkWidget *openMenuItem = NULL;
static GtkWidget *copyMenuItem = NULL;

static void handle_open(GtkStatusIcon *status_icon, gpointer user_data)
{
    pid_t pid = fork();
    if (pid == 0)
    {
        execlp("xdg-open", "xdg-open", url, (char*)NULL);
    }
}

static void handle_copy_to_clipboard(GtkStatusIcon *status_icon, gpointer user_data)
{
    GtkClipboard* clipboard = gtk_clipboard_get(GDK_SELECTION_CLIPBOARD);
    gtk_clipboard_set_text(clipboard, url, -1);
    gtk_clipboard_store(clipboard);
}

static void tray_exit(GtkMenuItem *item, gpointer user_data) 
{
    gtk_main_quit();
}

static void tray_icon_on_menu(GtkStatusIcon *status_icon, guint button, guint activate_time, gpointer user_data)
{
    GtkWidget *titleMenuItem = gtk_menu_item_new_with_label(menu_title);
    gtk_widget_set_sensitive(titleMenuItem, FALSE);
    openMenuItem = gtk_menu_item_new_with_label("Open");
    if (!url) {
        gtk_widget_set_sensitive(openMenuItem, FALSE);
        gtk_widget_set_sensitive(copyMenuItem, FALSE);
    }
    GtkWidget *exitMenuItem = gtk_menu_item_new_with_label("Exit");
    GtkWidget *menu = gtk_menu_new();

    g_signal_connect(G_OBJECT(openMenuItem), "activate", G_CALLBACK(handle_open), NULL);
    g_signal_connect(G_OBJECT(exitMenuItem), "activate", G_CALLBACK(tray_exit), NULL);

    gtk_menu_shell_append(GTK_MENU_SHELL(menu), openMenuItem);
    gtk_menu_shell_append(GTK_MENU_SHELL(menu), gtk_separator_menu_item_new());
    gtk_menu_shell_append(GTK_MENU_SHELL(menu), exitMenuItem);
    gtk_widget_show_all(menu);

    gtk_menu_popup(GTK_MENU(menu), NULL, NULL, NULL, NULL, 0, gtk_get_current_event_time());
}

static GtkStatusIcon *create_tray_icon(unsigned char *imageData, unsigned int imageDataLen) {
    GtkStatusIcon *tray_icon;
    GError *error = NULL;
    GInputStream *stream = g_memory_input_stream_new_from_data(imageData, imageDataLen, NULL);
    GdkPixbuf *pixbuf = gdk_pixbuf_new_from_stream(stream, NULL, &error);
    if (error)
        fprintf(stderr, "Unable to create PixBuf: %s\n", error->message);

    tray_icon = gtk_status_icon_new_from_pixbuf(pixbuf);
    g_signal_connect(G_OBJECT(tray_icon), "activate", G_CALLBACK(handle_open), NULL);
    g_signal_connect(G_OBJECT(tray_icon), "popup-menu", G_CALLBACK(tray_icon_on_menu), NULL);
    gtk_status_icon_set_tooltip_text(tray_icon, menu_title);
    gtk_status_icon_set_visible(tray_icon, TRUE);

    return tray_icon;
}

void set_url(const char* theUrl)
{
    url = theUrl;
    if (openMenuItem)
        gtk_widget_set_sensitive(openMenuItem, TRUE);
    if (copyMenuItem)
        gtk_widget_set_sensitive(copyMenuItem, TRUE);
}

void native_loop(const char* title, unsigned char *imageData, unsigned int imageDataLen) 
{
    int argc = 0;
    char *argv[] = { "" };
    menu_title = title;

    gtk_init(&argc, (char***)&argv);
    create_tray_icon(imageData, imageDataLen);
    gtk_main();
}

#endif // NATIVE_C
