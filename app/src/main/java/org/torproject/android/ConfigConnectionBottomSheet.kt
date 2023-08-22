package org.torproject.android

import android.content.Context
import android.os.Build
import android.os.Bundle
import android.telephony.TelephonyManager
import android.util.Log
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Button
import android.widget.CompoundButton
import android.widget.RadioButton
import org.torproject.android.circumvention.Bridges
import org.torproject.android.circumvention.CircumventionApiManager
import org.torproject.android.circumvention.SettingsRequest
import org.torproject.android.service.util.Prefs
import java.util.*

class ConfigConnectionBottomSheet(private val callbacks: ConnectionHelperCallbacks) : OrbotBottomSheetDialogFragment() {

    private lateinit var rbDirect: RadioButton
    private lateinit var rbSnowflake: RadioButton
  //  private lateinit var rbSnowflakeAmp: RadioButton
    private lateinit var rbRequestBridge: RadioButton
    private lateinit var rbCustom: RadioButton
    // a choice to subscribe to new bridges via push notification
    private lateinit var rbPushBridge: RadioButton

    private lateinit var btnAction: Button
    private lateinit var btnAskTor: Button

    override fun onCreateView(
        inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?): View? {
        val v = inflater.inflate(R.layout.config_connection_bottom_sheet, container, false)

        rbDirect = v.findViewById(R.id.rbDirect)
        rbSnowflake = v.findViewById(R.id.rbSnowflake)
    //    rbSnowflakeAmp = v.findViewById(R.id.rbSnowflakeAmp)
        rbRequestBridge = v.findViewById(R.id.rbRequest)
        rbCustom = v.findViewById(R.id.rbCustom)
        rbPushBridge = v.findViewById(R.id.rbPush)

        val tvDirectSubtitle = v.findViewById<View>(R.id.tvDirectSubtitle)
        val tvSnowflakeSubtitle = v.findViewById<View>(R.id.tvSnowflakeSubtitle)
    //   val tvSnowflakeAmpSubtitle = v.findViewById<View>(R.id.tvSnowflakeAmpSubtitle)
        val tvRequestSubtitle = v.findViewById<View>(R.id.tvRequestSubtitle)
        val tvCustomSubtitle = v.findViewById<View>(R.id.tvCustomSubtitle)
        val tvPushSubtitle = v.findViewById<View>(R.id.tvPushSubtitle)

        val radios = arrayListOf(rbDirect, rbSnowflake, rbRequestBridge, rbCustom, rbPushBridge)
        val radioSubtitleMap = mapOf<CompoundButton, View>(
            rbDirect to tvDirectSubtitle, rbSnowflake to tvSnowflakeSubtitle,
            rbRequestBridge to tvRequestSubtitle, rbCustom to tvCustomSubtitle,
            rbPushBridge to tvPushSubtitle)
        val allSubtitles = arrayListOf(tvDirectSubtitle, tvSnowflakeSubtitle,
            tvRequestSubtitle, tvCustomSubtitle, tvPushSubtitle)
        btnAction = v.findViewById(R.id.btnAction)
        btnAskTor = v.findViewById(R.id.btnAskTor)

        btnAskTor.setOnClickListener {
            askTor()
        }

        // setup containers so radio buttons can be checked if labels are clicked on
     //   v.findViewById<View>(R.id.smartContainer).setOnClickListener {rbSmart.isChecked = true}
        v.findViewById<View>(R.id.directContainer).setOnClickListener {rbDirect.isChecked = true}
        v.findViewById<View>(R.id.snowflakeContainer).setOnClickListener {rbSnowflake.isChecked = true}
      //  v.findViewById<View>(R.id.snowflakeAmpContainer).setOnClickListener {rbSnowflakeAmp.isChecked = true}
        v.findViewById<View>(R.id.requestContainer).setOnClickListener {rbRequestBridge.isChecked = true}
        v.findViewById<View>(R.id.customContainer).setOnClickListener {rbCustom.isChecked = true}
        v.findViewById<View>(R.id.pushContainer).setOnClickListener {rbPushBridge.isChecked = true}
        v.findViewById<View>(R.id.tvCancel).setOnClickListener { dismiss() }

        rbDirect.setOnCheckedChangeListener { buttonView, isChecked ->
            if (isChecked) {
                nestedRadioButtonKludgeFunction(buttonView as RadioButton, radios)
                radioSubtitleMap[buttonView]?.let { onlyShowActiveSubtitle(it, allSubtitles) }
            }
        }
        rbSnowflake.setOnCheckedChangeListener { buttonView, isChecked ->
            if (isChecked) {
                nestedRadioButtonKludgeFunction(buttonView as RadioButton, radios)
                radioSubtitleMap[buttonView]?.let { onlyShowActiveSubtitle(it, allSubtitles) }
            }
        }
        /**
        rbSnowflakeAmp.setOnCheckedChangeListener { buttonView, isChecked ->
            if (isChecked) {
                nestedRadioButtonKludgeFunction(buttonView as RadioButton, radios)
                radioSubtitleMap[buttonView]?.let { onlyShowActiveSubtitle(it, allSubtitles) }
            }
        }**/
        rbRequestBridge.setOnCheckedChangeListener { buttonView, isChecked ->
            if (isChecked) {
                nestedRadioButtonKludgeFunction(buttonView as RadioButton, radios)
                radioSubtitleMap[buttonView]?.let { onlyShowActiveSubtitle(it, allSubtitles) }
                btnAction.text = getString(R.string.next)
            } else {
                btnAction.text = getString(R.string.connect)
            }
        }
        rbCustom.setOnCheckedChangeListener { buttonView, isChecked ->
            if (isChecked) {
                nestedRadioButtonKludgeFunction(buttonView as RadioButton, radios)
                radioSubtitleMap[buttonView]?.let { onlyShowActiveSubtitle(it, allSubtitles) }
                btnAction.text = getString(R.string.next)
            } else {
                btnAction.text = getString(R.string.connect)
            }
        }
        rbPushBridge.setOnCheckedChangeListener { buttonView, isChecked ->
            if (isChecked) {
                nestedRadioButtonKludgeFunction(buttonView as RadioButton, radios)
                radioSubtitleMap[buttonView]?.let { onlyShowActiveSubtitle(it, allSubtitles) }
                btnAction.text = getString(R.string.next)
            } else {
                btnAction.text = getString(R.string.connect)
            }
        }

        selectRadioButtonFromPreference()

        btnAction.setOnClickListener {
            if (rbRequestBridge.isChecked) {
                MoatBottomSheet(object : ConnectionHelperCallbacks {
                    override fun tryConnecting() {
                        Prefs.putConnectionPathway(Prefs.PATHWAY_CUSTOM)
                        rbCustom.isChecked = true
                        dismiss()
                        callbacks.tryConnecting()
                    }
                }).show(requireActivity().supportFragmentManager, MoatBottomSheet.TAG)
            }
            else if (rbDirect.isChecked) {
                Prefs.putConnectionPathway(Prefs.PATHWAY_DIRECT)
                closeAndConnect()
            } else if (rbSnowflake.isChecked) {
                Prefs.putConnectionPathway(Prefs.PATHWAY_SNOWFLAKE)
                closeAndConnect()
            } /**else if (rbSnowflakeAmp.isChecked) {
                Prefs.putConnectionPathway(Prefs.PATHWAY_SNOWFLAKE_AMP)
                closeAndConnect()
            } **/else if (rbCustom.isChecked) {
                CustomBridgeBottomSheet(callbacks).show(requireActivity().supportFragmentManager, CustomBridgeBottomSheet.TAG)
            } else if (rbPushBridge.isChecked) {
                PushBridgeBottomSheet(callbacks).show(requireActivity().supportFragmentManager, PushBridgeBottomSheet.TAG)
            }
        }

        return v
    }

    private fun closeAndConnect() {
        closeAllSheets()
        callbacks.tryConnecting()
    }

    // it's 2022 and android makes you do ungodly things for mere radio button functionality
    private fun nestedRadioButtonKludgeFunction(rb: RadioButton, all: List<RadioButton>) =
        all.forEach { if (it != rb) it.isChecked = false }

    private fun onlyShowActiveSubtitle(showMe: View, all: List<View>) = all.forEach {
            if (it == showMe) it.visibility = View.VISIBLE
            else it.visibility = View.GONE
        }

    private fun selectRadioButtonFromPreference() {
        val pref = Prefs.getConnectionPathway()
        if (pref.equals(Prefs.PATHWAY_CUSTOM)) rbCustom.isChecked = true
        if (pref.equals(Prefs.PATHWAY_SNOWFLAKE)) rbSnowflake.isChecked = true
       // if (pref.equals(Prefs.PATHWAY_SNOWFLAKE_AMP)) rbSnowflakeAmp.isChecked = true
        if (pref.equals(Prefs.PATHWAY_DIRECT)) rbDirect.isChecked = true
    }

    private var circumventionApiBridges: List<Bridges?>? = null
    private var circumventionApiIndex = 0

    private fun askTor () {

        val dLeft = activity?.getDrawable(R.drawable.ic_faq)
        btnAskTor.text = getString(R.string.asking)
        btnAskTor.setCompoundDrawablesWithIntrinsicBounds(dLeft, null, null, null)

        val countryCodeValue: String? = getDeviceCountryCode(requireContext())
        Log.d("bim", "The country code is $countryCodeValue")

        CircumventionApiManager().getSettings(SettingsRequest(countryCodeValue), {
            it?.let {
                circumventionApiBridges = it.settings
                if (circumventionApiBridges == null) {
                    Log.d("bim", "settings is null, we can assume a direct connect is fine ")
                    rbDirect.isChecked = true;

                } else {

                    Log.d("bim", "settings is $circumventionApiBridges")
                    circumventionApiBridges?.forEach { b->
                        Log.d("bim", "BRIDGE $b")
                    }

                    //got bridges, let's set them
                    setPreferenceForSmartConnect()
                }
            }
        }, {
            // TODO what happens to the app in this case?!
            Log.e("bim", "Couldn't hit circumvention API... $it")
        })
    }

    private fun getDeviceCountryCode(context: Context): String? {
        var countryCode: String?

        // Try to get country code from TelephonyManager service
        val tm = context.getSystemService(Context.TELEPHONY_SERVICE) as TelephonyManager

        // Query first getSimCountryIso()
        countryCode = tm.simCountryIso
        if (countryCode != null && countryCode.length == 2)
            return countryCode.lowercase(Locale.getDefault())

        countryCode = tm.networkCountryIso
        if (countryCode != null && countryCode.length == 2)
                  return countryCode.lowercase(Locale.getDefault())


        // If network country not available (tablets maybe), get country code from Locale class
        countryCode = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N) {
            context.resources.configuration.locales[0].country
        } else {
            context.resources.configuration.locale.country
        }

        return if (countryCode != null && countryCode.length == 2)
            countryCode.lowercase(Locale.getDefault()) else "us"

    }

    private fun setPreferenceForSmartConnect() {

        val dLeft = activity?.getDrawable(R.drawable.ic_green_check)
        btnAskTor.setCompoundDrawablesWithIntrinsicBounds(dLeft, null, null, null)

        circumventionApiBridges?.let {
            if (it.size == circumventionApiIndex) {
                circumventionApiBridges = null
                circumventionApiIndex = 0
                rbDirect.isChecked = true
                btnAskTor.text = getString(R.string.connection_direct)

                Log.d("bim", "smart connect: Direct is chosen")

                return
            }
            val b = it[circumventionApiIndex]!!.bridges
            if (b.type == CircumventionApiManager.BRIDGE_TYPE_SNOWFLAKE) {
                Prefs.putConnectionPathway(Prefs.PATHWAY_SNOWFLAKE)
                rbSnowflake.isChecked = true
                btnAskTor.text = getString(R.string.connection_snowflake)

                Log.d("bim", "smart connect: Snowflake is chosen")

            } else if (b.type == CircumventionApiManager.BRIDGE_TYPE_OBFS4) {

                rbCustom.isChecked = true
                btnAskTor.text = getString(R.string.connection_custom)

                var bridgeStrings = ""
                b.bridge_strings!!.forEach { bridgeString ->
                    bridgeStrings += "$bridgeString\n"
                }
                Prefs.setBridgesList(bridgeStrings)
                Prefs.putConnectionPathway(Prefs.PATHWAY_CUSTOM)

                Log.d("bim", "smart connect: Custom is chosen with bridges")
                Log.d("bim", "smart connect: bridgeStrings: $bridgeStrings")

            }
            else
            {
                rbDirect.isChecked = true

                Log.d("bim", "smart connect: Falling back to Direct")
            }

            circumventionApiIndex += 1
        }
    }


}