package org.torproject.android

import android.app.NotificationChannel
import android.app.NotificationManager
import android.content.Context
import android.os.Build
import android.util.Log
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import com.google.firebase.messaging.FirebaseMessagingService
import com.google.firebase.messaging.RemoteMessage
import com.google.gson.Gson
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking
import okhttp3.Call
import okhttp3.Callback
import okhttp3.MediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody
import okhttp3.Response
import org.torproject.android.circumvention.CircumventionApiManager
import org.torproject.android.circumvention.SettingsResponse
import org.torproject.android.service.util.Prefs
import java.io.IOException

class MyFirebaseMessagingService : FirebaseMessagingService() {

    /**
     * Called if the FCM registration token is updated. This may occur if the security of
     * the previous token had been compromised. Note that this is called when the
     * FCM registration token is initially generated so this is where you would retrieve the token.
     */
    override fun onNewToken(token: String) {
        Log.d(TAG, "Refreshed token: $token")

        // If you want to send messages to this application instance or
        // manage this apps subscriptions on the server side, send the
        // FCM registration token to your app server.
        sendRegistrationToServer(token)
    }

    override fun onMessageReceived(remoteMessage: RemoteMessage) {
        // ...

        // TODO(developer): Handle FCM messages here.
        // Not getting messages here? See why this may be: https://goo.gl/39bRNJ
        Log.d(TAG, "From: ${remoteMessage.from}")

        // Check if message contains a data payload.
        if (remoteMessage.data.isNotEmpty()) {
            Log.d(TAG, "Message data payload: ${remoteMessage.data}")

            // parse remoteMessage.data
            val gson = Gson()
            val settingsResponse = gson.fromJson(remoteMessage.data.getOrDefault("payload", "{}"), SettingsResponse::class.java)

            var selectedMethod = "(no change)"

            // Save bridge in settings
            settingsResponse.settings?.let {
                if (it.size == circumventionApiIndex) {
                    // Don't do anything for now
                    // Prefs.putConnectionPathway(Prefs.PATHWAY_DIRECT)
                    Log.d(TAG, "push notif: Direct is chosen")
                    selectedMethod = "direct"

                    return
                }
                val b = it[circumventionApiIndex].bridges
                if (b.type == CircumventionApiManager.BRIDGE_TYPE_SNOWFLAKE) {
                    Prefs.putConnectionPathway(Prefs.PATHWAY_SNOWFLAKE)
                    Log.d(TAG, "push notif: Snowflake is chosen")
                    selectedMethod = "snowflake"
                } else if (b.type == CircumventionApiManager.BRIDGE_TYPE_OBFS4) {
                    var bridgeStrings = ""
                    b.bridge_strings!!.forEach { bridgeString ->
                        bridgeStrings += "$bridgeString\n"
                    }
                    Prefs.setBridgesList(bridgeStrings)
                    Prefs.putConnectionPathway(Prefs.PATHWAY_CUSTOM)

                    Log.d(TAG, "push notif: Custom is chosen with bridges")
                    Log.d(TAG, "push notif: bridgeStrings: $bridgeStrings")
                    selectedMethod = "obfs4 with pushed bridges"
                } else {
                    // Don't do anything for now
                    // Prefs.putConnectionPathway(Prefs.PATHWAY_DIRECT)

                    Log.d(TAG, "push notif: falling back to direct")
                    selectedMethod = "direct (fallback)"
                }
            }

            // (if available) use channel to notify the UI thread for connection, or display notification for user
            // TODO: can I use runBlocking instead of lifecycleScope.launch(Dispatchers.Main) here?
            runBlocking {
                launch {
                    val channel = waitingChannel
                    if (channel != null) {
                        Log.d(TAG, "channel send")
                        channel.send(selectedMethod)
                    } else {
                        Log.d(TAG, "display notification")
                        showNotification(applicationContext, NOTIFICATION_CHANNEL_ID, getString(R.string.bridges_updated), getString(R.string.restart_orbot_to_use_this_bridge_))
                    }
                }
            }


//            if (/* Check if data needs to be processed by long running job */ true) {
//                // For long-running tasks (10 seconds or more) use WorkManager.
//                scheduleJob()
//            } else {
//                // Handle message within 10 seconds
//                handleNow()
//            }
        }

        // Check if message contains a notification payload.
        remoteMessage.notification?.let {
            Log.d(TAG, "Message Notification Body: ${it.body}")
        }

        // Also if you intend on generating your own notifications as a result of a received FCM
        // message, here is where that should be initiated. See sendNotification method below.
    }

    companion object {
        private const val TAG = "push-notification"

        private const val NOTIFICATION_CHANNEL_ID = "orbot_channel_2"

        private val circumventionApiIndex = 0


        var waitingChannel: Channel<String>? = null

        fun sendRegistrationToServer(
            token: String,
            callbackIfSuccess: (() -> Unit)? = null,
            callbackIfFail: (() -> Unit)? = null
        ) {
            // this is the computer's address in Android Virtual Machine
            val url = "http://10.0.2.2:8888/fcm/register"
            val client = OkHttpClient()

            // or I could use kotlinx.serialization here
            // TODO: use actual country here
            val jsonString = """{ "token": "$token", "country": "cn" }"""

            val requestBody = RequestBody.create(MediaType.parse("application/json; charset=utf-8"), jsonString)
            val request = Request.Builder()
                .url(url)
                .post(requestBody)
                .build()

            client.newCall(request).enqueue(object : Callback {
                override fun onFailure(call: Call, e: IOException) {
                    Log.e(TAG, "Failed to send token to server: ${e.message}")
                    // TODO: handle the error, e.g. display user notification?
                    callbackIfFail?.invoke()
                }

                override fun onResponse(call: Call, response: Response) {
                    Log.d(TAG, "Token sent to server successfully")
                    response.body()?.charStream()?.readText().let {
                        if (it != null) {
                            Log.d(TAG, it)
                        } else {
                            Log.d(TAG, response.message())
                        }
                    }
                    callbackIfSuccess?.invoke()
                }
            })
        }

        fun showNotification(context: Context, channelId: String, title: String, content: String) {
            // Create a notification channel (required for Android Oreo and above)
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                val channel = NotificationChannel(channelId, "Channel Name", NotificationManager.IMPORTANCE_DEFAULT)
                val notificationManager = context.getSystemService(NOTIFICATION_SERVICE) as NotificationManager
                notificationManager.createNotificationChannel(channel)
            }

            // Create the notification
            val builder = NotificationCompat.Builder(context, channelId)
                .setSmallIcon(R.drawable.ic_onion)
                .setContentTitle(title)
                .setContentText(content)
                .setPriority(NotificationCompat.PRIORITY_DEFAULT)

            // Show the notification
            with(NotificationManagerCompat.from(context)) {
                notify(/*notificationId=*/0, builder.build())
            }
        }
    }
}
