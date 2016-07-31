package org.golang.app;

import android.app.Activity;
import android.os.Bundle;
import android.content.Intent;
import android.util.Log;
import android.net.Uri;
import android.os.Handler;
import android.os.SystemClock;
import java.util.concurrent.TimeUnit;
import android.widget.LinearLayout;
import android.util.Log;

public class MainActivity extends Activity {

    @Override
    protected void onCreate(Bundle savedInstanceState) {
		super.onCreate(savedInstanceState);
		Log.d("JavaGo", "MainActivity onCreate");
		Intent intent=new Intent("org.golang.app.MyService");
		this.startService(intent);
    }

    protected void onStart(Bundle savedInstanceState) {
		Log.d("JavaGo", "MainActivity onStart");
    }
}
