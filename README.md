# WatchTower

WatchTower is a tool that intelligently manages system package updates by assessing the risk of applying updates using Large Language Models (LLMs).

## Main Feature: AI-Powered Update Risk Assessment

The core functionality of WatchTower is using an LLM to evaluate package changelogs and determine the risk of updating each package. The risk assessment focuses on the likelihood of breaking the system, not the security properties of the update.

### How It Works

1. WatchTower identifies available package updates on the system
2. For each update, it fetches the changelog
3. If a changelog is available, it sends the changelog to an LLM (OpenAI)
4. The LLM analyzes the changelog content and assesses the risk as:
   - **Low Risk**: Safe to auto-update (e.g., documentation changes, minor bug fixes)
   - **Medium Risk**: Exercise caution (e.g., changes to non-critical functionality)
   - **High Risk**: Manual review recommended (e.g., major changes, configuration changes)
5. Only low-risk updates are automatically applied
6. If no changelog is available, the update is not applied

### Risk Assessment Criteria

The LLM evaluates updates based on factors like:

- Type of changes (bug fixes, feature additions, refactoring)
- Scope of changes (localized vs. system-wide)
- Configuration file changes
- API/interface changes
- Dependencies affected
- Breaking changes mentioned in changelog

## Usage

### Environment Variables

- `OPENAI_API_KEY`: API key for OpenAI (required)
- `TELEGRAM_BOT_TOKEN`: Telegram bot token (optional, for notifications)
- `TELEGRAM_CHAT_ID`: Telegram chat ID (optional, for notifications)

### Building and Running

```shell
# Build the application
go build ./cmd/watchtower

# Run the application
./watchtower
```

## Features

- AI-powered risk assessment of package updates
- Only applies updates deemed low-risk
- Generates detailed reports on available updates
- Creates human-readable summaries
- Sends notifications via Telegram (optional)
- Maintains history of updates and risk assessments

## Output Examples

### Update Summary Example

```
Update Summary - Mon, 19 Aug 2023 12:34:56 UTC

Total packages available for update: 15
High risk updates: 2
Medium risk updates: 5
Low risk updates: 8

High risk packages (manual review recommended):
- openssh-server
- postgresql-13

Medium risk packages (caution advised):
- nginx
- nodejs
- python3.10
- docker.io
- systemd

Low risk packages will be updated automatically.
```

### Risk Assessment Example

For a package like `curl`:

```json
{
  "name": "curl",
  "risk_level": "low",
  "risk_reason": "Contains only minor bug fixes and documentation updates. No configuration changes or API changes."
}
```

For a package like `openssh-server`:

```json
{
  "name": "openssh-server",
  "risk_level": "high",
  "risk_reason": "Contains significant changes to default configuration behavior and authentication mechanisms. May require manual configuration updates."
}
```