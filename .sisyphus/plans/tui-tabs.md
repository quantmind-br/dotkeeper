# TUI Tab Bar — Padronização de Navegação

## TL;DR

> **Quick Summary**: Criar um componente TabBar para o TUI dotkeeper, padronizando todas as telas com uma barra de tabs visual que indica claramente a seção ativa, com navegação por números (1-5) e Tab cycling.
> 
> **Deliverables**:
> - Componente TabBar reutilizável em `internal/tui/components/tabbar.go`
> - Estilos de tab (ativa/inativa/separador) em `views/styles.go`
> - Navegação via teclas numéricas 1-5 com proteção contra input-consuming states
> - Tab cycling ajustado (sem FileBrowser no ciclo)
> - Ajuste de altura em todas as views para acomodar a tab bar
> - Help text e overlay atualizados
> - Testes unitários para o componente e integração
> 
> **Estimated Effort**: Medium
> **Parallel Execution**: YES — 3 waves
> **Critical Path**: Task 1 (styles) → Task 2 (component) → Task 4 (integration) → Task 6 (help) → Task 7 (tests)

---

## Context

### Original Request
Usuário quer padronizar todas as telas do TUI adicionando tabs para seção de forma que fosse de fácil identificação o menu que o usuário está.

### Interview Summary
**Key Discussions**:
- **Views com tabs**: Só as 5 principais — Dashboard, Backups, Restore, Settings, Logs. FileBrowser excluído (sub-componente). Setup excluído (wizard).
- **Estilo visual**: Underline na tab ativa (roxo #7D56F4 + bold), inativas em cinza (#666666). Separadores │ entre tabs.
- **Navegação**: Tab key cicla + teclas 1-5 para acesso direto. Números exibidos nas tabs.
- **Layout**: Tab bar entre título global e conteúdo da view, com separadores │.
- **Responsividade**: Abreviar nomes em terminais < 80 cols (Dash, Bkps, Rest, Sett, Logs).
- **FileBrowser**: Removido do tab cycling — só acessível como sub-componente.
- **Testes**: Testes unitários para componente tab bar.

### Research Findings
- `internal/tui/components/` existe mas está vazio — pronto para o TabBar
- ViewState enum: DashboardView=0, FileBrowserView=1, BackupListView=2, RestoreView=3, SettingsView=4, LogsView=5, SetupView=6
- `viewCount = 6` em update.go:11 — usado no Tab cycling
- Tab cycling é `(m.state + 1) % viewCount` — passa por FileBrowser
- Views usam magic numbers para height (`msg.Height - 6`, `m.height - 10`)
- `propagateWindowSize` passa raw WindowSizeMsg — NÃO subtrai header height
- Settings tem `editMode` e `editingField` para input-consuming state
- BackupList tem `creatingBackup` para password input state
- Restore tem `phase` (0-5) — phase 1 é password input
- Help overlay usa `globalHelp()` e `HelpProvider` interface
- Nenhum conflito com teclas numéricas em nenhuma view atualmente
- Testes existentes seguem padrão table-driven com assertions em View() output

### Metis Review
**Identified Gaps** (addressed):
- **Input-consuming states**: Teclas 1-5 seriam interceptadas durante password input ou edit mode — resolvido com `isInputActive()` guard
- **ViewState enum ordering**: FileBrowserView=1 no meio do enum — resolvido com `tabOrder` slice ao invés de alterar enum
- **Height propagation**: `propagateWindowSize` precisa subtrair tab bar height — resolvido centralizando ajuste
- **Help text sync**: Footer e overlay precisam refletir novos keybindings — incluído como task
- **View title redundancy**: Views renderizam seus próprios títulos E agora terão a tab bar — mantido (usuário não pediu remoção)
- **Dashboard shortcuts b/r/s**: Redundantes com 1-5 — mantidos para retrocompatibilidade
- **Shift+Tab reverse cycling**: Não solicitado — fora de escopo
- **Tab bar no setup mode**: Tab bar NÃO deve aparecer durante setup wizard
- **bubbles list.Model filtering**: `/` para filtrar listas consome keys — coberto pelo `isInputActive()`

---

## Work Objectives

### Core Objective
Adicionar uma barra de tabs visual ao TUI dotkeeper, padronizando a navegação entre seções e fornecendo feedback visual claro da seção atual.

### Concrete Deliverables
- `internal/tui/components/tabbar.go` — componente TabBar
- `internal/tui/components/tabbar_test.go` — testes do componente
- Estilos TabActive, TabInactive, TabSeparator em `views/styles.go`
- `tabOrder` slice + `isInputActive()` em `model.go`/`update.go`
- Tab bar integrada no layout principal (`view.go`)
- Help text atualizado (`help.go` + `view.go` footer)

### Definition of Done
- [x] Tab bar visível em todas as 5 views principais (Dashboard, Backups, Restore, Settings, Logs)
- [x] Tab ativa com underline roxo + bold, inativas em cinza
- [x] Teclas 1-5 navegam direto para cada seção
- [x] Tab cycling pula FileBrowser
- [x] Nomes abreviados quando terminal < 80 cols
- [x] Tab bar NÃO aparece durante Setup wizard
- [x] Teclas numéricas NÃO interferem com input fields (password, edit mode, list filter)
- [x] `make test` passa sem falhas
- [x] `make build` compila sem erros

### Must Have
- Tab bar visual entre título e conteúdo
- Underline na tab ativa + separadores │
- Números 1-5 para navegação direta
- Proteção contra input-consuming states
- Responsividade (< 80 cols → nomes abreviados)

### Must NOT Have (Guardrails)
- NÃO reordenar ou renumerar constantes do ViewState enum (quebraria referências existentes)
- NÃO deletar FileBrowserView do enum ou remover seu código — apenas excluir do cycling
- NÃO fazer TabBar um `tea.Model` com Init/Update — é render puro, sem estado próprio
- NÃO mostrar tab bar durante setup mode (`m.setupMode == true`)
- NÃO adicionar handling de teclas numéricas dentro das views individuais — interceptar no main model
- NÃO remover shortcuts do Dashboard (b/r/s) — manter para retrocompatibilidade
- NÃO adicionar animações, suporte a mouse, badges, tabs fecháveis, ou persistência de tab
- NÃO mexer na lógica interna de nenhuma view (backup flow, restore phases, etc.)
- NÃO remover títulos individuais das views (Dashboard, Settings, etc.)

---

## Verification Strategy (MANDATORY)

> **UNIVERSAL RULE: ZERO HUMAN INTERVENTION**
>
> ALL tasks MUST be verifiable WITHOUT any human action.

### Test Decision
- **Infrastructure exists**: YES
- **Automated tests**: YES (Tests-after)
- **Framework**: Go testing + `make test`

### Agent-Executed QA Scenarios (MANDATORY — ALL tasks)

**Verification Tool by Deliverable Type:**

| Type | Tool | How Agent Verifies |
|------|------|-------------------|
| **Tab bar component** | Bash (`go test`) | Unit tests assert View() output contains expected strings |
| **Navigation logic** | Bash (`go test`) | Unit tests simulate key presses and assert state changes |
| **TUI visual** | interactive_bash (tmux) | Launch TUI, visually verify tab bar renders, screenshot |
| **Build** | Bash (`make build && make test`) | Exit code 0 |

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately):
├── Task 1: Add tab styles to styles.go
└── Task 3: Add tabOrder + isInputActive to model.go

Wave 2 (After Wave 1):
├── Task 2: Create TabBar component (depends: 1)
└── Task 4: Wire navigation logic in update.go (depends: 3)

Wave 3 (After Wave 2):
├── Task 5: Integrate tab bar in view.go + adjust heights (depends: 2, 4)
├── Task 6: Update help text (depends: 5)
└── Task 7: Write tests + final verification (depends: 2, 4, 5, 6)

Critical Path: Task 1 → Task 2 → Task 5 → Task 7
Parallel Speedup: ~35% faster than sequential
```

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|------|------------|--------|---------------------|
| 1 | None | 2 | 3 |
| 2 | 1 | 5, 7 | 4 |
| 3 | None | 4 | 1 |
| 4 | 3 | 5, 7 | 2 |
| 5 | 2, 4 | 6, 7 | None |
| 6 | 5 | 7 | None |
| 7 | 2, 4, 5, 6 | None | None (final) |

### Agent Dispatch Summary

| Wave | Tasks | Recommended Agents |
|------|-------|-------------------|
| 1 | 1, 3 | delegate_task(category="quick", load_skills=[], run_in_background=false) |
| 2 | 2, 4 | delegate_task(category="quick", load_skills=[], run_in_background=false) |
| 3 | 5, 6, 7 | delegate_task(category="unspecified-low", load_skills=[], run_in_background=false) |

---

## TODOs

- [x] 1. Adicionar estilos de tab ao Styles struct

  **What to do**:
  - Adicionar 3 novos estilos ao struct `Styles` em `views/styles.go`:
    - `TabActive`: Bold + Foreground(`#7D56F4`) + Underline(true) — para a tab ativa
    - `TabInactive`: Foreground(`#666666`) — para tabs inativas
    - `TabSeparator`: Foreground(`#444444`) — para o caractere │ entre tabs
  - Instanciar os 3 estilos em `DefaultStyles()`

  **Must NOT do**:
  - NÃO modificar nenhum estilo existente
  - NÃO adicionar MarginLeft aos estilos de tab (o componente gerencia margem)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `[]`
    - Tarefa trivial: adicionar 3 campos a um struct Go e instanciá-los

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 3)
  - **Blocks**: Task 2
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/tui/views/styles.go:6-17` — Struct `Styles` com 10 estilos existentes. Adicionar TabActive, TabInactive, TabSeparator seguindo o mesmo padrão.
  - `internal/tui/views/styles.go:20-50` — `DefaultStyles()` onde instanciar os novos estilos. Usar mesma estrutura de `lipgloss.NewStyle().Attr(value)`.

  **WHY Each Reference Matters**:
  - `styles.go:6-17`: Mostra a convenção de naming e tipos usados no struct. Seguir exatamente.
  - `styles.go:20-50`: Mostra como cada estilo é construído com lipgloss. Usar mesmo padrão.

  **Acceptance Criteria**:

  - [x] `styles.go` compila: `go build ./internal/tui/views/`
  - [x] `Styles` struct contém campos `TabActive`, `TabInactive`, `TabSeparator` do tipo `lipgloss.Style`
  - [x] `DefaultStyles()` retorna os 3 estilos com cores corretas (#7D56F4, #666666, #444444)
  - [x] Testes existentes continuam passando: `go test ./internal/tui/views/...`

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Styles struct has new tab fields
    Tool: Bash (go build)
    Preconditions: None
    Steps:
      1. Run: go build ./internal/tui/views/
      2. Assert: exit code 0
      3. Run: go test ./internal/tui/views/...
      4. Assert: exit code 0, all tests pass
    Expected Result: Build and tests pass
    Evidence: Command output captured
  ```

  **Commit**: YES (groups with Task 2)
  - Message: `feat(tui): add tab bar styles to shared Styles struct`
  - Files: `internal/tui/views/styles.go`
  - Pre-commit: `go test ./internal/tui/views/...`

---

- [x] 2. Criar componente TabBar

  **What to do**:
  - Criar `internal/tui/components/tabbar.go` com package `components`
  - Definir struct `TabBar` (NÃO tea.Model):
    ```go
    type TabItem struct {
        Key       string // "1", "2", etc.
        Label     string // "Dashboard", "Backups", etc.
        ShortLabel string // "Dash", "Bkps", etc.
    }
    
    type TabBar struct {
        items  []TabItem
        styles views.Styles
    }
    ```
  - Construtor: `func NewTabBar(styles views.Styles) TabBar` que inicializa com os 5 tabs:
    1. Dashboard (Dash)
    2. Backups (Bkps)
    3. Restore (Rest)
    4. Settings (Sett)
    5. Logs (Logs)
  - Método de renderização: `func (t TabBar) View(activeIndex int, width int) string`
    - Se `width < 80`: usar ShortLabel
    - Se `width >= 80`: usar Label
    - Tab ativa: `styles.TabActive` com Underline
    - Tabs inativas: `styles.TabInactive`
    - Separador entre tabs: `styles.TabSeparator.Render(" │ ")`
    - Formato: `1 Dashboard │ 2 Backups │ 3 Restore │ 4 Settings │ 5 Logs`
    - MarginLeft(2) no resultado final (consistência com resto do layout)
  - **IMPORTANTE**: `activeIndex` é 0-4 (índice no tabOrder), NÃO o ViewState enum value

  **Must NOT do**:
  - NÃO implementar como tea.Model (sem Init/Update)
  - NÃO armazenar estado (sem activeTab field — recebido como parâmetro)
  - NÃO importar bubbletea — só lipgloss

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `[]`
    - Componente Go puro com rendering lipgloss, sem complexidade

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Task 4)
  - **Blocks**: Tasks 5, 7
  - **Blocked By**: Task 1

  **References**:

  **Pattern References**:
  - `internal/tui/views/styles.go:6-50` — Struct `Styles` com os novos estilos TabActive/TabInactive/TabSeparator. Importar `views` package para acessar.
  - `internal/tui/view.go:50-84` — Layout principal. O TabBar será inserido após o título (line 57). Observar padrão de `strings.Builder` e `MarginLeft(2)`.

  **API/Type References**:
  - `internal/tui/model.go:16-24` — ViewState enum. O TabBar NÃO usa ViewState diretamente — recebe `activeIndex int` (0-4).

  **External References**:
  - Lipgloss docs: `https://github.com/charmbracelet/lipgloss` — JoinHorizontal, Underline, Foreground.

  **WHY Each Reference Matters**:
  - `styles.go`: Fornece os estilos prontos para usar no rendering. Não criar estilos inline.
  - `view.go`: Mostra o padrão de rendering do layout para manter consistência visual (MarginLeft, strings.Builder).
  - `model.go`: Entender que ViewState enum values NÃO são 0-4 contíguas (FileBrowser=1 no meio). O TabBar usa activeIndex simples.

  **Acceptance Criteria**:

  - [x] Arquivo `internal/tui/components/tabbar.go` criado
  - [x] Package é `components`, não `views` ou `tui`
  - [x] `TabBar` NÃO implementa `tea.Model` (sem Init/Update/View com receiver)
  - [x] `View(activeIndex, width)` renderiza 5 tabs com separadores
  - [x] Tab ativa usa underline + bold + cor roxa
  - [x] Tabs inativas usam cor cinza
  - [x] `width < 80` usa labels abreviados (Dash, Bkps, Rest, Sett, Logs)
  - [x] `width >= 80` usa labels completos
  - [x] Compila: `go build ./internal/tui/components/`

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: TabBar renders with full labels at width 100
    Tool: Bash (go test)
    Preconditions: Task 1 complete (styles available)
    Steps:
      1. Create tabbar_test.go with test:
         tb := NewTabBar(views.DefaultStyles())
         output := tb.View(0, 100)
      2. Assert: output contains "Dashboard"
      3. Assert: output contains "Backups"
      4. Assert: output contains "│"
      5. Assert: output contains "1"
      6. Run: go test ./internal/tui/components/...
      7. Assert: exit code 0
    Expected Result: All labels present with separators
    Evidence: Test output captured

  Scenario: TabBar renders abbreviated at width 60
    Tool: Bash (go test)
    Preconditions: Task 1 complete
    Steps:
      1. Test: tb.View(0, 60)
      2. Assert: output contains "Dash"
      3. Assert: output does NOT contain "Dashboard"
      4. Assert: output contains "Bkps"
    Expected Result: Abbreviated labels used
    Evidence: Test output captured

  Scenario: Active tab has different styling
    Tool: Bash (go test)
    Preconditions: Task 1 complete
    Steps:
      1. Test: output0 := tb.View(0, 100) // Dashboard active
      2. Test: output2 := tb.View(2, 100) // Restore active
      3. Assert: output0 != output2 (different tabs highlighted)
    Expected Result: Active tab changes with index
    Evidence: Test output captured
  ```

  **Commit**: YES (with Task 1)
  - Message: `feat(tui): create TabBar component with responsive rendering`
  - Files: `internal/tui/components/tabbar.go`, `internal/tui/components/tabbar_test.go`
  - Pre-commit: `go test ./internal/tui/...`

---

- [x] 3. Adicionar tabOrder, isInputActive e constantes ao main model

  **What to do**:
  - Em `internal/tui/model.go`, adicionar:
    ```go
    // tabOrder defines the views accessible via tabs (excludes FileBrowser and Setup)
    var tabOrder = []ViewState{DashboardView, BackupListView, RestoreView, SettingsView, LogsView}
    ```
  - Adicionar constante: `const tabBarHeight = 2` (1 linha tabs + 1 linha espaçamento)
  - Em `internal/tui/model.go` ou `internal/tui/update.go`, adicionar método:
    ```go
    func (m Model) isInputActive() bool
    ```
    Que retorna `true` quando a view atual está consumindo keyboard input:
    - `m.state == SettingsView && m.settings.IsEditing()` (settings em edit mode ou editando campo)
    - `m.state == BackupListView && m.backupList.IsCreating()` (criando backup com password input)
    - `m.state == RestoreView && m.restore.IsInputActive()` (restore em phase 1 password ou qualquer input)
  - Adicionar métodos públicos nas views para expor estado de input:
    - `views/settings.go`: `func (m SettingsModel) IsEditing() bool { return m.editMode || m.editingField }`
    - `views/backuplist.go`: `func (m BackupListModel) IsCreating() bool { return m.creatingBackup }`
    - `views/restore.go`: `func (m RestoreModel) IsInputActive() bool { return m.phase == 1 }` (password phase)
  - Adicionar método para obter tabIndex da ViewState atual:
    ```go
    func (m Model) activeTabIndex() int {
        for i, v := range tabOrder {
            if v == m.state {
                return i
            }
        }
        return 0 // fallback to Dashboard
    }
    ```

  **Must NOT do**:
  - NÃO alterar valores do ViewState enum
  - NÃO alterar `viewCount` ainda (será ajustado na Task 4)
  - NÃO exportar `tabOrder` fora do package — manter como var não-exportada

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `[]`
    - Adições simples de métodos e variáveis a structs existentes

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 1)
  - **Blocks**: Task 4
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/tui/model.go:16-24` — ViewState enum. O `tabOrder` slice referencia esses valores. NÃO reordenar o enum.
  - `internal/tui/model.go:27-48` — Model struct. O `isInputActive()` acessa sub-models via campos existentes.
  - `internal/tui/update.go:11` — `const viewCount = 6`. Será ajustado na Task 4, não tocar aqui.

  **API/Type References**:
  - `internal/tui/views/settings.go:18-22` — Campos `editMode`, `editingField` (não exportados). Criar método `IsEditing()` para expor.
  - `internal/tui/views/backuplist.go:43` — Campo `creatingBackup` (não exportado). Criar método `IsCreating()` para expor.
  - `internal/tui/views/restore.go:26` — Campo `phase` (não exportado). Criar método `IsInputActive()` para expor.

  **WHY Each Reference Matters**:
  - `model.go:16-24`: Os valores do enum são usados no `tabOrder` — crucial não alterar a ordem.
  - `settings.go:18-22`: `editMode` e `editingField` determinam se o Settings está consumindo input. Precisam de getter público.
  - `backuplist.go:43`: `creatingBackup` determina se password input está ativo.
  - `restore.go:26`: `phase == 1` é o estado de password input no restore flow.

  **Acceptance Criteria**:

  - [x] `tabOrder` definido com 5 ViewStates na ordem correta
  - [x] `tabBarHeight` constante definida como 2
  - [x] `isInputActive()` retorna true para: Settings edit mode, BackupList creating, Restore password phase
  - [x] `activeTabIndex()` retorna índice 0-4 correto para cada ViewState em tabOrder
  - [x] `activeTabIndex()` retorna 0 (fallback) para ViewStates fora do tabOrder (FileBrowser, Setup)
  - [x] Métodos `IsEditing()`, `IsCreating()`, `IsInputActive()` exportados nas views
  - [x] Compila: `go build ./internal/tui/...`
  - [x] Testes existentes passam: `go test ./internal/tui/...`

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Build succeeds with new additions
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run: go build ./internal/tui/...
      2. Assert: exit code 0
      3. Run: go test ./internal/tui/...
      4. Assert: exit code 0
    Expected Result: No compile errors, existing tests pass
    Evidence: Command output captured
  ```

  **Commit**: NO (groups with Task 4)

---

- [x] 4. Implementar navegação por números e ajustar Tab cycling

  **What to do**:
  - Em `internal/tui/update.go`:
    - **Remover** a constante `viewCount = 6`
    - **Substituir** o Tab cycling para usar `tabOrder` ao invés de modulo aritmético:
      ```go
      if key.Matches(msg, keys.Tab) {
          if !m.isInputActive() {
              currentIdx := m.activeTabIndex()
              nextIdx := (currentIdx + 1) % len(tabOrder)
              prevState := m.state
              m.state = tabOrder[nextIdx]
              // ... refresh triggers mantidos iguais ...
          }
          return m, tea.Batch(cmds...)
      }
      ```
    - **Adicionar** handling de teclas numéricas 1-5 APÓS o Tab handling e ANTES do dispatch para views (linha ~130, antes dos dashboard shortcuts):
      ```go
      // Number key navigation (only when not in input-consuming state)
      if !m.isInputActive() {
          switch msg.String() {
          case "1":
              m.state = tabOrder[0] // Dashboard
              return m, nil
          case "2":
              m.state = tabOrder[1] // BackupList
              return m, m.backupList.Refresh()
          case "3":
              m.state = tabOrder[2] // Restore
              return m, m.restore.Refresh()
          case "4":
              m.state = tabOrder[3] // Settings
              return m, nil
          case "5":
              m.state = tabOrder[4] // Logs
              return m, m.logs.LoadHistory()
          }
      }
      ```
    - **Manter** os shortcuts do Dashboard (b/r/s) — NÃO remover
    - **Adicionar** refresh triggers para as views que precisam (BackupList, Restore, Logs) ao navegar por número, mesma lógica do Tab cycling existente

  **Must NOT do**:
  - NÃO alterar ViewState enum
  - NÃO remover shortcuts do Dashboard (b/r/s)
  - NÃO adicionar handling de números dentro das views individuais
  - NÃO mudar a ordem dos handlers (global keys → tab → numbers → dashboard shortcuts → view dispatch)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `[]`
    - Modificação localizada em update.go com lógica de navegação

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Task 2)
  - **Blocks**: Tasks 5, 7
  - **Blocked By**: Task 3

  **References**:

  **Pattern References**:
  - `internal/tui/update.go:115-128` — Tab cycling atual com modulo. Substituir para usar `tabOrder` slice.
  - `internal/tui/update.go:130-143` — Dashboard shortcuts (b/r/s). Manter inalterados. Os number keys vão ANTES deste bloco.
  - `internal/tui/update.go:99-128` — Fluxo de key handling global. Os number keys devem seguir o mesmo padrão (handled at main level, before view dispatch).

  **API/Type References**:
  - `internal/tui/model.go` — `tabOrder` slice e `isInputActive()` (criados na Task 3).

  **WHY Each Reference Matters**:
  - `update.go:115-128`: Este é o código EXATO que será substituído. Entender a lógica de refresh triggers para replicar.
  - `update.go:130-143`: Estes shortcuts NÃO devem ser tocados. Os number keys vão ANTES na chain.
  - `model.go`: `tabOrder` fornece o mapeamento index→ViewState. `isInputActive()` é o guard.

  **Acceptance Criteria**:

  - [x] `viewCount` constante removida de update.go
  - [x] Tab cycling usa `tabOrder` e `activeTabIndex()` — NÃO mais modulo aritmético
  - [x] Tab cycling pula FileBrowserView (nunca para nela)
  - [x] Teclas 1-5 navegam para a view correspondente no tabOrder
  - [x] Teclas 1-5 NÃO funcionam quando `isInputActive()` retorna true
  - [x] Tab key NÃO funciona quando `isInputActive()` retorna true
  - [x] Refresh triggers mantidos: BackupList.Refresh(), Restore.Refresh(), Logs.LoadHistory()
  - [x] Dashboard shortcuts (b/r/s) continuam funcionando
  - [x] Compila: `go build ./internal/tui/...`
  - [x] Testes existentes passam: `go test ./internal/tui/...`

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Tab cycles through 5 views only
    Tool: Bash (go test)
    Preconditions: Task 3 complete
    Steps:
      1. Create test in update_test.go:
         model := NewModel() // starts at DashboardView
         // Press Tab 5 times, record states
         // Assert: states are BackupList, Restore, Settings, Logs, Dashboard
         // Assert: FileBrowserView NEVER appears
      2. Run: go test ./internal/tui/...
      3. Assert: exit code 0
    Expected Result: FileBrowser never in cycle
    Evidence: Test output captured

  Scenario: Number keys blocked during input
    Tool: Bash (go test)
    Preconditions: Task 3 complete
    Steps:
      1. Create test:
         model := NewModel()
         model.state = SettingsView
         // Set settings to edit mode (need exported method)
         // Send key "1"
         // Assert: model.state is still SettingsView
      2. Run: go test ./internal/tui/...
      3. Assert: exit code 0
    Expected Result: Number keys don't switch views during input
    Evidence: Test output captured
  ```

  **Commit**: YES (with Task 3)
  - Message: `feat(tui): implement tab navigation with number keys and input guards`
  - Files: `internal/tui/model.go`, `internal/tui/update.go`, `internal/tui/views/settings.go`, `internal/tui/views/backuplist.go`, `internal/tui/views/restore.go`
  - Pre-commit: `go test ./internal/tui/...`

---

- [x] 5. Integrar TabBar no layout principal e ajustar heights

  **What to do**:
  - Em `internal/tui/view.go`:
    - Importar `"github.com/diogo/dotkeeper/internal/tui/components"`
    - Criar instância do TabBar (pode ser field no Model ou criado inline no View — preferir inline por simplicidade, ou field se performance importar):
      ```go
      // No início do View(), após o check de setupMode e quitting:
      tabBar := components.NewTabBar(views.DefaultStyles())
      ```
    - Inserir renderização do TabBar ENTRE o título e o conteúdo:
      ```go
      b.WriteString(titleStyle.Render("dotkeeper - Dotfiles Backup Manager"))
      b.WriteString("\n")
      b.WriteString(tabBar.View(m.activeTabIndex(), m.width))
      b.WriteString("\n\n")
      // ... conteúdo da view ...
      ```
    - **Remover** o case `FileBrowserView` do switch de rendering em View() (lines 64-65) — se o usuário chegar nesse state por algum motivo, fazer fallback para Dashboard
    - **Remover** o case `FileBrowserView` do switch de help overlay (lines 26-29)
  - Em `internal/tui/update.go`:
    - Em `propagateWindowSize()`, subtrair `tabBarHeight` do `msg.Height` ANTES de propagar para as views:
      ```go
      func (m *Model) propagateWindowSize(msg tea.WindowSizeMsg) {
          // Subtract tab bar height from available space for views
          viewMsg := tea.WindowSizeMsg{
              Width:  msg.Width,
              Height: msg.Height - tabBarHeight,
          }
          // Use viewMsg instead of msg for all view updates
          // ...
      }
      ```
    - Manter a propagação para FileBrowser (pode ser usada como sub-componente) mas com o viewMsg ajustado

  **Must NOT do**:
  - NÃO renderizar tab bar quando `m.setupMode == true` (já retorna early em View())
  - NÃO modificar os View() das views individuais — o ajuste de height é via WindowSizeMsg
  - NÃO deletar o fileBrowser field do Model struct — ele pode ser usado como sub-componente
  - NÃO criar nova instância de TabBar em cada chamada de View() se isso causar alocação desnecessária — considerar como field no Model

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
  - **Skills**: `[]`
    - Integração entre componentes existentes, requer atenção ao layout

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (sequential)
  - **Blocks**: Tasks 6, 7
  - **Blocked By**: Tasks 2, 4

  **References**:

  **Pattern References**:
  - `internal/tui/view.go:10-85` — Layout principal COMPLETO. Este é o arquivo principal a modificar. A tab bar vai entre line 57 (`\n\n` após título) e line 59 (content).
  - `internal/tui/view.go:50-57` — Title rendering. A tab bar será inserida DEPOIS desta seção.
  - `internal/tui/view.go:59-74` — Content rendering switch. Remover FileBrowserView case.
  - `internal/tui/view.go:19-46` — Help overlay switch. Remover FileBrowserView case.
  - `internal/tui/update.go:38-58` — `propagateWindowSize()`. Ajustar para subtrair tabBarHeight.

  **API/Type References**:
  - `internal/tui/components/tabbar.go` — `NewTabBar(styles)` e `View(activeIndex, width)` (criado na Task 2).
  - `internal/tui/model.go` — `activeTabIndex()` e `tabBarHeight` (criados na Task 3).

  **WHY Each Reference Matters**:
  - `view.go:10-85`: Este é o arquivo CENTRAL da mudança. Toda a integração visual acontece aqui.
  - `update.go:38-58`: Ajuste de height é CRÍTICO — sem isso, views terão overlap com a tab bar.
  - `components/tabbar.go`: API do componente que será chamada no View().

  **Acceptance Criteria**:

  - [x] Tab bar renderizada entre título e conteúdo em todas as 5 views
  - [x] Tab bar NÃO renderizada durante setup mode
  - [x] FileBrowserView removida do switch de rendering em View()
  - [x] FileBrowserView removida do switch de help overlay
  - [x] `propagateWindowSize` subtrai `tabBarHeight` antes de propagar
  - [x] Views recebem height correto (sem overlap com tab bar)
  - [x] Compila: `go build ./cmd/dotkeeper/...`
  - [x] Testes passam: `go test ./internal/tui/...`

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: TUI renders with tab bar visible
    Tool: interactive_bash (tmux)
    Preconditions: make build succeeds
    Steps:
      1. tmux new-session: ./bin/dotkeeper
      2. Wait for: TUI to render (timeout: 3s)
      3. Assert: "Dashboard" visible in output (tab bar)
      4. Assert: "│" visible (separators)
      5. Assert: "1" visible (number labels)
      6. Send keys: "q" to quit
      7. Assert: Process exited cleanly
    Expected Result: Tab bar visible at top with all 5 tabs
    Evidence: Terminal output captured

  Scenario: Build and test pass after integration
    Tool: Bash
    Preconditions: Tasks 1-4 complete
    Steps:
      1. Run: make build
      2. Assert: exit code 0
      3. Run: make test
      4. Assert: exit code 0
    Expected Result: Clean build and all tests pass
    Evidence: Command output captured
  ```

  **Commit**: YES
  - Message: `feat(tui): integrate tab bar into main layout with height adjustment`
  - Files: `internal/tui/view.go`, `internal/tui/update.go`
  - Pre-commit: `make test`

---

- [x] 6. Atualizar help text (footer e overlay)

  **What to do**:
  - Em `internal/tui/view.go`, line 81, atualizar o footer help text:
    ```go
    // Antes:
    b.WriteString(helpStyle.Render("Tab: switch views | q: quit | ?: help"))
    // Depois:
    b.WriteString(helpStyle.Render("Tab/1-5: switch views | q: quit | ?: help"))
    ```
  - Em `internal/tui/help.go`, atualizar `globalHelp()`:
    ```go
    func globalHelp() []views.HelpEntry {
        return []views.HelpEntry{
            {Key: "Tab", Description: "Next view"},
            {Key: "1-5", Description: "Go to view"},
            {Key: "q", Description: "Quit"},
            {Key: "?", Description: "Toggle help"},
        }
    }
    ```

  **Must NOT do**:
  - NÃO alterar o layout/estilo do help overlay
  - NÃO adicionar entries view-specific (já existem via HelpProvider)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `[]`
    - Duas edições de string triviais

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (after Task 5)
  - **Blocks**: Task 7
  - **Blocked By**: Task 5

  **References**:

  **Pattern References**:
  - `internal/tui/view.go:78-82` — Footer help text. String a ser atualizada.
  - `internal/tui/help.go:11-17` — `globalHelp()`. Adicionar entry para "1-5".

  **WHY Each Reference Matters**:
  - `view.go:78-82`: Texto exato que o usuário vê no footer. Precisa refletir novos keybindings.
  - `help.go:11-17`: Help overlay entries. Adicionar "1-5" para consistência.

  **Acceptance Criteria**:

  - [x] Footer mostra "Tab/1-5: switch views | q: quit | ?: help"
  - [x] Help overlay inclui entry "1-5: Go to view"
  - [x] Help overlay mantém "Tab: Next view" (atualizado de "Switch views")
  - [x] Compila: `go build ./internal/tui/...`

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Footer shows updated keybindings
    Tool: Bash (grep)
    Preconditions: Changes applied
    Steps:
      1. Search view.go for "1-5"
      2. Assert: found in footer text
      3. Search help.go for "Go to view"
      4. Assert: found in globalHelp entries
      5. Run: go build ./internal/tui/...
      6. Assert: exit code 0
    Expected Result: New keybinding text present
    Evidence: Search results captured
  ```

  **Commit**: YES (groups with Task 5)
  - Message: `feat(tui): update help text with tab navigation keybindings`
  - Files: `internal/tui/view.go`, `internal/tui/help.go`
  - Pre-commit: `make test`

---

- [x] 7. Escrever testes unitários e verificação final

  **What to do**:
  - Criar `internal/tui/components/tabbar_test.go` com testes:
    ```go
    func TestTabBarActiveHighlight(t *testing.T)
    // Testar que tab ativa tem styling diferente das inativas
    // Renderizar com activeIndex=0 e activeIndex=2, comparar outputs
    
    func TestTabBarResponsive(t *testing.T)
    // Testar width < 80: labels abreviados (Dash, Bkps, Rest, Sett, Logs)
    // Testar width >= 80: labels completos (Dashboard, Backups, Restore, Settings, Logs)
    
    func TestTabBarSeparators(t *testing.T)
    // Testar que │ aparece 4 vezes (entre 5 tabs)
    
    func TestTabBarNumberLabels(t *testing.T)
    // Testar que "1", "2", "3", "4", "5" aparecem no output
    
    func TestTabBarAllTabs(t *testing.T)
    // Testar que todos os 5 nomes aparecem no output
    ```
  - Criar ou atualizar `internal/tui/update_test.go` com testes de navegação:
    ```go
    func TestTabCycleSkipsFileBrowser(t *testing.T)
    // Ciclar Tab 5 vezes e verificar que FileBrowserView nunca é o state
    
    func TestNumberKeysNavigate(t *testing.T)
    // Testar que "1" vai para Dashboard, "2" para BackupList, etc.
    
    func TestNumberKeysBlockedDuringInput(t *testing.T)
    // Colocar settings em edit mode, pressionar "1", verificar que state não muda
    
    func TestTabBlockedDuringInput(t *testing.T)
    // Colocar backuplist em creating mode, pressionar Tab, verificar que state não muda
    ```
  - Rodar `make test` e `make build` para verificação final
  - Rodar `make lint` se disponível

  **Must NOT do**:
  - NÃO testar visual rendering pixel-perfect — apenas presença de strings esperadas
  - NÃO mockar lipgloss (framework de styling)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
  - **Skills**: `[]`
    - Testes Go padrão, sem complexidade especial

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (final task)
  - **Blocks**: None (final)
  - **Blocked By**: Tasks 2, 4, 5, 6

  **References**:

  **Test References**:
  - `internal/tui/views/dashboard_test.go:1-33` — Padrão de teste existente. Usa `config.Config` mock, chama NewX(), testa View() com `strings.Contains`. Seguir este padrão.
  - `internal/tui/views/settings_test.go` — Testes de view com estado. Referência para testar interações.
  - `internal/tui/views/backuplist_test.go` — Testes de view com mock store.

  **Pattern References**:
  - `internal/tui/model.go:50-80` — `NewModel()` para criar model completo nos testes de navegação.
  - `internal/tui/update.go:60-180` — `Update()` para simular key presses nos testes.

  **WHY Each Reference Matters**:
  - `dashboard_test.go`: Padrão de teste oficial do projeto. Copiar estilo (setup, assertions, naming).
  - `model.go:50-80`: Precisa de `NewModel()` para testes de integração — mas isso requer config. Pode ser necessário criar helper de teste.

  **Acceptance Criteria**:

  - [x] `tabbar_test.go` com 5+ testes cobrindo: highlight, responsividade, separadores, números, todos os tabs
  - [x] `update_test.go` com 4+ testes cobrindo: tab cycling sem FileBrowser, number keys, input blocking
  - [x] `make test` → exit code 0, todas as suítes passam
  - [x] `make build` → exit code 0
  - [x] `make lint` → sem novos warnings (se disponível) [NOTE: pre-existing lint issues exist, none from our changes]
  - [x] Zero testes falhando nos arquivos existentes

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: All tests pass including new ones
    Tool: Bash
    Preconditions: All previous tasks complete
    Steps:
      1. Run: make test
      2. Assert: exit code 0
      3. Assert: output shows tabbar tests passing
      4. Assert: output shows update tests passing
      5. Run: make build
      6. Assert: exit code 0
      7. Assert: ./bin/dotkeeper binary exists
    Expected Result: Full test suite green, binary builds
    Evidence: Test output and build output captured

  Scenario: TUI launches and tabs work end-to-end
    Tool: interactive_bash (tmux)
    Preconditions: make build succeeds
    Steps:
      1. tmux new-session: ./bin/dotkeeper
      2. Wait for: TUI renders (timeout: 3s)
      3. Assert: Tab bar visible with "Dashboard" highlighted
      4. Send keys: "2"
      5. Wait for: 0.5s
      6. Assert: "Backups" tab is now active
      7. Send keys: Tab
      8. Wait for: 0.5s
      9. Assert: "Restore" tab is now active (next after Backups)
      10. Send keys: "q"
      11. Assert: Process exited cleanly
    Expected Result: Tab navigation works via numbers and Tab key
    Evidence: Terminal output captured at each step
  ```

  **Commit**: YES
  - Message: `test(tui): add tab bar component and navigation tests`
  - Files: `internal/tui/components/tabbar_test.go`, `internal/tui/update_test.go`
  - Pre-commit: `make test`

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 1+2 | `feat(tui): add tab bar styles and create TabBar component` | `views/styles.go`, `components/tabbar.go`, `components/tabbar_test.go` | `go test ./internal/tui/...` |
| 3+4 | `feat(tui): implement tab navigation with number keys and input guards` | `model.go`, `update.go`, `views/settings.go`, `views/backuplist.go`, `views/restore.go` | `go test ./internal/tui/...` |
| 5+6 | `feat(tui): integrate tab bar into layout and update help text` | `view.go`, `update.go`, `help.go` | `make test` |
| 7 | `test(tui): add tab bar component and navigation tests` | `components/tabbar_test.go`, `update_test.go` | `make test && make build` |

---

## Success Criteria

### Verification Commands
```bash
make build   # Expected: exit code 0, binary at ./bin/dotkeeper
make test    # Expected: exit code 0, all tests pass (including new ones)
make lint    # Expected: no new warnings
```

### Final Checklist
- [x] Tab bar visível em todas as 5 views (Dashboard, Backups, Restore, Settings, Logs)
- [x] Tab ativa com underline roxo + bold
- [x] Tabs inativas em cinza com separadores │
- [x] Números 1-5 exibidos em cada tab
- [x] Teclas 1-5 navegam para a view correspondente
- [x] Tab cycling pula FileBrowser
- [x] Números e Tab bloqueados durante input-consuming states
- [x] Tab bar NÃO aparece durante Setup wizard
- [x] Responsividade: nomes abreviados em terminais < 80 cols
- [x] Help footer e overlay atualizados com "1-5"
- [x] Todos os testes existentes passam (zero regressão)
- [x] Novos testes cobrem: componente + navegação + input blocking
- [x] `make build && make test` → sucesso
