package kvm

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erikdubbelboer/gspt"
	"github.com/gwatts/rootcerts"
)

// var appCtx context.Context
var procPrefix string = "xkvm: [app]"

func setProcTitle(status string) {
	if status != "" {
		status = " " + status
	}
	title := fmt.Sprintf("%s%s", procPrefix, status)
	gspt.SetProcTitle(title)
}

func Main() {
	setProcTitle("starting")

	logger.Log().Msg("XKVM Starting Up")

	checkFailsafeReason()
	if failsafeModeActive {
		procPrefix = "xkvm: [app+failsafe]"
		logger.Warn().Str("reason", failsafeModeReason).Msg("failsafe mode activated")
	}

	LoadConfig()

	systemVersionLocal, appVersionLocal, err := GetLocalVersion()
	if err != nil {
		logger.Warn().Err(err).Msg("failed to get local version")
	}

	logger.Info().
		Interface("system_version", systemVersionLocal).
		Interface("app_version", appVersionLocal).
		Msg("starting XKVM")

	// initialize usb gadget
	setProcTitle("initUsbGadget")
	initUsbGadget()

	setProcTitle("initNative")
	initNative(systemVersionLocal, appVersionLocal)

	http.DefaultClient.Timeout = 1 * time.Minute

	err = rootcerts.UpdateDefaultTransport()
	if err != nil {
		logger.Warn().Err(err).Msg("failed to load Root CA certificates")
	}
	logger.Info().
		Int("ca_certs_loaded", len(rootcerts.Certs())).
		Msg("loaded Root CA certificates")

	// OTA functionality disabled
	// initOta()

	http.DefaultClient.Timeout = 1 * time.Minute

	// Initialize network
	setProcTitle("initNetwork")
	if err := initNetwork(); err != nil {
		logger.Error().Err(err).Msg("failed to initialize network")
		// TODO: reset config to default
		os.Exit(1)
	}

	// Initialize mDNS
	setProcTitle("initMdns")
	if err := initMdns(); err != nil {
		logger.Error().Err(err).Msg("failed to initialize mDNS")
	}

	setProcTitle("initPrometheus")
	initPrometheus()
	if err := setInitialVirtualMediaState(); err != nil {
		logger.Warn().Err(err).Msg("failed to set initial virtual media state")
	}

	if err := initImagesFolder(); err != nil {
		logger.Warn().Err(err).Msg("failed to init images folder")
	}
	initJiggler()

	// start video sleep mode timer
	startVideoSleepModeTicker()

	// OTA auto-update functionality disabled
	/*
		go func() {
			// wait for 15 minutes before starting auto-update checks
			// this is to avoid interfering with initial setup processes
			// and to ensure the system is stable before checking for updates
			time.Sleep(15 * time.Minute)

			for {
				logger.Info().Bool("auto_update_enabled", config.AutoUpdateEnabled).Msg("auto-update check")
				if !config.AutoUpdateEnabled {
					logger.Debug().Msg("auto-update disabled")
					time.Sleep(5 * time.Minute) // we'll check if auto-updates are enabled in five minutes
					continue
				}

				if currentSession != nil {
					logger.Debug().Msg("skipping update since a session is active")
					time.Sleep(1 * time.Minute)
					continue
				}

				includePreRelease := config.IncludePreRelease
				err = otaState.TryUpdate(context.Background(), ota.UpdateParams{
					DeviceID:          GetDeviceID(),
					IncludePreRelease: includePreRelease,
				})
				if err != nil {
					logger.Warn().Err(err).Msg("failed to auto update")
				}

				time.Sleep(1 * time.Hour)
			}
		}()
	*/

	//go RunFuseServer()
	go RunWebServer()

	go RunWebSecureServer()
	// Web secure server is started only if TLS mode is enabled
	if config.TLSMode != "" {
		startWebSecureServer()
	}

	// As websocket client already checks if the cloud token is set, we can start it here.
	// go RunWebsocketClient()

	initSerialPort()
	initHostMetricsListener()
	initVFD()

	setProcTitle("ready")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	logger.Log().Msg("XKVM Shutting Down")
	if gadget != nil {
		if err := gadget.Close(); err != nil {
			usbLogger.Warn().Err(err).Msg("failed to close USB gadget")
		}
	}
	//if fuseServer != nil {
	//	err := setMassStorageImage(" ")
	//	if err != nil {
	//		logger.Infof("Failed to unmount mass storage image: %v", err)
	//	}
	//	err = fuseServer.Unmount()
	//	if err != nil {
	//		logger.Infof("Failed to unmount fuse: %v", err)
	//	}

	// os.Exit(0)
}
