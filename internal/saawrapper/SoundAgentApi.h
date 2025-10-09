#ifndef SOUND_AGENT_API_H
#define SOUND_AGENT_API_H

/**
 * @file SoundAgentApi.h
 * @brief C API to monitor and query default audio devices.
 * Steps:
 * 1. Call ::SaaInitialize -> get handle.
 * 2. (Optional) ::SaaRegisterCallbacks to receive change events.
 * 3. Call ::SaaGetDefaultRender / ::SaaGetDefaultCapture to query devices.
 * 4. Call ::SaaUnInitialize before exit / unloading.
 * Threading: Serialize initialize/uninitialize. Callbacks may fire on worker thread; keep them fast and thread-safe.
 * Errors: 0 = success (non‑zero reserved).
 * Strings: ANSI, truncated with null terminator.
 * Logging: Supply log callback at init to receive async log messages.
 */

#include <Windows.h>

#ifdef ED_EXPORTS
#define SAA_EXPORT_IMPORT_DECL __declspec(dllexport)  /**< Export when building DLL */
#else
#define SAA_EXPORT_IMPORT_DECL __declspec(dllimport)  /**< Import when consuming DLL */
#endif

#ifdef __cplusplus
    extern "C" {
#endif

    /** Opaque instance handle (do not interpret). Obtained via ::SaaInitialize. */
    typedef DWORD64 SaaHandle;

    /** Result code. 0 = success. Future values reserved. */
    typedef INT32 SaaResult;

    /** Device description. Unused fields zeroed. BOOL uses Win32 TRUE/FALSE. */
    typedef struct {
        CHAR   PnpId[80];          /**< Device PnP id. */
        CHAR   Name[128];          /**< Device friendly name. */
        BOOL   IsRender;           /**< Device can render audio. */
        BOOL   IsCapture;          /**< Device can capture audio. */
        UINT16 RenderVolume;       /**< Current render volume (implementation units). */
        UINT16 CaptureVolume;      /**< Current capture volume (implementation units). */
    } SaaDescription;

    /** Log message forwarded from internal logger. */
    typedef struct {
        CHAR Level[12];    /**< Severity: trace, debug, info, warn, warning, error, critical. */
        CHAR Content[256]; /**< Message text. */
    } SaaLogMessage;

    /** Default device / volume change notification. presentOrAbsent: TRUE if a default device exists. */
    typedef void(__stdcall* TSaaDefaultChangedCallback)(
        _In_ BOOL presentOrAbsent
        );

    /** Asynchronous log message callback. */
    typedef void(__stdcall* TSaaGotLogMessageCallback)(
        _In_ SaaLogMessage message
        );

    /**
     * Initialize library. Acquire handle. Optionally set log callback + app id info.
     * handle: out handle. gotLogMessageCallback: optional. appName/appVersion: optional identifiers.
     */
    SAA_EXPORT_IMPORT_DECL
        SaaResult __stdcall SaaInitialize(
            _Out_ SaaHandle* handle,
            _In_opt_ TSaaGotLogMessageCallback gotLogMessageCallback,
            _In_opt_ const CHAR* appName,
            _In_opt_ const CHAR* appVersion
        );

    /**
     * Register or replace callbacks for default render/capture changes. Pass NULL to disable each.
     * Implicitly refreshes internal device list.
     */
    SAA_EXPORT_IMPORT_DECL
        SaaResult __stdcall SaaRegisterCallbacks(
            _In_ SaaHandle handle,
            _In_opt_ TSaaDefaultChangedCallback defaultRenderChangedCallback,
            _In_opt_ TSaaDefaultChangedCallback defaultCaptureChangedCallback
        );

    /** Get current default render device (or zeroed struct if none). description must be non-null. */
    SAA_EXPORT_IMPORT_DECL
        SaaResult __stdcall SaaGetDefaultRender(
            _In_ SaaHandle handle,
            _Out_ SaaDescription* description
        );

    /** Get current default capture device (or zeroed struct if none). description must be non-null. */
    SAA_EXPORT_IMPORT_DECL
        SaaResult __stdcall SaaGetDefaultCapture(
            _In_ SaaHandle handle,
            _Out_ SaaDescription* description
        );

    /** Uninitialize library. Invalidate handle. Safe to call multiple times (idempotent). */
    SAA_EXPORT_IMPORT_DECL
        SaaResult __stdcall SaaUnInitialize(
            _In_ SaaHandle handle
        );

#ifdef __cplusplus
}
#endif

#endif // SOUND_AGENT_API_H
