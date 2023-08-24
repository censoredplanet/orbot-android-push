package org.torproject.android

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.util.Log
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Button
import android.widget.EditText
import androidx.activity.result.contract.ActivityResultContracts
import androidx.core.content.ContextCompat
import androidx.lifecycle.lifecycleScope
import com.google.android.gms.tasks.OnCompleteListener
import com.google.firebase.messaging.FirebaseMessaging
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.launch
import org.torproject.android.service.util.Prefs


// TODO: make this bottom sheet a place for getting permissions and confirming subscriptions.
class PushBridgeBottomSheet(private val callbacks: ConnectionHelperCallbacks): OrbotBottomSheetDialogFragment() {
    companion object {
        const val TAG = "PushBridgeBottomSheet"
        private const val bridgeStatement = "obfs4"
    }

    private lateinit var btnRequestPermission: Button

    // Declare the launcher at the top of your Activity/Fragment:
    private val requestPermissionLauncher = registerForActivityResult(
        ActivityResultContracts.RequestPermission(),
    ) { isGranted: Boolean ->
        val btnRequestPermission = requireView().findViewById<Button>(R.id.btnRequestPermission)
        if (isGranted) {
            // FCM SDK (and your app) can post notifications.
            btnRequestPermission.isEnabled = false
            btnRequestPermission.text = "Notifications Enabled ✔"

            FirebaseMessaging.getInstance().token.addOnCompleteListener(OnCompleteListener { task ->
                if (!task.isSuccessful) {
                    Log.w(TAG, "Fetching FCM registration token failed", task.exception)
                    return@OnCompleteListener
                }

                // Get new FCM registration token
                val token = task.result

                // Log and toast
                Log.d(TAG, token)
                etBridges.setText(token)
            })
        } else {
            // TODO: Inform user that that your app will not show notifications.
            btnRequestPermission.isEnabled = true
            btnRequestPermission.text = "Enable Notifications"
        }
    }

    private fun askNotificationPermission() {
        // This is only necessary for API level >= 33 (TIRAMISU)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            // Ye Shu question: in requireContext(), could the context be null? When?
            if (ContextCompat.checkSelfPermission(requireContext(), Manifest.permission.POST_NOTIFICATIONS) ==
                PackageManager.PERMISSION_GRANTED
            ) {
                // FCM SDK (and your app) can post notifications.
                btnRequestPermission.isEnabled = false
                btnRequestPermission.text = "Notifications Enabled ✔"

                FirebaseMessaging.getInstance().token.addOnCompleteListener(OnCompleteListener { task ->
                    if (!task.isSuccessful) {
                        Log.w(TAG, "Fetching FCM registration token failed", task.exception)
                        return@OnCompleteListener
                    }

                    // Get new FCM registration token
                    val token = task.result

                    // Log and toast
                    Log.d(TAG, token)
                    etBridges.setText(token)
                })
            } else if (shouldShowRequestPermissionRationale(Manifest.permission.POST_NOTIFICATIONS)) {
                // TODO: display an educational UI explaining to the user the features that will be enabled
                //       by them granting the POST_NOTIFICATION permission. This UI should provide the user
                //       "OK" and "No thanks" buttons. If the user selects "OK," directly request the permission.
                //       If the user selects "No thanks," allow the user to continue without notifications.
                btnRequestPermission.isEnabled = true
                btnRequestPermission.text = "Enable Notifications"

                // For now, directly ask for the permission
                requestPermissionLauncher.launch(Manifest.permission.POST_NOTIFICATIONS)
            } else {
                btnRequestPermission.isEnabled = true
                btnRequestPermission.text = "Enable Notifications"

                // Directly ask for the permission
                requestPermissionLauncher.launch(Manifest.permission.POST_NOTIFICATIONS)
            }
        }
        // TODO: what about older SDK? test this on an older Android Virtual Device
    }

    private lateinit var btnAction: Button
    private lateinit var etBridges: EditText

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View? {
        // to post a Runnable to the main thread's message queue, avoiding CalledFromWrongThreadException
        val mainHandler = Handler(Looper.getMainLooper())

        val v =  inflater.inflate(R.layout.push_bridge_bottom_sheet, container, false)
        v.findViewById<View>(R.id.tvCancel).setOnClickListener { dismiss() }

        btnRequestPermission = v.findViewById(R.id.btnRequestPermission)
        btnRequestPermission.setOnClickListener {
            askNotificationPermission()
        }

        // TODO: only allow connect when push notification has been received (use a blocking Go-like channel?)
        btnAction = v.findViewById(R.id.btnAction)
        btnAction.setOnClickListener {
            Prefs.setBridgesList(etBridges.text.toString())
            callbacks.tryConnecting()
            closeAllSheets()
        }
        btnAction.isEnabled = false

        // TODO: maybe use the textfield for out-of-band initialization?
        etBridges = v.findViewById(R.id.etBridges)
        configureMultilineEditTextScrollEvent(etBridges)
//        var bridges = Prefs.getBridgesList()
//        if (!bridges.contains(bridgeStatement)) bridges = ""
//        etBridges.setText(bridges)
//        updateUi()

        if (ContextCompat.checkSelfPermission(requireContext(), Manifest.permission.POST_NOTIFICATIONS) ==
            PackageManager.PERMISSION_GRANTED
        ) {
            // FCM SDK (and your app) can post notifications.
            btnRequestPermission.isEnabled = false
            btnRequestPermission.text = "Notifications Enabled ✔"

            FirebaseMessaging.getInstance().token.addOnCompleteListener(OnCompleteListener { task ->
                if (!task.isSuccessful) {
                    Log.w(TAG, "Fetching FCM registration token failed", task.exception)

                    // TODO: display a toast for error?
                    return@OnCompleteListener
                }

                // Get new FCM registration token
                val token = task.result

                // Log and toast
                Log.d(TAG, token)

                // Send the token to web server
                // TODO: check if has already been initialized?
                MyFirebaseMessagingService.sendRegistrationToServer(token, {
                    mainHandler.post { // Update UI elements here
                        etBridges.setText("Registered with server successfully. Awaiting bridges to be posted via push notification")

                        // use channel to wait for push messages. before then, user cannot proceed
                        MyFirebaseMessagingService.waitingChannel = Channel()
                        Log.d(TAG, "channel set. waiting " + MyFirebaseMessagingService.waitingChannel)
                        // TODO: why does runBlocking here result in Application Not Responding?
                        // How does switching to this fix the issue?
                        lifecycleScope.launch(Dispatchers.Main) {
                            launch {
                                val channel = MyFirebaseMessagingService.waitingChannel

                                if (channel == null) {
                                    // TODO: error. race condition? display error toast and go back?
                                    Log.w(TAG, "channel is null. race condition?")
                                    return@launch
                                }

                                Log.d(TAG, "channel wait to receive")
                                // TODO: add a timeout and prompt user to change method?
                                val selectedMethod = channel.receive()
                                Log.d(TAG, "channel receive successful")
                                channel.close()
                                MyFirebaseMessagingService.waitingChannel = null

                                // bridges will be set in MyFirebaseMessaingService
                                etBridges.setText("Registered with server successfully. Bridges received via push notification. You're all set! Instructed Method: " + selectedMethod)
                                btnAction.isEnabled = true
                            }
                        }

                    }
                }, {
                    mainHandler.post { // Update UI elements here
                        etBridges.setText("Cannot register with server. Please try registering out of band with the following token:\n$token")
                    }
                })
            })
        }

        Log.d(TAG, "initialized")
        return v
    }

    private fun updateUi() {
        btnAction.isEnabled =
            !(etBridges.text.isEmpty() || !etBridges.text.contains(bridgeStatement))
    }

}