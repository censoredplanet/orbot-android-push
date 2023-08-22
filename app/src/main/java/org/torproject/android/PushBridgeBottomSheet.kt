package org.torproject.android

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import android.util.Log
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Button
import android.widget.EditText
import androidx.activity.result.contract.ActivityResultContracts
import androidx.core.content.ContextCompat
import com.google.android.gms.tasks.OnCompleteListener
import com.google.firebase.messaging.FirebaseMessaging
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

        // TODO: maybe use the textfield for out-of-band initialization?
        etBridges = v.findViewById(R.id.etBridges)
        configureMultilineEditTextScrollEvent(etBridges)
//        var bridges = Prefs.getBridgesList()
//        if (!bridges.contains(bridgeStatement)) bridges = ""
//        etBridges.setText(bridges)
//        updateUi()

        Log.d(TAG, "initialized")
        return v
    }

    private fun updateUi() {
        btnAction.isEnabled =
            !(etBridges.text.isEmpty() || !etBridges.text.contains(bridgeStatement))
    }

}