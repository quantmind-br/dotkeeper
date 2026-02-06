# Plano de Cobertura de Testes - Meta: 90%

## Status Atual

### Cobertura Geral
- **Cobertura atual**: ~44.5%
- **Meta**: 90%
- **Gap**: ~45.5%

### Cobertura por Pacote/Módulo

| Pacote | Cobertura Atual | Meta | Gap | Prioridade |
|--------|----------------|------|-----|------------|
| internal/pathutil | 92.7% | 90% | ✅ Atingido | - |
| internal/crypto | 84.6% | 90% | 5.4% | Média |
| internal/keyring | 81.8% | 90% | 8.2% | Média |
| internal/backup | 79.5% | 90% | 10.5% | Alta |
| internal/restore | 79.0% | 90% | 11.0% | Alta |
| internal/history | 78.5% | 90% | 11.5% | Média |
| internal/notify | 75.0% | 90% | 15.0% | Baixa |
| internal/git | 65.1% | 90% | 24.9% | Alta |
| internal/config | 60.3% | 90% | 29.7% | Alta |
| internal/tui/components | 55.0% | 90% | 35.0% | Média |
| internal/tui/views | 43.9% | 90% | 46.1% | Média |
| internal/cli | 10.2% | 90% | 79.8% | Alta |
| cmd/dotkeeper | 0.0% | N/A | - | Baixa |
| internal/tui | 0.0% | 90% | 90.0% | Média |
| internal/tui/styles | 0.0% | N/A | - | Baixa |

## Análise Detalhada

### Arquivos Sem Testes (0% de Cobertura)

#### 1. `cmd/dotkeeper/main.go`
- **Funções**: `main()`, `printHelp()`
- **Análise**: Entry point da aplicação. Difícil de testar unitariamente, mas pode ser testado via E2E
- **Estratégia**: Expandir testes E2E para cobrir diferentes argumentos de linha de comando

#### 2. `internal/tui/model.go`
- **Funções**: `NewModel()`, `Init()`, `GetConfig()`, `activeTabIndex()`, `isInputActive()`
- **Análise**: Core do TUI BubbleTea. Inicialização de modelos e estado
- **Estratégia**: Testes de integração do modelo

#### 3. `internal/tui/update.go`
- **Funções**: `DefaultKeyMap()`, `propagateWindowSize()`, `refreshCmdForState()`, `Update()`
- **Análise**: Loop principal de atualização do TUI. Lógica complexa de roteamento de mensagens
- **Estratégia**: Testes com mensagens simuladas (tea.KeyMsg, tea.WindowSizeMsg)

#### 4. `internal/tui/view.go`
- **Funções**: `View()`, `currentViewHelpText()`, `currentViewHelp()`
- **Análise**: Renderização das views
- **Estratégia**: Testes de renderização (verificar output string)

#### 5. `internal/tui/help.go`
- **Funções**: `globalHelp()`, `renderHelpOverlay()`
- **Análise**: Sistema de ajuda global
- **Estratégia**: Testes de renderização

#### 6. `internal/tui/styles/styles.go`
- **Funções**: `DefaultStyles()`, `NewListDelegate()`, `NewMinimalList()`
- **Análise**: Configuração de estilos Lipgloss - puramente declarativo
- **Estratégia**: Testes simples de verificação de criação

### Arquivos com Baixa Cobertura (< 50%)

#### 1. `internal/cli/backup.go` - 0.0%
**Funções não testadas**:
- [ ] `BackupCommand(args []string) int`
  - Cenários: backup bem-sucedido, erro de config, erro de senha, notificação habilitada
  - Estratégia: Mock de backup.Backup(), config.Load(), notify.SendSuccess/Error
- [ ] `getPassword(passwordFile string) (string, error)`
  - Cenários: arquivo de senha, variável de ambiente, keyring, erro em todos os casos
  - Estratégia: Mock de os.ReadFile, os.Getenv, keyring.Retrieve

#### 2. `internal/cli/list.go` - 0.0%
**Funções não testadas**:
- [ ] `ListCommand(args []string) int`
  - Cenários: listagem vazia, com backups, formato JSON, erro de diretório
  - Estratégia: Mock de filesystem, config.Load()
- [ ] `findBackups(backupDir string) ([]BackupInfo, error)`
  - Cenários: diretório vazio, com arquivos, com metadata, erro de leitura
  - Estratégia: Criar diretório temporário com fixtures
- [ ] `printBackupTable(backups []BackupInfo)`
  - Cenários: output em tabela
  - Estratégia: Capturar stdout
- [ ] `formatSize(size int64) string`
  - Cenários: bytes, KB, MB, GB, TB
  - Estratégia: Testes simples com table-driven tests

#### 3. `internal/cli/config.go` - 0.0%
**Funções não testadas**:
- [ ] `ConfigCommand(args []string) int`
  - Cenários: subcomandos get/set/list, erro de argumentos
- [ ] `configGet(args []string) int`
  - Cenários: chave válida, chave inválida
- [ ] `configSet(args []string) int`
  - Cenários: setar cada tipo de config
- [ ] `configList() int`
  - Cenários: listar configurações
- [ ] `getConfigValue(cfg *config.Config, key string) (string, error)`
  - Cenários: todas as chaves válidas, chave inválida
- [ ] `setConfigValue(cfg *config.Config, key, value string) error`
  - Cenários: todas as chaves, booleanos com diferentes formatos
- [ ] `normalizeKey(key string) string`
  - Cenários: normalização de case e hífens

#### 4. `internal/cli/history.go` - 0.0%
**Funções não testadas**:
- [ ] `HistoryCommand(args []string) int`
  - Cenários: com/sem filtro, JSON/plain
- [ ] `printHistoryTable(entries []history.HistoryEntry)`
  - Cenários: renderização de tabela
- [ ] `formatDuration(ms int64) string`
  - Cenários: 0ms, <1s, >1s

#### 5. `internal/cli/schedule.go` - 0.0%
**Funções não testadas**:
- [ ] `EnableSchedule() error`
- [ ] `DisableSchedule() error`
- [ ] `StatusSchedule() error`
- [ ] `ScheduleCommand(args []string) int`
- [ ] `isSystemdAvailable() bool`
- [ ] `getUserSystemdDir() (string, error)`
- [ ] `runSystemctl(args ...string) error`
- [ ] `copyFile(src, dst string) error`

**Nota**: Estas funções executam comandos do sistema e interagem com systemd. Testes requerem mock de `exec.Command` ou testes de integração.

#### 6. `internal/cli/delete.go` - 0.0%
**Funções não testadas**:
- [ ] `DeleteCommand(args []string) int`
  - Cenários: deletar backup existente, inexistente, erro

#### 7. `internal/tui/views/*.go` - 43.9% média
**Funções com 0% de cobertura**:
- `backuplist.go`: `runBackup()`, `HelpBindings()`, `StatusHelpText()`
- `restore.go`: `Title()`, `Description()`, `FilterValue()`, `Refresh()`, `validatePassword()`, `loadFiles()`, `loadDiff()`, `runRestore()`, `updateFileListSelection()`, `HelpBindings()`, `IsInputActive()`
- `logs.go`: `Title()`, `Description()`, `FilterValue()`, `HelpBindings()`, `formatBytes()`
- `settings.go`: `Title()`, `Description()`, `FilterValue()`, `Init()`, `scanPathDescs()`, `HelpBindings()`, `StatusHelpText()`
- `dashboard.go`: `HelpBindings()`, `Refresh()`, `refreshStatus()`
- `helpers.go`: `FilterValue()`, `ValidateFolderPath()`, `PlaceOverlay()`

### Código Crítico Não Testado

#### Prioridade Alta - Código de Negócio Core

1. **`internal/cli/backup.go:BackupCommand()`**
   - Impacto: Fluxo principal de backup via CLI
   - Risco: Erros aqui afetam todos os backups automatizados
   - Complexidade: Média (múltiplas dependências)

2. **`internal/cli/restore.go:RestoreCommand()`**
   - Impacto: Fluxo de restauração
   - Risco: Perda de dados em caso de bugs
   - Complexidade: Alta (resolução de conflitos)

3. **`internal/tui/update.go:Update()`**
   - Impacto: Loop principal do TUI
   - Risco: Travamentos, comportamentos inesperados
   - Complexidade: Alta (máquina de estados)

4. **`internal/backup/backup.go`** (79.5% - parcial)
   - Impacto: Lógica core de backup
   - Risco: Falhas de backup, corrupção de dados
   - Complexidade: Alta (I/O, criptografia)

5. **`internal/restore/restore.go`** (79.0% - parcial)
   - Impacto: Lógica core de restauração
   - Risco: Perda de dados
   - Complexidade: Alta

## Estratégia de Implementação

### Fase 1: Quick Wins (Semana 1-2)

Objetivo: Aumentar cobertura rapidamente com testes fáceis

#### 1.1 Funções Utilitárias Simples
- [ ] `internal/cli/list.go:formatSize()` - Estimativa: +0.5% cobertura
  - Testes: table-driven para diferentes tamanhos
  - Tempo estimado: 30 minutos

- [ ] `internal/cli/history.go:formatDuration()` - Estimativa: +0.3% cobertura
  - Testes: cenários de ms, segundos
  - Tempo estimado: 20 minutos

- [ ] `internal/cli/config.go:normalizeKey()` - Estimativa: +0.3% cobertura
  - Testes: lowercase, replace de hífen
  - Tempo estimado: 15 minutos

- [ ] `internal/tui/views/logs.go:formatBytes()` - Estimativa: +0.3% cobertura
  - Testes: similar a formatSize
  - Tempo estimado: 20 minutos

#### 1.2 Funções de Configuração
- [ ] `internal/cli/config.go:getConfigValue()` - Estimativa: +1% cobertura
  - Testes: todas as chaves, chave inválida
  - Tempo estimado: 45 minutos

- [ ] `internal/cli/config.go:setConfigValue()` - Estimativa: +1.5% cobertura
  - Testes: todas as chaves, booleanos em diferentes formatos
  - Tempo estimado: 1 hora

- [ ] `internal/config/config.go:GetConfigDir()`, `GetConfigPath()` - Estimativa: +1% cobertura
  - Testes: verificar paths esperados
  - Tempo estimado: 30 minutos

**Total Fase 1**: Estimativa de +5% cobertura geral, ~3.5 horas de trabalho

### Fase 2: Código Crítico CLI (Semana 3-4)

#### 2.1 CLI Backup
- [ ] `internal/cli/backup.go:getPassword()` - Estimativa: +2% cobertura
  - Cenários de teste:
    - Arquivo de senha existente
    - Arquivo de senha inexistente
    - Variável de ambiente DOTKEEPER_PASSWORD
    - Keyring com senha
    - Keyring vazio (erro)
  - Mocks necessários: `os.ReadFile`, `os.Getenv`, `keyring.Retrieve`
  - Tempo estimado: 2 horas

#### 2.2 CLI List
- [ ] `internal/cli/list.go:findBackups()` - Estimativa: +2% cobertura
  - Cenários de teste:
    - Diretório vazio
    - Diretório com arquivos .tar.gz.enc
    - Arquivos com metadata válida
    - Arquivos com metadata inválida
    - Erro de leitura de diretório
  - Estratégia: Diretórios temporários com fixtures
  - Tempo estimado: 1.5 horas

- [ ] `internal/cli/list.go:ListCommand()` - Estimativa: +2% cobertura
  - Cenários: config válida, config inválida, --json flag
  - Tempo estimado: 1.5 horas

#### 2.3 CLI Config
- [ ] `internal/cli/config.go:configList()` - Estimativa: +1% cobertura
  - Cenários: listar todas as configurações
  - Tempo estimado: 45 minutos

- [ ] `internal/cli/config.go:configGet()` - Estimativa: +1.5% cobertura
  - Cenários: chave existente, chave inexistente, sem argumentos
  - Tempo estimado: 1 hora

- [ ] `internal/cli/config.go:configSet()` - Estimativa: +2% cobertura
  - Cenários: setar cada tipo de config, salvar config
  - Tempo estimado: 1.5 horas

**Total Fase 2**: Estimativa de +12% cobertura geral, ~10 horas de trabalho

### Fase 3: TUI Core (Semana 5-6)

#### 3.1 Model e Inicialização
- [ ] `internal/tui/model.go:NewModel()` - Estimativa: +3% cobertura
  - Cenários: modo setup, modo normal, erro de history
  - Mocks: config.Load(), history.NewStore()
  - Tempo estimado: 2 horas

- [ ] `internal/tui/model.go:activeTabIndex()`, `isInputActive()` - Estimativa: +2% cobertura
  - Cenários: diferentes estados de view
  - Tempo estimado: 1 hora

#### 3.2 Update Loop
- [ ] `internal/tui/update.go:DefaultKeyMap()` - Estimativa: +0.5% cobertura
  - Testes simples de criação
  - Tempo estimado: 20 minutos

- [ ] `internal/tui/update.go:propagateWindowSize()` - Estimativa: +2% cobertura
  - Cenários: diferentes tamanhos de janela
  - Tempo estimado: 1.5 horas

- [ ] `internal/tui/update.go:refreshCmdForState()` - Estimativa: +1% cobertura
  - Cenários: cada view state
  - Tempo estimado: 45 minutos

- [ ] `internal/tui/update.go:Update()` - Estimativa: +8% cobertura
  - Cenários:
    - tea.WindowSizeMsg
    - tea.KeyMsg (quit, tab, shift+tab, help, number keys)
    - views.RefreshBackupListMsg
    - views.DashboardNavigateMsg
    - Setup mode / SetupCompleteMsg
  - Complexidade: Alta - requer mensagens BubbleTea simuladas
  - Tempo estimado: 4 horas

#### 3.3 View Rendering
- [ ] `internal/tui/view.go:View()` - Estimativa: +4% cobertura
  - Cenários: setup mode, quitting, terminal pequeno, help overlay, views normais
  - Estratégia: Verificar output string contém elementos esperados
  - Tempo estimado: 2 horas

- [ ] `internal/tui/view.go:currentViewHelpText()`, `currentViewHelp()` - Estimativa: +1% cobertura
  - Cenários: cada view state
  - Tempo estimado: 1 hora

#### 3.4 Help System
- [ ] `internal/tui/help.go:globalHelp()`, `renderHelpOverlay()` - Estimativa: +1% cobertura
  - Testes de renderização
  - Tempo estimado: 1 hora

**Total Fase 3**: Estimativa de +22.5% cobertura geral, ~13.5 horas de trabalho

### Fase 4: TUI Views (Semana 7-8)

#### 4.1 Views - Métodos Interface
- [ ] Implementar testes para métodos Title(), Description(), FilterValue() em todas as views
  - Arquivos: backuplist.go, restore.go, logs.go, settings.go, helpers.go
  - Estimativa: +3% cobertura
  - Tempo estimado: 2 horas

#### 4.2 Views - Help Bindings
- [ ] `internal/tui/views/dashboard.go:HelpBindings()` - Estimativa: +0.5%
- [ ] `internal/tui/views/backuplist.go:HelpBindings()` - Estimativa: +0.5%
- [ ] `internal/tui/views/restore.go:HelpBindings()` - Estimativa: +0.5%
- [ ] `internal/tui/views/settings.go:HelpBindings()` - Estimativa: +0.5%
- [ ] `internal/tui/views/logs.go:HelpBindings()` - Estimativa: +0.5%
  - Tempo estimado: 1.5 horas total

#### 4.3 Views - Status Help Text
- [ ] `internal/tui/views/backuplist.go:StatusHelpText()` - Estimativa: +0.5%
- [ ] `internal/tui/views/restore.go:StatusHelpText()` - Estimativa: +0.5%
- [ ] `internal/tui/views/settings.go:StatusHelpText()` - Estimativa: +0.5%
- [ ] `internal/tui/views/logs.go:StatusHelpText()` - Estimativa: +0.5%
  - Tempo estimado: 1 hora total

#### 4.4 Views - Funções de Ação
- [ ] `internal/tui/views/backuplist.go:runBackup()` - Estimativa: +2% cobertura
  - Cenários: backup sucesso, backup erro
  - Mocks: backup.Backup()
  - Tempo estimado: 1.5 horas

- [ ] `internal/tui/views/restore.go:runRestore()` - Estimativa: +2% cobertura
  - Cenários: restauração sucesso, erro de senha, erro de restore
  - Tempo estimado: 1.5 horas

- [ ] `internal/tui/views/restore.go:validatePassword()` - Estimativa: +1% cobertura
  - Cenários: senha válida, senha inválida
  - Tempo estimado: 45 minutos

#### 4.5 Views - Refresh
- [ ] `internal/tui/views/dashboard.go:Refresh()`, `refreshStatus()` - Estimativa: +2% cobertura
- [ ] `internal/tui/views/backuplist.go:Refresh()` - Estimativa: +1% cobertura
- [ ] `internal/tui/views/restore.go:Refresh()` - Estimativa: +1% cobertura
  - Cenários: dados disponíveis, diretório vazio, erros
  - Tempo estimado: 2 horas

#### 4.6 Views - Helpers
- [ ] `internal/tui/views/helpers.go:ValidateFolderPath()` - Estimativa: +1% cobertura
- [ ] `internal/tui/views/helpers.go:PlaceOverlay()` - Estimativa: +1% cobertura
  - Tempo estimado: 1 hora

**Total Fase 4**: Estimativa de +17% cobertura geral, ~11 horas de trabalho

### Fase 5: Melhorias nos Pacotes Existentes (Semana 9-10)

#### 5.1 Crypto (84.6% → 90%)
- [ ] Identificar funções não cobertas e adicionar testes
  - Estimativa: +5.4% cobertura
  - Tempo estimado: 2 horas

#### 5.2 Keyring (81.8% → 90%)
- [ ] Aumentar cobertura de casos de erro
  - Estimativa: +8.2% cobertura
  - Tempo estimado: 2 horas

#### 5.3 Git (65.1% → 90%)
- [ ] Adicionar testes para operações não cobertas
  - Estimativa: +24.9% cobertura (parcial - focar em 85%)
  - Tempo estimado: 4 horas

#### 5.4 Config (60.3% → 90%)
- [ ] `Load()`, `Save()`, `LoadFromPath()`, `SaveToPath()` - Estimativa: +15% cobertura
  - Cenários: arquivo existe, não existe, erro de parse
  - Tempo estimado: 3 horas

#### 5.5 Backup (79.5% → 90%)
- [ ] Identificar branches não cobertas
  - Estimativa: +10.5% cobertura
  - Tempo estimado: 3 horas

#### 5.6 Restore (79.0% → 90%)
- [ ] Identificar branches não cobertas
  - Estimativa: +11% cobertura
  - Tempo estimado: 3 horas

**Total Fase 5**: Estimativa de ~20% cobertura geral (distribuído entre pacotes), ~17 horas de trabalho

### Fase 6: CLI Avançado e Edge Cases (Semana 11-12)

#### 6.1 CLI Delete
- [ ] `internal/cli/delete.go:DeleteCommand()` - Estimativa: +1.5% cobertura
  - Cenários: deletar existente, inexistente, erro de permissão
  - Tempo estimado: 1.5 horas

#### 6.2 CLI Schedule (Parcial)
- [ ] `internal/cli/schedule.go:isSystemdAvailable()` - Estimativa: +0.5%
- [ ] `internal/cli/schedule.go:getUserSystemdDir()` - Estimativa: +0.5%
- [ ] `internal/cli/schedule.go:copyFile()` - Estimativa: +0.5%
  - Tempo estimado: 1.5 horas

**Nota**: Funções que executam systemctl são melhor testadas via integração/E2E

#### 6.3 Edge Cases e Error Handling
- [ ] Adicionar testes de error handling em todos os pacotes
- [ ] Testar condições de corrida onde aplicável
- [ ] Testar timeouts e cancellations
  - Tempo estimado: 4 horas

**Total Fase 6**: Estimativa de +5% cobertura geral, ~7 horas de trabalho

## Checklist de Testes por Arquivo

### `internal/cli/backup.go`
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `BackupCommand(args []string) int`
  - Cenários:
    - [ ] Backup bem-sucedido com config válida
    - [ ] Erro ao parsear flags
    - [ ] Erro ao carregar config
    - [ ] Config inválida
    - [ ] Erro ao obter senha (arquivo inexistente)
    - [ ] Erro ao obter senha (keyring vazio)
    - [ ] Backup falha (erro de I/O)
    - [ ] Backup falha (erro de criptografia)
    - [ ] Notificação habilitada e sucesso
    - [ ] Notificação habilitada e erro
    - [ ] History store disponível
    - [ ] History store indisponível

- [ ] `getPassword(passwordFile string) (string, error)`
  - Cenários:
    - [ ] Senha de arquivo (newline no final)
    - [ ] Senha de arquivo (sem newline)
    - [ ] Senha de variável de ambiente
    - [ ] Senha do keyring
    - [ ] Erro: arquivo não existe
    - [ ] Erro: arquivo não legível
    - [ ] Erro: keyring vazio

### `internal/cli/list.go`
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `ListCommand(args []string) int`
  - Cenários:
    - [ ] Erro ao parsear flags
    - [ ] Erro ao carregar config
    - [ ] Diretório de backup não existe
    - [ ] Diretório existe mas está vazio (plain)
    - [ ] Diretório existe mas está vazio (JSON)
    - [ ] Backups encontrados (plain)
    - [ ] Backups encontrados (JSON)

- [ ] `findBackups(backupDir string) ([]BackupInfo, error)`
  - Cenários:
    - [ ] Diretório não existe
    - [ ] Diretório vazio
    - [ ] Arquivos que não são .tar.gz.enc
    - [ ] Arquivo .tar.gz.enc com metadata válida
    - [ ] Arquivo .tar.gz.enc com metadata inválida
    - [ ] Arquivo .tar.gz.enc sem metadata
    - [ ] Múltiplos backups, ordenados por data

- [ ] `printBackupTable(backups []BackupInfo)`
  - Cenários:
    - [ ] Output formatado corretamente
    - [ ] Múltiplos backups
    - [ ] Tamanhos formatados corretamente

- [ ] `formatSize(size int64) string`
  - Cenários:
    - [ ] 0 bytes
    - [ ] < 1 KB
    - [ ] 1 KB
    - [ ] 1 MB
    - [ ] 1 GB
    - [ ] 1 TB

### `internal/cli/config.go`
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `ConfigCommand(args []string) int`
  - Cenários:
    - [ ] Sem subcomando (usage)
    - [ ] Subcomando get
    - [ ] Subcomando set
    - [ ] Subcomando list
    - [ ] Subcomando inválido

- [ ] `configGet(args []string) int`
  - Cenários:
    - [ ] Sem argumentos (erro)
    - [ ] Chave válida: backup_dir
    - [ ] Chave válida: git_remote
    - [ ] Chave válida: schedule
    - [ ] Chave válida: notifications
    - [ ] Chave válida: files
    - [ ] Chave válida: folders
    - [ ] Chave inválida
    - [ ] Erro ao carregar config

- [ ] `configSet(args []string) int`
  - Cenários:
    - [ ] Argumentos insuficientes (erro)
    - [ ] Chave válida com valor
    - [ ] Criar config se não existe
    - [ ] Erro ao setar valor
    - [ ] Erro ao salvar config

- [ ] `configList() int`
  - Cenários:
    - [ ] Listar todas as configurações
    - [ ] Erro ao carregar config

- [ ] `getConfigValue(cfg *config.Config, key string) (string, error)`
  - Cenários:
    - [ ] backup_dir
    - [ ] git_remote
    - [ ] schedule
    - [ ] notifications (true/false)
    - [ ] files (lista vazia e com items)
    - [ ] folders (lista vazia e com items)
    - [ ] Chave normalizada (ex: "backup-dir")
    - [ ] Chave inválida

- [ ] `setConfigValue(cfg *config.Config, key, value string) error`
  - Cenários:
    - [ ] backup_dir
    - [ ] git_remote
    - [ ] schedule
    - [ ] notifications: true
    - [ ] notifications: yes
    - [ ] notifications: 1
    - [ ] notifications: on
    - [ ] notifications: false
    - [ ] notifications: no
    - [ ] notifications: 0
    - [ ] notifications: off
    - [ ] notifications: valor inválido
    - [ ] files: lista vazia
    - [ ] files: múltiplos items
    - [ ] folders: lista vazia
    - [ ] folders: múltiplos items
    - [ ] Chave normalizada
    - [ ] Chave inválida

- [ ] `normalizeKey(key string) string`
  - Cenários:
    - [ ] Minúsculas
    - [ ] Maiúsculas
    - [ ] Mixed case
    - [ ] Com hífen
    - [ ] Com underscore

### `internal/cli/history.go`
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `HistoryCommand(args []string) int`
  - Cenários:
    - [ ] Listar histórico (plain)
    - [ ] Listar histórico (JSON)
    - [ ] Filtrar por tipo backup
    - [ ] Filtrar por tipo restore
    - [ ] Erro ao criar store
    - [ ] Erro ao ler histórico
    - [ ] Histórico vazio (plain)

- [ ] `printHistoryTable(entries []history.HistoryEntry)`
  - Cenários:
    - [ ] Output formatado
    - [ ] Entradas com erro (sem files/size)
    - [ ] Múltiplas entradas

- [ ] `formatDuration(ms int64) string`
  - Cenários:
    - [ ] 0ms
    - [ ] < 1 segundo
    - [ ] 1 segundo
    - [ ] > 1 segundo

### `internal/cli/schedule.go`
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `EnableSchedule() error`
  - Cenários (mocks necessários):
    - [ ] Systemd disponível
    - [ ] Systemd não disponível
    - [ ] Arquivos de serviço existem
    - [ ] Arquivos de serviço não existem
    - [ ] Diretório systemd criado com sucesso
    - [ ] Erro ao copiar arquivos
    - [ ] Sucesso completo

- [ ] `DisableSchedule() error`
  - Cenários similares ao EnableSchedule

- [ ] `StatusSchedule() error`
  - Cenários:
    - [ ] Timer ativo
    - [ ] Timer inativo
    - [ ] Systemd não disponível

- [ ] `ScheduleCommand(args []string) int`
  - Cenários:
    - [ ] enable
    - [ ] disable
    - [ ] status
    - [ ] subcomando inválido
    - [ ] sem subcomando

- [ ] `isSystemdAvailable() bool`
  - Cenários:
    - [ ] systemctl no PATH
    - [ ] systemctl não disponível

- [ ] `getUserSystemdDir() (string, error)`
  - Cenários:
    - [ ] Home directory disponível
    - [ ] Erro ao obter home directory

- [ ] `runSystemctl(args ...string) error`
  - Cenários:
    - [ ] Comando sucesso
    - [ ] Comando falha

- [ ] `copyFile(src, dst string) error`
  - Cenários:
    - [ ] Arquivo existe, copia sucesso
    - [ ] Arquivo não existe
    - [ ] Erro de permissão

### `internal/tui/model.go`
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `NewModel() Model`
  - Cenários:
    - [ ] Config não existe (setup mode)
    - [ ] Config existe (modo normal)
    - [ ] History store disponível
    - [ ] History store indisponível

- [ ] `Init() tea.Cmd`
  - Cenários:
    - [ ] Setup mode
    - [ ] Modo normal

- [ ] `GetConfig() *config.Config`
  - Cenários:
    - [ ] Retorna config corretamente

- [ ] `activeTabIndex() int`
  - Cenários:
    - [ ] Cada view state em tabOrder
    - [ ] View state fora de tabOrder (fallback)

- [ ] `isInputActive() bool`
  - Cenários:
    - [ ] SettingsView editing
    - [ ] SettingsView not editing
    - [ ] BackupListView creating
    - [ ] BackupListView not creating
    - [ ] RestoreView input active
    - [ ] RestoreView input inactive
    - [ ] Outras views

### `internal/tui/update.go`
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `DefaultKeyMap() KeyMap`
  - Testar criação de todas as bindings

- [ ] `propagateWindowSize(msg tea.WindowSizeMsg)`
  - Cenários:
    - [ ] Propagação para todas as views
    - [ ] Width/height negativos
    - [ ] Diferentes tamanhos

- [ ] `refreshCmdForState(state ViewState) tea.Cmd`
  - Cenários:
    - [ ] DashboardView
    - [ ] BackupListView
    - [ ] RestoreView
    - [ ] LogsView
    - [ ] SettingsView (nil)

- [ ] `Update(msg tea.Msg) (tea.Model, tea.Cmd)`
  - Cenários críticos:
    - [ ] Setup mode + WindowSizeMsg
    - [ ] Setup mode + SetupCompleteMsg
    - [ ] Setup mode + outras mensagens
    - [ ] Modo normal + WindowSizeMsg
    - [ ] Modo normal + RefreshBackupListMsg
    - [ ] Modo normal + DashboardNavigateMsg (backups)
    - [ ] Modo normal + DashboardNavigateMsg (restore)
    - [ ] Modo normal + DashboardNavigateMsg (settings)
    - [ ] Modo normal + KeyMsg Help (toggle)
    - [ ] Modo normal + KeyMsg Quit
    - [ ] Modo normal + KeyMsg Tab (não input active)
    - [ ] Modo normal + KeyMsg Tab (input active - bloqueado)
    - [ ] Modo normal + KeyMsg ShiftTab
    - [ ] Modo normal + KeyMsg number keys (1-5)
    - [ ] Modo normal + KeyMsg number keys (input active - bloqueado)
    - [ ] Modo normal + DashboardView + KeyMsg 'b'
    - [ ] Modo normal + DashboardView + KeyMsg 'r'
    - [ ] Modo normal + DashboardView + KeyMsg 's'
    - [ ] Update delegado para cada view state

### `internal/tui/view.go`
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `View() string`
  - Cenários:
    - [ ] Setup mode
    - [ ] Quitting
    - [ ] Terminal muito pequeno
    - [ ] Showing help
    - [ ] DashboardView
    - [ ] BackupListView
    - [ ] RestoreView
    - [ ] SettingsView
    - [ ] LogsView
    - [ ] Default fallback

- [ ] `currentViewHelpText() string`
  - Cenários:
    - [ ] Cada view state
    - [ ] View padrão

- [ ] `currentViewHelp() []views.HelpEntry`
  - Cenários:
    - [ ] Cada view state

### `internal/tui/help.go`
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `globalHelp() []views.HelpEntry`
  - Testar retorno de help entries

- [ ] `renderHelpOverlay(global, view []views.HelpEntry, width, height int) string`
  - Cenários:
    - [ ] Renderização básica
    - [ ] Diferentes tamanhos

### `internal/tui/styles/styles.go`
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `DefaultStyles() Styles`
  - Testar que retorna styles não vazios

- [ ] `NewListDelegate() list.DefaultDelegate`
  - Testar criação do delegate

- [ ] `NewMinimalList() list.Model`
  - Testar criação da lista com configurações corretas

## Recomendações Técnicas

### Setup de Infraestrutura de Testes

- [ ] Configurar coverage reports automatizados no CI/CD
  - Adicionar ao GitHub Actions:
    ```yaml
    - name: Run tests with coverage
      run: go test -race -coverprofile=coverage.out ./...
    
    - name: Check coverage threshold
      run: |
        go tool cover -func=coverage.out | tail -1 | awk '{print "Total coverage: " $3}'
        # Fail if coverage < 90%
    ```

- [ ] Configurar badge de cobertura no README
  - Usar codecov ou similar para badge dinâmico

- [ ] Definir threshold mínimo de cobertura por pacote:
  | Pacote | Threshold |
  |--------|-----------|
  | internal/backup | 90% |
  | internal/restore | 90% |
  | internal/crypto | 90% |
  | internal/config | 85% |
  | internal/cli | 85% |
  | internal/tui | 70% |

### Boas Práticas

1. **Estrutura de Testes**
   - Usar table-driven tests para cenários múltiplos
   - Nomear testes de forma descritiva: `TestFuncName_Scenario_ExpectedResult`
   - Agrupar testes relacionados com subtests (`t.Run()`)

2. **Mocking**
   - Usar interfaces para dependências externas (filesystem, network)
   - Para TUI BubbleTea: usar `tea.Program` com `tea.WithInput/Output` ou testar models diretamente
   - Para execução de comandos: wrap `exec.Command` em interface testável

3. **Fixtures**
   - Criar diretórios temporários para testes de I/O
   - Limpar sempre com `t.Cleanup()`
   - Usar `testing/fstest` para filesystem em memória onde possível

4. **TUI Testing**
   - Testar models individualmente (Init, Update, View)
   - Simular mensagens: `tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}`
   - Verificar output de View() com string.Contains()

5. **Cobertura de Erros**
   - Testar sempre os caminhos de erro
   - Verificar mensagens de erro específicas
   - Testar comportamento em edge cases (nil, empty, overflow)

### Ferramentas Sugeridas

#### Mocking
- `github.com/golang/mock/gomock` - Mock generator
- `github.com/stretchr/testify/mock` - Mocking manual
- Interfaces nativas do Go para dependency injection

#### Coverage
- `go test -cover` - Built-in
- `github.com/axw/gocov` - Visualização alternativa
- `github.com/AlekSi/gocov-xml` - XML output para CI

#### TUI Testing
- `github.com/charmbracelet/bubbletea` testing patterns
- Captura de stdout/stderr para testes de output

#### Test Data
- `testing/quick` - Property-based testing
- `github.com/brianvoe/gofakeit` - Geração de dados falsos

## Métricas de Acompanhamento

### Objetivos Semanais

| Semana | Cobertura Meta | Pacotes/Módulos Focus |
|--------|----------------|----------------------|
| 1 | 50% | Quick wins: formatSize, formatDuration, normalizeKey, getConfigValue, setConfigValue |
| 2 | 55% | Config functions: GetConfigDir, GetConfigPath, configList |
| 3 | 65% | CLI Backup: getPassword |
| 4 | 72% | CLI List: findBackups, ListCommand, CLI Config: configGet, configSet |
| 5 | 78% | TUI Model: NewModel, activeTabIndex, isInputActive |
| 6 | 82% | TUI Update: DefaultKeyMap, propagateWindowSize, refreshCmdForState, Update parcial |
| 7 | 85% | TUI View: View, currentViewHelpText, currentViewHelp |
| 8 | 87% | TUI Views: Todos os métodos de interface (Title, Description, FilterValue) |
| 9 | 89% | Melhorias: crypto, keyring, config (Load/Save) |
| 10 | 90% | Finalização: backup, restore, git, edge cases |
| 11 | 90%+ | CLI Delete, Schedule (parcial), revisão |
| 12 | 90%+ | Otimização, documentação, CI/CD |

### Métricas de Qualidade

Além de cobertura percentual, monitorar:

1. **Branch Coverage** (não apenas statement coverage)
   - Go não reporta branch coverage diretamente, mas analisar manualmente

2. **Test Execution Time**
   - Meta: < 30 segundos para suite completa
   - Usar `t.Parallel()` onde apropriado

3. **Test Reliability**
   - Zero flaky tests
   - Isolamento completo entre testes

4. **Documentation Coverage**
   - 100% dos testes devem ter descrição clara do cenário

## Riscos e Mitigações

### Riscos Identificados

1. **Risco**: CLI commands com múltiplas dependências difíceis de mock
   - **Impacto**: Alto
   - **Mitigação**: Refatorar para usar interfaces (ex: `PasswordGetter`, `ConfigLoader`)
   - **Timing**: Semana 3 (antes de implementar testes)

2. **Risco**: TUI BubbleTea tem curva de aprendizado para testes
   - **Impacto**: Médio
   - **Mitigação**: Estudar patterns do BubbleTea, começar com testes simples
   - **Timing**: Semana 5-6

3. **Risco**: Testes de sistema (systemd, keyring) requerem ambiente específico
   - **Impacto**: Médio
   - **Mitigação**: 
     - Separar em testes de integração
     - Usar build tags: `//go:build integration`
     - Mockar exec.Command para testes unitários
   - **Timing**: Semana 6

4. **Risco**: E2E tests podem ser lentos e instáveis
   - **Impacto**: Médio
   - **Mitigação**: 
     - Manter suite E2E separada
     - Rodar apenas em CI ou sob demanda
     - Foco em testes unitários/integração
   - **Timing**: Contínuo

5. **Risco**: Refatorações para testabilidade podem introduzir bugs
   - **Impacto**: Alto
   - **Mitigação**:
     - Fazer refatorações pequenas e isoladas
     - Sempre ter testes antes de refatorar
     - Code review rigoroso
   - **Timing**: Durante todo o processo

6. **Risco**: Complexidade excessiva de mocks
   - **Impacto**: Médio
   - **Mitigação**:
     - Preferir integração sobre mocking quando custo é baixo
     - Criar builders/fixtures reutilizáveis
   - **Timing**: Contínuo

## Conclusão

### Resumo Executivo

A aplicação dotkeeper atualmente tem **44.5% de cobertura de testes**, com grande variação entre pacotes:

- **Pacotes bem cobertos (80%+)**: pathutil (92.7%), crypto (84.6%), keyring (81.8%)
- **Pacotes medianamente cobertos (60-80%)**: backup (79.5%), restore (79.0%), history (78.5%), git (65.1%), config (60.3%)
- **Pacotes com baixa cobertura (< 60%)**: cli (10.2%), tui/views (43.9%), tui/components (55.0%)
- **Pacotes sem cobertura**: cmd/dotkeeper (0%), tui core (0%), tui/styles (0%)

### Próximos Passos Imediatos (Prioridade Alta)

1. **Esta semana**: Implementar Fase 1 (Quick Wins)
   - Começar com funções utilitárias simples
   - Estimativa: +5% cobertura em ~3.5 horas

2. **Próxima semana**: CLI Backup
   - Implementar `getPassword()` com mocks
   - Estimativa: +2% cobertura

3. **Semanas 3-4**: CLI List e Config
   - Funções principais de listagem e configuração
   - Estimativa: +10% cobertura

### Custo-Benefício

- **Investimento total estimado**: ~59 horas (8 semanas a ~7 horas/semana)
- **Ganho de cobertura**: 44.5% → 90%
- **ROI**: Código significativamente mais confiável, especialmente em:
  - Fluxo de backup (business critical)
  - Fluxo de restore (business critical)
  - TUI (experiência do usuário)

### Recomendação

Atingir 90% de cobertura é **viável e recomendado**, com foco em:
1. Código de negócio core (backup/restore/cli)
2. TUI (melhorar experiência e prevenir regressões)
3. Testes de edge cases e error handling

Não é necessário (ou prático) cobrir:
- `cmd/dotkeeper/main.go` (coberto por E2E)
- `internal/tui/styles/styles.go` (puramente declarativo)

---

**Documento criado em**: 2026-02-06  
**Última atualização**: 2026-02-06  
**Versão**: 1.0
