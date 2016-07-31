package org.golang.app;

import android.app.Notification;
import android.app.PendingIntent;
import android.app.Service ;
import android.content.Intent;
import android.os.IBinder;
import android.util.Log;

import java.lang.System;
import java.net.Socket;
import android.os.SystemClock;
import java.net.HttpURLConnection;
import java.net.URL;
import java.io.OutputStream;


public class MyService extends Service  {

    @Override
    public IBinder onBind(Intent intent) {
        Log.d("JavaGo", "MyService onBind");
        return null;
    }

    public int onStartCommand(Intent intent, int flags, int startId) {
        Log.d("JavaGo", "MyService onStartCommand");

        Runnable r = new Runnable() {
            public void run() {
                SystemClock.sleep(500);
                Intent i = new Intent(MyService.this, WViewActivity.class);
                i.addFlags(i.FLAG_ACTIVITY_NEW_TASK);
                startActivity(i);
            }
        };
        Thread t = new Thread(r);
        t.start();

        return super.onStartCommand(intent, flags, startId);
    }

    public void onStart() {
        Log.d("JavaGo", "MyService onStart");
    }

    public void onDestroy() {
        Log.d("JavaGo", "MyService onDestroy");
        super.onDestroy();
    }

    @Override
    public void onCreate() {
        Log.d("JavaGo", "MyService onCreate");
        super.onCreate();

        ShortcutIcon();

        sendNotif();


        //Runnable r = new Runnable() {
        //    public void run() {
                GoNativeActivity.load();
        //    }
        //};
        //Thread t = new Thread(r);
        //t.start();

        Intent dialogIntent = new Intent(this, GoNativeActivity.class);
        dialogIntent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
        startActivity(dialogIntent);


/*
        try {
            Intent intent1 = new Intent(Intent.ACTION_VIEW);
            Uri data = Uri.parse("http://localhost:8089");
            intent1.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
            intent1.setData(data);
            startActivity(intent1);
        } catch (Exception e) {
            Log.e("JavaGo", "http://localhost:8089 failed", e);
        }*/
    }

    @SuppressWarnings("deprecation")
    void sendNotif() {

        Log.d("JavaGo", "MyService sendNotif");
//        Notification notif = new Notification(R.drawable.icon, "Dcoin", System.currentTimeMillis());
        Notification.Builder builder = new Notification.Builder(this);

        Intent intent = new Intent(this, MainActivity.class);
        PendingIntent pIntent = PendingIntent.getActivity(this, 0, intent, 0);

        builder.setAutoCancel(false);
        builder.setContentTitle("Dcoin");
        builder.setSmallIcon(R.drawable.icon);
        builder.setContentIntent(pIntent);
        builder.setOngoing(true);
        builder.setContentText(Long.toString(System.currentTimeMillis()));

//        builder.setNumber(System.currentTimeMillis());
        builder.build();

        Notification notif = builder.getNotification();

//        notif.setLatestEventInfo(this, "Dcoin", "Running", pIntent);

        startForeground(1, notif);

    }
    private void ShortcutIcon(){

        /*Log.d("JavaGo", "MyService ShortcutIcon");
        Intent shortcutIntent = new Intent(getApplicationContext(), MainActivity.class);
        shortcutIntent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
        shortcutIntent.addFlags(Intent.FLAG_ACTIVITY_CLEAR_TOP);

        Intent addIntent = new Intent();
        addIntent.putExtra(Intent.EXTRA_SHORTCUT_INTENT, shortcutIntent);
        addIntent.putExtra(Intent.EXTRA_SHORTCUT_NAME, "Dcoin");
        addIntent.putExtra(Intent.EXTRA_SHORTCUT_ICON_RESOURCE, Intent.ShortcutIconResource.fromContext(getApplicationContext(), R.drawable.icon));
        addIntent.setAction("com.android.launcher.action.INSTALL_SHORTCUT");
        getApplicationContext().sendBroadcast(addIntent);*/
    }

    public static boolean DcoinStarted(int port) {
        for (int i=0;i<35;i++) {
            try {
                URL url = new URL("http://localhost:8089/");
                HttpURLConnection connection = (HttpURLConnection)url.openConnection();
                Log.d("JavaGo", "0");
                connection.setRequestMethod("GET");
                Log.d("JavaGo", "1");
                connection.connect();
                Log.d("JavaGo", "2");
                Log.d("JavaGo", "huc.getResponseCode()"+connection.getResponseCode());
                if (connection.getResponseCode() == 200 ) {
                    return true;
                } else {
                    SystemClock.sleep(500);
                }
            } catch (Exception e) {
                Log.e("JavaGo", "http://localhost:8089 failed", e);
                SystemClock.sleep(500);
            }
            /*
            try (Socket ignored = new Socket("localhost", port)) {
                return true;
            } catch (Exception ignored) {
                SystemClock.sleep(500);
            }*/
        }
        return true;
    }

}
