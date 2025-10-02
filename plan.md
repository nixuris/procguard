# ProcGuard Development Roadmap

This document outlines the planned features and refactoring work for the ProcGuard application.

## Current Phase: Hybrid Web Blocking & Self-Installation

This architecture uses a browser extension for web blocking and includes a self-installing mechanism for a seamless user experience.

-   **Goal:** Block websites using a browser extension controlled by the Go daemon, and have the application automatically install itself on first run.
-   **Architecture:**
    -   The **Go Daemon** is the core application. On first run, it will install itself to a user-specific directory.
    -   The **Browser Extension** acts as the enforcer within the browser.
    -   **Native Messaging** connects the daemon and the extension.

### Step 1: Implement First-Run Self-Installation

-   **Status:** In Progress
-   **Goal:** The main executable will manage its own installation when double-clicked.
-   **Tasks:**
    -   [x] **Code Refactoring:** Centralize application path logic into a single `config.GetAppDataDir()` function.
    -   [ ] **Go Daemon: Implement First-Run Logic**
        -   On startup, the app will check if it is running from its correct installation path in `%LOCALAPPDATA%`.
        -   **If not installed:** It will perform the installation:
            1.  Copy its own executable to the target installation path.
            2.  Create the Native Messaging Host manifest file in the same directory.
            3.  Register the host with Chrome by creating a registry key.
            4.  Relaunch itself from the new path and exit the current process.
        -   **If installed:** It will proceed with normal operation.

    -   [ ] **Go Daemon: Register Native Host with Chrome**
        -   **Method:** Create a registry key in `HKEY_CURRENT_USER`.
        -   **Justification:** This is the only installation method supported by Google for a non-admin, user-wide installation on Windows. It does **not** require administrator rights as it only modifies the current user's private part of the registry.
        -   **Function:** The key simply points the browser to the location of the `.json` manifest file, telling Chrome how to launch the Go application.

### Step 2: Implement the Core Hybrid Components

-   **Status:** Not Started
-   **Tasks:**
    -   [ ] **Go Daemon: Implement Native Messaging Host**
        -   Create a new execution mode triggered by a command-line flag (e.g., `--native-host`).
        -   In this mode, the application will handle length-prefixed JSON messages from the browser over `stdin`/`stdout`.

    -   [ ] **Browser Extension: Implement Native Client**
        -   Use `chrome.runtime.connectNative` to communicate with the Go host.
        -   Request the blocklist on startup and update `declarativeNetRequest` rules.

### Step 3: Implement User-Guided Extension Installation

-   **Status:** Not Started
-   **Goal:** Dynamically unlock the web management UI only when the extension is detected.
-   **Tasks:**
    -   [ ] **Browser Extension:** Use a content script to dispatch a custom event on the web GUI page.
    -   [ ] **Web GUI:** Listen for the custom event. If detected, show the web management UI. If not, show a prompt to install the extension from the Chrome Web Store.
