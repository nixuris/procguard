# ProcGuard Development Roadmap

This document outlines the planned features and refactoring work for the ProcGuard application.

## Current Phase: Web Blocking

### Step 1: Implement `hosts` File-Based Web Blocking
-   **Status:** Not Started
-   **Goal:** Block websites by modifying the system's `hosts` file.
-   **Tasks:**
    -   [ ] **Backend:** Create a new `internal/hosts` package to manage reading and writing to the `hosts` file on both Windows and Linux.
    -   [ ] **Backend:** Implement a mechanism to request administrator/root privileges when the `hosts` file needs to be modified.
    -   [ ] **Backend:** Create a new `web_blocklist.json` file and a corresponding `internal/webblocklist` package to manage the list of blocked domains.
    -   [ ] **Backend:** Create API endpoints (e.g., `/api/web-block`, `/api/web-unblock`) that trigger the `hosts` file modification.
    -   [ ] **UI:** Create a "Web Management" view in the GUI with controls to add, remove, and view blocked domains.
