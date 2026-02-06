# Refactoring/Design Plan: UI Standardization & Enhancement

## 1. Executive Summary & Goals
The primary objective of this plan is to unify the User Interface (UI) of the `dotkeeper` application across all tabs, ensuring a consistent, "beautiful," and intuitive user experience. The current implementation uses mixed rendering approaches (BubbleTea components vs. manual string building), resulting in visual inconsistencies.

**Key Goals:**
1.  **Visual Consistency:** Establish a shared design language (typography, spacing, colors) across all views.
2.  **Component Standardization:** Unify list rendering, status bars, and headers using shared delegates and helper functions.
3.  **Enhanced Aesthetics:** Implement "Card" layouts for the dashboard and polished selection styles for lists and settings.
4.  **Intuitiveness:** Clarify user feedback (status/error messages) and navigation cues.

## 2. Current Situation Analysis
*   **Inconsistent Lists:** `backuplist.go`, `logs.go`, and `restore.go` use `bubbles/list` but likely with default styling that may not perfectly match the custom styling in `settings.go`.
*   **Plain Dashboard:** `dashboard.go` renders plain text lines, which feels unpolished compared to the interactive components.
*   **Manual Layouts:** `settings.go` manually renders its menu items, leading to potential misalignment with standard list components.
*   **Status/Footer Variation:** Error messages and success notifications are rendered differently in each view (positioning, colors, formatting).
*   **Styling:** `internal/tui/views/styles.go` exists but lacks comprehensive definitions for layout containers and specific component states.

## 3. Proposed Solution / Refactoring Strategy

### 3.1. High-Level Design
The refactoring will enforce a "Common View Layout" pattern across all tabs:
1.  **Content Area:** The primary interaction space (List, Grid, or Form).
2.  **Status/Footer Area:** A fixed-height region at the bottom for status messages (Success/Error) and help keys.

We will leverage `lipgloss` heavily to create defined boundaries and consistent padding.

### 3.2. Key Components / Modules

*   **`views/styles.go`**: Expanded to include a **Standard List Delegate** factory (to ensure all lists look identical) and **Layout Helpers** (for headers and footers).
*   **`views/dashboard.go`**: Refactored to use a Grid/Card layout for statistics.
*   **`views/settings.go`**: Refactored to visually mimic the `list` component for consistency, or migrated to use `bubbles/list` for navigation.
*   **`views/helpers.go`**: New helper for rendering the standardized Status Bar.

### 3.3. Detailed Action Plan / Phases

#### **Phase 1: Design System Foundation**
*   **Objective:** Define the core styles and shared components to be used by all views.
*   **Priority:** High
*   **Task 1.1: Expand `styles.go`**
    *   **Rationale:** Centralize all visual definitions.
    *   **Deliverable:** Update `Styles` struct.
        *   Add `AppContainer` and `ViewContainer` styles (padding/margins).
        *   Add `StatusBar` style (fixed height, border/separator).
        *   Create `NewListDelegate() list.DefaultDelegate` function. This will return a delegate configured with the app's Purple/Green/Grey palette, ensuring consistent selection highlighting and description coloring across Backups, Logs, and Restore lists.
*   **Task 1.2: Create Status Bar Helper**
    *   **Rationale:** Uniform feedback presentation.
    *   **Deliverable:** Function `RenderStatusBar(width int, status, err string, helpView string) string` in `helpers.go`. This function should handle text truncation and color coding (Green for status, Red for error).

#### **Phase 2: List Views Standardization**
*   **Objective:** Apply the design system to existing list-based views.
*   **Priority:** Medium
*   **Task 2.1: Refactor `backuplist.go`**
    *   **Rationale:** Align with new styles.
    *   **Deliverable:**
        *   Use `NewListDelegate()` for initialization.
        *   Remove manual title rendering if `bubbles/list` title is enabled, OR disable list title and use standard `styles.Title`.
        *   Replace the manual footer rendering string builder logic with `RenderStatusBar()`.
*   **Task 2.2: Refactor `logs.go`**
    *   **Rationale:** Align with new styles.
    *   **Deliverable:** Same as Task 2.1. Ensure icons (✓/✗) are aligned with the new delegate's spacing.
*   **Task 2.3: Refactor `restore.go`**
    *   **Rationale:** This is a complex multi-phase view; needs careful alignment.
    *   **Deliverable:**
        *   Apply `NewListDelegate()` to both `backupList` and `fileList`.
        *   Ensure the "Diff Preview" viewport has a border style matching other containers.
        *   Use `RenderStatusBar()` for feedback during all phases.

#### **Phase 3: Dashboard & Settings Polish**
*   **Objective:** Bring non-list views up to the same visual standard.
*   **Priority:** Medium
*   **Task 3.1: Dashboard UI Overhaul**
    *   **Rationale:** Make the landing screen "beautiful".
    *   **Deliverable:** Refactor `View()`.
        *   Create "Stat Cards" using `lipgloss`: Box with border, bold label, large value.
        *   Example Cards: "Last Backup", "Total Files", "Storage Used".
        *   Arrange cards horizontally using `lipgloss.JoinHorizontal` (or wrap if width is low).
        *   Style the "Quick Actions" section as a compact list or button row.
*   **Task 3.2: Settings UI Standardization**
    *   **Rationale:** Manual text rendering currently looks disjointed from the rest of the app.
    *   **Deliverable:**
        *   Update rendering to mimic the `list.Delegate` spacing and selection style (Purple foreground + dark background for active row).
        *   Align "Label" and "Value" columns strictly.
        *   Use `RenderStatusBar()` for save confirmation/errors.

#### **Phase 4: Final Consistency Check**
*   **Objective:** Ensure navigation and global elements are cohesive.
*   **Priority:** Low
*   **Task 4.1: TUI Container Adjustment (`tui/view.go`)**
    *   **Rationale:** The outer frame needs to support the inner views cleanly.
    *   **Deliverable:** Ensure `tui/view.go` applies correct margins so that inner views (which might have their own borders) don't double-up margins or overflow.
*   **Task 4.2: Help Overlay (`tui/help.go`)**
    *   **Deliverable:** Style the help popup to match the new color theme (Purple borders/titles).

## 4. Key Considerations & Risk Mitigation

### 4.1. Technical Risks
*   **Terminal Width Constraints:** "Card" layouts in the Dashboard might break on very narrow terminals.
    *   *Mitigation:* Use `lipgloss.Place` or conditional logic in `View()` to switch to a vertical stack layout if `width < 60`.
*   **List Resize Flickering:** Aggressive style updates might cause layout shifts.
    *   *Mitigation:* Ensure `SetSize` is called correctly in `Update` on `WindowSizeMsg`.

### 4.2. Dependencies
*   Depends on `charmbracelet/lipgloss` (already present). No new dependencies required.

### 4.3. Non-Functional Requirements
*   **Usability:** Consistent visual cues (e.g., "Purple means Active/Selected") reduce cognitive load.
*   **Maintainability:** Centralizing styles in `styles.go` means future theme changes only need to happen in one place.

## 5. Success Metrics
*   **Visual Check:** All 5 tabs share identical margins, font sizes, and color logic.
*   **Code Quality:** Reduction in duplicate styling code across views.
*   **Feedback:** User (you) perceives the application as "prettier" and "intuitive" (subjective but primary goal).

## 6. Assumptions
*   The current standard terminal width is at least 80 columns.
*   Nerd Fonts are **not** strictly required, but standard Unicode symbols (✓, •, │) are supported by the user's terminal.

## 7. Open Questions
*   Should we implement a "Dark/Light" theme toggle in Settings? (Assuming Dark mode default for now based on colors like `#2A2A2A`).
*   Does the "Files Tracked" stat on Dashboard need to be real-time? (Assuming current implementation of recalculating on load/refresh is sufficient).