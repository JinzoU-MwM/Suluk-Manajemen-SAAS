# Suluk Mobile — Capacitor (Android) wrapper

Wraps the existing Svelte PWA as a native **Android** app via Capacitor 6.
**Hosted mode:** the native shell loads `https://suluk.site/#/app` directly, so
web changes ship through the normal frontend deploy — **no app rebuild / store
resubmit needed** unless you change native config, icons, plugins, or the loaded
URL. (iOS can be added later on a Mac — see bottom.)

- **App ID:** `site.suluk.app`  ·  **Name:** Suluk  ·  **Theme:** deep green `#0F3D2E`
- Config: `capacitor.config.json`  ·  Native project: `android/`

## Prerequisites (build machine)
- Node 20+ (already used for the web app)
- **JDK 17** (Android Gradle Plugin 8 requires it)
- **Android Studio** (easiest) or the Android SDK + platform-tools.
  Windows is fine — no Mac needed for Android.

## Build & run
```bash
cd frontend-svelte
npm install              # first time
npm run build            # produces dist/ (required by cap sync)
npx cap sync android     # copies config + plugins into android/
npx cap open android     # opens Android Studio → Run ▶ on device/emulator
```
Shortcut: `npm run cap:android` (build + sync + open in one step).

### Command-line APK (no Android Studio UI)
```bash
cd frontend-svelte
npm run build && npx cap sync android
cd android
./gradlew assembleDebug      # → app/build/outputs/apk/debug/app-debug.apk
```
Install on a connected device: `adb install -r app/build/outputs/apk/debug/app-debug.apk`.

## App icon & splash (recommended before release)
The project still uses Capacitor's placeholder icon. Generate branded assets from
the Suluk mark:
```bash
# put a 1024x1024 icon at frontend-svelte/assets/icon.png (and optional splash.png)
npm i -D @capacitor/assets
npx @capacitor/assets generate --android
```
(Source mark lives in `public/brand/suluk-icon.png` — upscale to 1024² first.)

## Releasing to Google Play
1. Generate a keystore (once):
   `keytool -genkey -v -keystore suluk-release.keystore -alias suluk -keyalg RSA -keysize 2048 -validity 10000`
   — keep this file + passwords safe and OUT of git.
2. Add a `signingConfigs.release` + `buildTypes.release.signingConfig` in
   `android/app/build.gradle` (reference the keystore via `~/.gradle` props, not committed).
3. `cd android && ./gradlew bundleRelease` → `app/build/outputs/bundle/release/app-release.aab`.
4. Upload the `.aab` in the Play Console. Bump `versionCode`/`versionName` in
   `android/app/build.gradle` each release.

## Updating the app
- **Web/UI/API changes:** just deploy the frontend as usual (`docker compose build
  frontend && up -d frontend`). The hosted webview picks them up on next launch —
  no APK rebuild.
- **Native changes** (icon, splash, permissions, plugins, the loaded URL in
  `capacitor.config.json`): edit, `npx cap sync android`, rebuild + resubmit the APK/AAB.

## Permissions
`android/app/src/main/AndroidManifest.xml` declares `INTERNET` (required) and
`CAMERA` (the AI Scanner uses `<input type="file" capture>`). Android prompts for
camera at first use.

## Add iOS later (needs macOS + Xcode + Apple Developer account)
```bash
npm i @capacitor/ios
npx cap add ios
npx cap sync ios
npx cap open ios     # Xcode → set Team/signing → Run
```
`capacitor.config.json` already carries the shared settings (server URL, splash,
status bar); iOS will inherit them.
