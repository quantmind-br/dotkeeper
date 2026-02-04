# Draft: dotkeeper - Dotfiles Backup Manager (Go + Bubbletea)

## Requirements (confirmed)
- Linguagem: Go
- Framework TUI: Bubbletea
- Funcionalidade principal: Gerenciamento de backups de dotfiles
- Storage: Repositório GitHub local

### Decisões do Usuário (confirmadas)
- **Formato do backup**: Arquivo único compactado (.tar.gz.enc)
- **Restauração**: Seletiva (usuário escolhe o que restaurar)
- **Conflitos**: Perguntar para cada arquivo antes de sobrescrever
- **Versionamento**: Backups nomeados por data (backup-YYYY-MM-DD.enc)
- **Interface de seleção**: File browser interativo (tipo ranger/nnn)
- **Configuração**: ~/.config/dotbackup/config.yaml
- **Senha**: Solicitar a cada operação (não armazena)
- **Backup automático**: Sim, com agendamento (daemon/serviço)
- **Setup Git**: Wizard interativo na TUI
- **Agendamento**: Systemd timer (unit files)
- **Notificações**: Desktop notifications (sucesso/falha)
- **Detecção de mudanças**: Indicador visual na TUI (quais arquivos mudaram)
- **Nome da aplicação**: dotkeeper
- **Modo de operação**: TUI interativa + CLI (comandos como `dotkeeper backup`)

### Telas da TUI (confirmadas)
1. Dashboard/Home - Status geral, último backup, próximo agendado
2. File Browser - Navegar e selecionar arquivos/pastas
3. Lista de Backups - Histórico com datas
4. Restauração - Selecionar backup e itens para restaurar
5. Configurações - Repo, agendamento, preferências
6. Logs/Histórico - Logs de operações anteriores

### Decisões Técnicas (confirmadas)
- **Criptografia**: AES-256-GCM + Argon2id (key derivation)
- **Escopo MVP**: Completo (todas features de uma vez)
- **Profiles**: Profile único (uma lista de arquivos, um repo)
- **Symlinks**: Seguir e copiar conteúdo (não preservar links)
- **Testes**: TDD (Test-Driven Development)
- **Idioma da UI**: Inglês

### Exclusões de Escopo (NÃO implementar)
- Suporte a Windows (apenas Linux)
- Múltiplos repositórios remotos (apenas um remote)
- macOS scheduling (apenas Linux com systemd)
- Templates/transforms de arquivos (copia as-is)
- Backups incrementais (sempre full backup)
- Sistema de plugins/hooks
- Auto-descoberta de dotfiles
- Diff viewer entre backups
- Provedores de cloud storage (apenas git)

### Decisões do Metis Gap Analysis
- **Senha + Agendamento**: Integração com Keyring do sistema (gnome-keyring/KWallet)
- **macOS Scheduling**: Fora do escopo (apenas Linux systemd)
- **Conflitos na Restauração**: Renomear existente (.bak) + Ver diff antes de decidir
- **CLI Password**: DOTKEEPER_PASSWORD env var
- **Arquivos Grandes**: Sem limite, usar streaming

## Technical Decisions
- [pending] Algoritmo de criptografia
- [pending] Formato do backup (arquivo único vs estrutura)
- [pending] Armazenamento de configuração

## Research Findings

### Dotfiles Tools Existentes (chezmoi, yadm, dotbot)
- **chezmoi**: Source directory separado, templates Go, encryption per-file (age/gpg), 14+ password managers
- **yadm**: Git wrapper, alternates por OS/hostname, archive-based encryption
- **dotbot**: Bootstrapper leve, YAML config, sem encryption built-in

**Padrões importantes identificados:**
- Idempotência (operações seguras para repetir)
- Preview/diff antes de aplicar
- Bootstrap em um único comando
- Separação entre source state e target
- Suporte a múltiplas máquinas

### Encryption Patterns (pesquisa completa)
- **Algoritmo recomendado**: AES-256-GCM (AEAD - autenticação + criptografia)
- **KDF recomendado**: Argon2id (resistente a GPU/ASIC)
- **Parâmetros Argon2id**: Time=3, Memory=64MB, Threads=4, KeyLen=32
- **Nonce**: 12 bytes random, prepend ao ciphertext
- **Salt**: 16 bytes random, único por derivação de chave
- **Referências de produção**: restic, ethereum/go-ethereum, kubernetes, portainer

**Fluxo de criptografia:**
1. Read file → SHA-256 checksum
2. Argon2id(password + salt) → key
3. AES-256-GCM(key, nonce, plaintext) → ciphertext
4. Prepend metadata header → encrypted file

### Bubbletea Patterns (pesquisa completa)
- **Multi-view**: State-based view switching (enums/flags para rotas)
- **File picker**: Usar bubbles/filepicker component
- **Forms**: bubbles/textinput com validadores
- **Git integration**: tea.Cmd para operações async, tea.ExecProcess para comandos interativos
- **Project structure**: cmd/, internal/ui/views/, internal/ui/components/, internal/git/

**Componentes recomendados (Bubbles):**
- `bubbles/filepicker` - navegação de arquivos
- `bubbles/list` - listas com fuzzy search
- `bubbles/textinput` - inputs de formulário
- `bubbles/key` - key bindings configuráveis

## Core Flow (descrito pelo usuário)
1. Usuário define lista de arquivos e pastas para backup
2. Aplicação criptografa tudo com senha definida pelo usuário
3. Armazena em pasta com repositório GitHub
4. Faz commit e push para repositório remoto
5. Opção de restauração disponível

## Open Questions
- Formato do backup: arquivo único compactado ou estrutura de arquivos?
- Restauração: total ou seletiva?
- Versionamento: manter histórico ou apenas backup mais recente?
- Conflitos na restauração: sobrescrever ou perguntar?
- Arquivos sensíveis: tratamento especial?
- Configuração da lista de arquivos: onde persistir?

## Scope Boundaries
- INCLUDE: [a definir]
- EXCLUDE: [a definir]
