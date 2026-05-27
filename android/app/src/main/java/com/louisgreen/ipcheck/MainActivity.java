package com.louisgreen.ipcheck;

import android.annotation.SuppressLint;
import android.app.Activity;
import android.os.Bundle;
import android.webkit.JavascriptInterface;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;

import com.louisgreen.ipcheck.mobilecore.Mobilecore;

public class MainActivity extends Activity {
    @SuppressLint("SetJavaScriptEnabled")
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        WebView webView = new WebView(this);
        WebSettings settings = webView.getSettings();
        settings.setJavaScriptEnabled(true);
        settings.setDomStorageEnabled(true);
        settings.setAllowFileAccess(true);
        settings.setAllowContentAccess(false);
        webView.setWebViewClient(new WebViewClient());
        webView.addJavascriptInterface(new Bridge(), "IpCheckAndroid");
        webView.loadUrl("file:///android_asset/index.html");
        setContentView(webView);
    }

    public static class Bridge {
        @JavascriptInterface
        public String resolve(String requestJson) {
            return Mobilecore.resolveJSON(requestJson);
        }

        @JavascriptInterface
        public String health() {
            return Mobilecore.healthJSON();
        }
    }
}
