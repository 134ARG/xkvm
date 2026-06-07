import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useResizeObserver } from "usehooks-ts";

import { cx } from "@/cva.config";
import { isWindows } from "@/utils";
import useKeyboard from "@hooks/useKeyboard";
import useMouse from "@hooks/useMouse";
import { useRTCStore, useSettingsStore, useVideoStore } from "@hooks/stores";
import VirtualKeyboard from "@components/VirtualKeyboard";
import Actionbar from "@components/ActionBar";
import MacroBar from "@components/MacroBar";
import InfoBar from "@components/InfoBar";
import VideoCasCanvas from "@components/VideoCasCanvas";
import {
  HDMIErrorOverlay,
  LoadingVideoOverlay,
  NoAutoplayPermissionsOverlay,
  PointerLockBar,
} from "@components/VideoOverlay";
import { keys } from "@/keyboardMappings";
import notifications from "@/notifications";
import { m } from "@localizations/messages.js";

export default function WebRTCVideo({ hasConnectionIssues }: { hasConnectionIssues: boolean }) {
  // Video and stream related refs and states
  const videoElm = useRef<HTMLVideoElement>(null);
  const fullscreenContainerRef = useRef<HTMLDivElement>(null);
  const videoFrameRef = useRef<HTMLDivElement>(null);
  const [videoElement, setLocalVideoElement] = useState<HTMLVideoElement | null>(null);
  const [casFrameSize, setCasFrameSize] = useState({ width: 0, height: 0 });
  const [casReadyKey, setCasReadyKey] = useState<string | null>(null);
  const { mediaStream, peerConnectionState } = useRTCStore();
  const [isPlaying, setIsPlaying] = useState(false);
  const [isPointerLockActive, setIsPointerLockActive] = useState(false);
  const [isKeyboardLockActive, setIsKeyboardLockActive] = useState(false);

  // Store hooks
  const settings = useSettingsStore();
  const { handleKeyPress, resetKeyboardState } = useKeyboard();
  const {
    getRelMouseMoveHandler,
    getAbsMouseMoveHandler,
    getMouseWheelHandler,
    resetMousePosition,
  } = useMouse();
  const {
    setClientSize: setVideoClientSize,
    setSize: setVideoSize,
    width: videoWidth,
    height: videoHeight,
    clientWidth: videoClientWidth,
    clientHeight: videoClientHeight,
    hdmiState,
    setVideoElement,
  } = useVideoStore();

  // Video enhancement settings
  const { videoSaturation, videoBrightness, videoContrast, videoSharpness } = useSettingsStore();
  const isCasEnabled = videoSharpness > 0;

  // RTC related states
  const { peerConnection } = useRTCStore();

  // HDMI and UI states
  const hdmiError = ["no_lock", "no_signal", "out_of_range"].includes(hdmiState);
  const isVideoLoading = !isPlaying;
  const shouldRenderCas =
    isCasEnabled &&
    isPlaying &&
    !hdmiError &&
    !hasConnectionIssues &&
    peerConnectionState === "connected";
  const hasCasFrameSize = casFrameSize.width > 0 && casFrameSize.height > 0;
  const casRenderKey =
    shouldRenderCas && hasCasFrameSize
      ? `${Math.round(casFrameSize.width)}x${Math.round(casFrameSize.height)}`
      : null;
  const isCasActive = casRenderKey !== null && casReadyKey === casRenderKey;

  const handleResize = useCallback(
    ({ width, height }: { width: number | undefined; height: number | undefined }) => {
      if (!videoElm.current) return;
      // Do something with width and height, e.g.:
      setVideoClientSize(width || 0, height || 0);
      setVideoSize(videoElm.current.videoWidth, videoElm.current.videoHeight);
    },
    [setVideoClientSize, setVideoSize],
  );

  // AltGr Fix for Windows Clients
  const altGrSyntheticThresholdMs = 3;
  const isWindowsClient = useMemo(() => isWindows(), []);
  const lastKeyDownRef = useRef<{ hidKey: number; time: number } | null>(null);
  const altGrLoopRef = useRef(false);

  useResizeObserver({
    ref: videoElm as React.RefObject<HTMLElement>,
    onResize: handleResize,
  });

  const updateCasFrameSize = useCallback(() => {
    if (!fullscreenContainerRef.current) return;

    const { width: containerWidth, height: containerHeight } =
      fullscreenContainerRef.current.getBoundingClientRect();
    if (containerWidth <= 0 || containerHeight <= 0) return;

    const videoAspectRatio = videoWidth > 0 && videoHeight > 0 ? videoWidth / videoHeight : 4 / 3;

    let width = containerWidth;
    let height = width / videoAspectRatio;

    if (height > containerHeight) {
      height = containerHeight;
      width = height * videoAspectRatio;
    }

    setCasFrameSize(prev => {
      if (
        Math.round(prev.width) === Math.round(width) &&
        Math.round(prev.height) === Math.round(height)
      ) {
        return prev;
      }
      return { width, height };
    });
    setVideoClientSize(width, height);
  }, [setVideoClientSize, videoHeight, videoWidth]);

  useResizeObserver({
    ref: fullscreenContainerRef as React.RefObject<HTMLElement>,
    onResize: updateCasFrameSize,
  });

  const updateVideoSizeStore = useCallback(
    (videoElm: HTMLVideoElement) => {
      setVideoClientSize(videoElm.clientWidth, videoElm.clientHeight);
      setVideoSize(videoElm.videoWidth, videoElm.videoHeight);
    },
    [setVideoClientSize, setVideoSize],
  );

  const onVideoPlaying = useCallback(() => {
    setIsPlaying(true);
    if (videoElm.current) updateVideoSizeStore(videoElm.current);
  }, [updateVideoSizeStore]);

  const handleCasReady = useCallback(() => {
    if (casRenderKey) setCasReadyKey(casRenderKey);
  }, [casRenderKey]);

  const handleCasUnavailable = useCallback(() => {
    setCasReadyKey(null);
  }, []);

  const setVideoNode = useCallback(
    (element: HTMLVideoElement | null) => {
      videoElm.current = element;
      setLocalVideoElement(element);
      setVideoElement(element);
      if (element) {
        if (mediaStream && element.srcObject !== mediaStream) element.srcObject = mediaStream;
        updateVideoSizeStore(element);
      }
    },
    [mediaStream, setVideoElement, updateVideoSizeStore],
  );

  // On mount, get the video size
  useEffect(
    function updateVideoSizeOnMount() {
      if (videoElm.current) updateVideoSizeStore(videoElm.current);
    },
    [updateVideoSizeStore],
  );

  useEffect(
    function updateCasFrameOnLayoutChange() {
      const animationFrame = requestAnimationFrame(updateCasFrameSize);
      const handleLayoutChange = () => requestAnimationFrame(updateCasFrameSize);

      window.addEventListener("resize", handleLayoutChange);
      document.addEventListener("fullscreenchange", handleLayoutChange);

      return () => {
        cancelAnimationFrame(animationFrame);
        window.removeEventListener("resize", handleLayoutChange);
        document.removeEventListener("fullscreenchange", handleLayoutChange);
      };
    },
    [updateCasFrameSize],
  );

  // Pointer lock and keyboard lock related
  const isFullscreenEnabled = document.fullscreenEnabled;

  const checkNavigatorPermissions = useCallback(async (permissionName: string) => {
    if (!navigator || !navigator.permissions || !navigator.permissions.query) {
      return false; // if can't query permissions, assume NOT granted
    }

    try {
      const name = permissionName as PermissionName;
      const { state } = await navigator.permissions.query({ name });
      return state === "granted";
    } catch {
      // ignore errors
    }
    return false; // if query fails, assume NOT granted
  }, []);

  const isVideoPointerLocked = useCallback(() => {
    const lockedElement = document.pointerLockElement;
    return lockedElement === videoElm.current || lockedElement === videoFrameRef.current;
  }, []);

  const requestPointerLock = useCallback(async () => {
    const pointerLockTarget = videoFrameRef.current ?? videoElm.current;
    if (settings.mouseMode !== "relative" || document.pointerLockElement) return;
    if (!pointerLockTarget) {
      console.warn("Pointer lock target unavailable");
      return;
    }
    if (typeof pointerLockTarget.requestPointerLock !== "function") {
      console.warn("Pointer lock is not supported by this browser");
      return;
    }

    try {
      await pointerLockTarget.requestPointerLock();
    } catch (error) {
      console.warn("Pointer lock request failed", error);
    }
  }, [settings.mouseMode]);

  const requestKeyboardLock = useCallback(async () => {
    if (videoElm.current === null) return;

    const isKeyboardLockGranted = await checkNavigatorPermissions("keyboard-lock");

    if (isKeyboardLockGranted && navigator && "keyboard" in navigator) {
      try {
        // @ts-expect-error - keyboard lock is not supported in all browsers
        await navigator.keyboard.lock();
        setIsKeyboardLockActive(true);
      } catch {
        // ignore errors
      }
    }
  }, [checkNavigatorPermissions, setIsKeyboardLockActive]);

  const releaseKeyboardLock = useCallback(async () => {
    if (
      fullscreenContainerRef.current === null ||
      document.fullscreenElement !== fullscreenContainerRef.current
    )
      return;

    if (navigator && "keyboard" in navigator) {
      try {
        // @ts-expect-error - keyboard unlock is not supported in all browsers
        await navigator.keyboard.unlock();
      } catch {
        // ignore errors
      }
      setIsKeyboardLockActive(false);
    }
  }, [setIsKeyboardLockActive]);

  useEffect(() => {
    const handlePointerLockChange = () => {
      if (isVideoPointerLocked()) {
        notifications.success(m.video_pointer_lock_enabled());
        setIsPointerLockActive(true);
      } else {
        notifications.success(m.video_pointer_lock_disabled());
        setIsPointerLockActive(false);
      }
    };
    const handlePointerLockError = () => {
      console.warn("Pointer lock request failed");
      setIsPointerLockActive(false);
    };

    const abortController = new AbortController();
    const signal = abortController.signal;

    document.addEventListener("pointerlockchange", handlePointerLockChange, { signal });
    document.addEventListener("pointerlockerror", handlePointerLockError, { signal });

    return () => {
      abortController.abort();
    };
  }, [isVideoPointerLocked]);

  const requestFullscreen = useCallback(async () => {
    if (!isFullscreenEnabled || !fullscreenContainerRef.current) return;

    // per https://wicg.github.io/keyboard-lock/#system-key-press-handler
    // If keyboard lock is activated after fullscreen is already in effect, then the user my
    // see multiple messages about how to exit fullscreen. For this reason, we recommend that
    // developers call lock() before they enter fullscreen:
    await requestKeyboardLock();
    await requestPointerLock();

    await fullscreenContainerRef.current.requestFullscreen({
      navigationUI: "show",
    });
  }, [isFullscreenEnabled, requestKeyboardLock, requestPointerLock]);

  // setup to release the keyboard lock anytime the fullscreen ends
  useEffect(() => {
    if (!videoElm.current) return;

    const handleFullscreenChange = () => {
      if (!document.fullscreenElement) {
        releaseKeyboardLock();
      }
    };

    document.addEventListener("fullscreenchange", handleFullscreenChange);
  }, [releaseKeyboardLock]);

  const absMouseMoveHandler = useMemo(
    () =>
      getAbsMouseMoveHandler({
        videoClientWidth,
        videoClientHeight,
        videoWidth,
        videoHeight,
      }),
    [getAbsMouseMoveHandler, videoClientWidth, videoClientHeight, videoWidth, videoHeight],
  );

  const relMouseMoveHandler = useMemo(() => {
    const handler = getRelMouseMoveHandler();
    return (e: MouseEvent) => {
      if (!isVideoPointerLocked()) return;
      handler(e);
    };
  }, [getRelMouseMoveHandler, isVideoPointerLocked]);

  const mouseWheelHandler = useMemo(() => getMouseWheelHandler(), [getMouseWheelHandler]);

  function getAdjustedKeyCode(e: KeyboardEvent) {
    const key = e.key;
    let code = e.code;

    if (code == "IntlBackslash" && ["`", "~"].includes(key)) {
      code = "Backquote";
    } else if (code == "Backquote" && ["§", "±"].includes(key)) {
      code = "IntlBackslash";
    }
    // For Japanese 106/109
    else if (code === "IntlYen") {
      code = "Yen";
    } else if (code === "IntlRo") {
      code = "KeyRO";
    } else if (code === "Convert") {
      code = "Henkan";
    } else if (code === "NonConvert") {
      code = "Muhenkan";
    }

    return code;
  }

  const keyDownHandler = useCallback(
    (e: KeyboardEvent) => {
      e.preventDefault();
      const code = getAdjustedKeyCode(e);
      const hidKey = keys[code];

      if (hidKey === undefined) {
        console.warn(`Key down not mapped: ${code}`);
        return;
      }

      // Detect Windows synthetic AltGr (CtrlLeft then AltRight within ~3ms) and cancel the synthetic Ctrl
      if (isWindowsClient) {
        // Buffer ControlLeft briefly; if no AltRight follows within the threshold, treat it as a real ControlLeft press.
        if (hidKey === keys.ControlLeft) {
          const controlLeftDownTime = e.timeStamp;
          lastKeyDownRef.current = { hidKey, time: controlLeftDownTime };
          setTimeout(() => {
            if (
              lastKeyDownRef.current?.hidKey === keys.ControlLeft &&
              lastKeyDownRef.current.time === controlLeftDownTime
            ) {
              lastKeyDownRef.current = null;
              handleKeyPress(keys.ControlLeft, true);
            }
          }, altGrSyntheticThresholdMs);
          return;
        }

        // If AltRight arrives shortly after ControlLeft, treat the pair as AltGr and cancel the pending ControlLeft.
        if (
          hidKey === keys.AltRight &&
          lastKeyDownRef.current?.hidKey === keys.ControlLeft &&
          e.timeStamp - lastKeyDownRef.current.time <= altGrSyntheticThresholdMs
        ) {
          altGrLoopRef.current = true;
          lastKeyDownRef.current = null;
        }
      }

      // When pressing the meta key + another key, the key will never trigger a keyup
      // event, so we need to clear the keys after a short delay
      // https://bugs.chromium.org/p/chromium/issues/detail?id=28089
      // https://bugzilla.mozilla.org/show_bug.cgi?id=1299553
      if (e.metaKey && hidKey < 0xe0) {
        setTimeout(() => {
          console.debug(`Forcing the meta key release of associated key: ${hidKey}`);
          handleKeyPress(hidKey, false);
        }, 10);
      }
      console.debug(`Key down: ${hidKey}`);
      handleKeyPress(hidKey, true);

      if (!isKeyboardLockActive && hidKey === keys.MetaLeft) {
        // If the left meta key was just pressed and we're not keyboard locked
        // we'll never see the keyup event because the browser is going to lose
        // focus so set a deferred keyup after a short delay
        setTimeout(() => {
          console.debug(`Forcing the left meta key release`);
          handleKeyPress(hidKey, false);
        }, 100);
      }
    },
    [handleKeyPress, isKeyboardLockActive, isWindowsClient],
  );

  const keyUpHandler = useCallback(
    async (e: KeyboardEvent) => {
      e.preventDefault();
      const code = getAdjustedKeyCode(e);
      const hidKey = keys[code];

      if (hidKey === undefined) {
        console.warn(`Key up not mapped: ${code}`);
        return;
      }

      // On Windows, handle ControlLeft specially to preserve FIFO semantics with AltGr buffering.
      if (isWindowsClient && hidKey === keys.ControlLeft) {
        // Synthetic AltGr ControlLeft: never sent a down, swallow the release as well.
        if (altGrLoopRef.current) {
          altGrLoopRef.current = false;
          return;
        }

        // Very fast real Ctrl tap: flush the pending down before the up.
        if (lastKeyDownRef.current?.hidKey === keys.ControlLeft) {
          handleKeyPress(keys.ControlLeft, true);
        }

        lastKeyDownRef.current = null;
      }

      console.debug(`Key up: ${hidKey}`);
      handleKeyPress(hidKey, false);
    },
    [handleKeyPress, isWindowsClient],
  );

  const videoKeyUpHandler = useCallback((e: KeyboardEvent) => {
    if (!videoElm.current) return;

    // In fullscreen mode in chrome & safari, the space key is used to pause/play the video
    // there is no way to prevent this, so we need to simply force play the video when it's paused.
    // Fix only works in chrome based browsers.
    if (e.code === "Space") {
      if (videoElm.current.paused) {
        console.debug("Force playing video");
        videoElm.current.play();
      }
    }
  }, []);

  const addStreamToVideoElm = useCallback(
    (mediaStream: MediaStream) => {
      if (!videoElm.current) return;
      const videoElmRefValue = videoElm.current;
      videoElmRefValue.srcObject = mediaStream;
      updateVideoSizeStore(videoElmRefValue);
    },
    [updateVideoSizeStore],
  );

  useEffect(
    function updateVideoStreamOnNewTrack() {
      if (!peerConnection) return;
      const abortController = new AbortController();
      const signal = abortController.signal;

      peerConnection.addEventListener(
        "track",
        (e: RTCTrackEvent) => {
          addStreamToVideoElm(e.streams[0]);
        },
        { signal },
      );

      return () => {
        abortController.abort();
      };
    },
    [addStreamToVideoElm, peerConnection],
  );

  useEffect(
    function updateVideoStream() {
      if (!mediaStream) return;
      // We set the as early as possible
      addStreamToVideoElm(mediaStream);
    },
    [addStreamToVideoElm, mediaStream],
  );

  // Setup Keyboard Events
  useEffect(
    function setupKeyboardEvents() {
      const abortController = new AbortController();
      const signal = abortController.signal;

      document.addEventListener("keydown", keyDownHandler, { signal });
      document.addEventListener("keyup", keyUpHandler, { signal });

      window.addEventListener("blur", resetKeyboardState, { signal });
      document.addEventListener("visibilitychange", resetKeyboardState, { signal });

      return () => {
        abortController.abort();
      };
    },
    [keyDownHandler, keyUpHandler, resetKeyboardState],
  );

  // Setup Video Event Listeners
  useEffect(
    function setupVideoEventListeners() {
      const videoElmRefValue = videoElm.current;
      if (!videoElmRefValue) return;

      const abortController = new AbortController();
      const signal = abortController.signal;

      // To prevent the video from being paused when the user presses a space in fullscreen mode
      videoElmRefValue.addEventListener("keyup", videoKeyUpHandler, { signal });

      // We need to know when the video is playing to update state and video size
      videoElmRefValue.addEventListener("playing", onVideoPlaying, { signal });

      return () => {
        abortController.abort();
      };
    },
    [onVideoPlaying, videoKeyUpHandler],
  );

  // Setup Mouse Events
  useEffect(
    function setMouseModeEventListeners() {
      const videoFrameRefValue = videoFrameRef.current;
      if (!videoFrameRefValue) return;

      const isRelativeMouseMode = settings.mouseMode === "relative";
      const mouseHandler = isRelativeMouseMode ? relMouseMoveHandler : absMouseMoveHandler;
      const handlePointerLockRequest = () => {
        if (!isVideoPointerLocked()) requestPointerLock();
      };

      const abortController = new AbortController();
      const signal = abortController.signal;

      if (isRelativeMouseMode) {
        videoFrameRefValue.addEventListener("pointerdown", handlePointerLockRequest, { signal });
        videoFrameRefValue.addEventListener("click", handlePointerLockRequest, { signal });
      }

      videoFrameRefValue.addEventListener("mousemove", mouseHandler, { signal });
      videoFrameRefValue.addEventListener("pointerdown", mouseHandler, { signal });
      videoFrameRefValue.addEventListener("pointerup", mouseHandler, { signal });
      videoFrameRefValue.addEventListener("wheel", mouseWheelHandler, {
        signal,
        passive: true,
      });

      if (!isRelativeMouseMode) {
        // Reset the mouse position when the window is blurred or the document is hidden
        window.addEventListener("blur", resetMousePosition, { signal });
        document.addEventListener("visibilitychange", resetMousePosition, { signal });
      }

      const preventContextMenu = (e: MouseEvent) => e.preventDefault();
      videoFrameRefValue.addEventListener("contextmenu", preventContextMenu, { signal });

      return () => {
        abortController.abort();
      };
    },
    [
      isVideoPointerLocked,
      requestPointerLock,
      absMouseMoveHandler,
      relMouseMoveHandler,
      mouseWheelHandler,
      resetMousePosition,
      settings.mouseMode,
    ],
  );

  const containerRef = useRef<HTMLDivElement>(null);

  const hasNoAutoPlayPermissions = useMemo(() => {
    if (peerConnection?.connectionState !== "connected") return false;
    if (isPlaying) return false;
    if (hdmiError) return false;
    if (videoHeight === 0 || videoWidth === 0) return false;
    return true;
  }, [hdmiError, isPlaying, peerConnection?.connectionState, videoHeight, videoWidth]);

  const showPointerLockBar = useMemo(() => {
    if (settings.mouseMode !== "relative") return false;
    if (isPointerLockActive) return false;
    if (isVideoLoading) return false;
    if (!isPlaying) return false;
    if (videoHeight === 0 || videoWidth === 0) return false;
    return true;
  }, [isPlaying, isPointerLockActive, isVideoLoading, settings.mouseMode, videoHeight, videoWidth]);

  // Conditionally set the filter style so we don't fallback to software rendering if these values are default of 1.0
  const videoStyle = useMemo(() => {
    const isDefault =
      isCasActive || (videoSaturation === 1.0 && videoBrightness === 1.0 && videoContrast === 1.0);
    return isDefault
      ? {} // No filter if all settings are default (1.0)
      : {
          filter: `saturate(${videoSaturation}) brightness(${videoBrightness}) contrast(${videoContrast})`,
        };
  }, [isCasActive, videoSaturation, videoBrightness, videoContrast]);

  return (
    <div className="grid h-full w-full grid-rows-(--grid-layout)">
      <div className="flex min-h-[39.5px] flex-col">
        <div className="flex flex-col">
          <fieldset disabled={peerConnection?.connectionState !== "connected"} className="contents">
            <Actionbar requestFullscreen={requestFullscreen} />
            <MacroBar />
          </fieldset>
        </div>
      </div>

      <div ref={containerRef} className="h-full overflow-hidden">
        <div className="relative h-full">
          <div
            className={cx(
              "absolute inset-0 -z-0 bg-blue-50/40 opacity-80 dark:bg-slate-800/40",
              "bg-[radial-gradient(var(--color-blue-300)_0.5px,transparent_0.5px),radial-gradient(var(--color-blue-300)_0.5px,transparent_0.5px)] dark:bg-[radial-gradient(var(--color-slate-700)_0.5px,transparent_0.5px),radial-gradient(var(--color-slate-700)_0.5px,transparent_0.5px)]",
              "bg-position-[0_0,10px_10px]",
              "bg-size-[20px_20px]",
            )}
          />
          <div className="flex h-full flex-col">
            <div className="relative grow overflow-hidden">
              <div className="flex h-full flex-col">
                <div className="relative grid grow grid-rows-(--grid-bodyFooter) overflow-hidden">
                  {/* In relative mouse mode and under https, we enable the pointer lock, and to do so we need a bar to show the user to click on the video to enable mouse control */}
                  <PointerLockBar show={showPointerLockBar} />
                  <div className="relative mx-4 my-2 flex items-center justify-center overflow-hidden">
                    <div
                      ref={fullscreenContainerRef}
                      className="relative flex h-full w-full items-center justify-center"
                    >
                      <div
                        ref={videoFrameRef}
                        className="relative bg-black/50"
                        style={{
                          width: casFrameSize.width,
                          height: casFrameSize.height,
                        }}
                      >
                        <video
                          ref={setVideoNode}
                          autoPlay
                          controls={false}
                          onPlaying={onVideoPlaying}
                          onPlay={onVideoPlaying}
                          muted
                          playsInline
                          disablePictureInPicture
                          controlsList="nofullscreen"
                          style={videoStyle}
                          className={cx(
                            "h-full w-full object-contain transition-all duration-1000",
                            {
                              "cursor-none": settings.isCursorHidden,
                              "opacity-0!":
                                isVideoLoading ||
                                isCasActive ||
                                hdmiError ||
                                hasConnectionIssues ||
                                peerConnectionState !== "connected",
                              "animate-slideUpFade border border-slate-800/30 shadow-xs dark:border-slate-300/20":
                                isPlaying,
                            },
                          )}
                        />
                        {shouldRenderCas && hasCasFrameSize && (
                          <VideoCasCanvas
                            video={videoElement}
                            width={casFrameSize.width}
                            height={casFrameSize.height}
                            sharpness={videoSharpness}
                            saturation={videoSaturation}
                            brightness={videoBrightness}
                            contrast={videoContrast}
                            onReady={handleCasReady}
                            onUnavailable={handleCasUnavailable}
                          />
                        )}
                      </div>
                      {peerConnection?.connectionState == "connected" && !hasConnectionIssues && (
                        <div
                          style={{ animationDuration: "500ms" }}
                          className="pointer-events-none absolute inset-0 flex animate-slideUpFade items-center justify-center"
                        >
                          <div className="relative h-full w-full rounded-md">
                            <LoadingVideoOverlay show={isVideoLoading} />
                            <HDMIErrorOverlay show={hdmiError} hdmiState={hdmiState} />
                            <NoAutoplayPermissionsOverlay
                              show={hasNoAutoPlayPermissions}
                              onPlayClick={() => {
                                videoElm.current?.play();
                              }}
                            />
                          </div>
                        </div>
                      )}
                    </div>
                  </div>
                  <VirtualKeyboard />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div>
        <InfoBar />
      </div>
    </div>
  );
}
