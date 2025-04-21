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
- `WATCHTURM_TELEGRAM_BOT_KEY`: Telegram bot token (optional, for notifications)
- `WATCHTURM_TELEGRAM_CHAT_ID`: Telegram chat ID (optional, for notifications)

### Building and Running

```shell
# Build the application
go build ./cmd/watchturm

# Run the application
./watchturm
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
# WatchTower

**WatchTower** is an intelligent update manager for Ubuntu systems. It uses a Large Language Model (LLM) to evaluate the compatibility risk of applying system package updates — prioritising system stability and operational continuity.

---

## 🔍 What It Does

WatchTower determines whether it's safe to apply system updates automatically based on changelog analysis. It focuses on *compatibility*, not security scanning.

---

## 🧠 How It Works

1. Identifies available package updates using APT
2. Fetches the changelog for each upgradable package
3. Sends changelogs to an LLM (OpenAI) for analysis
4. The LLM evaluates the update and scores it:
   - ✅ **Low Compatibility Risk** — safe to auto-update
   - ⚠️ **Medium Risk** — may impact non-critical functionality
   - 🚨 **High Risk** — requires manual review (e.g., breaking changes)
5. Applies only low-risk updates automatically
6. Generates human-readable summaries and optional Telegram notifications

---

## 📦 Compatibility Score Criteria

The LLM scores updates based on:
- Scope and nature of changes (e.g., bugfixes vs. new features)
- Configuration or default behaviour changes
- Impact on APIs, system services, or dependencies
- Keywords indicating deprecations or breaking changes

---

## ⚙️ Usage

### ✅ Environment Variables

| Variable             | Description                             |
|----------------------|-----------------------------------------|
| `OPENAI_API_KEY`     | Required – OpenAI key for LLM access    |
| `TELEGRAM_BOT_TOKEN` | Optional – to enable Telegram messages  |
| `TELEGRAM_CHAT_ID`   | Optional – chat or channel to notify    |

### 🏗️ Build & Run

```bash
# Build
go build ./cmd/watchturm

# Run
./watchturm
```

---

## 🔄 Example Output Summary

```
📦 Update Summary - Tue, 15 Apr 2025

✅ Low Risk (Auto-Updated):
- perl 5.34.0-3ubuntu1.4 — Security fix for CVE-2024-56406
- perl-base — Same security patch
- perl-modules — Matching version and fix

⚠️ Medium Risk:
- docker.io — Updated container logic, minor config changes

🚨 High Risk:
- postgresql-13 — Major config and permission changes
```

---

## 🧪 Compatibility Score Example

```json
{
  "name": "curl",
  "compatibility_score": "low",
  "compatibility_score_reason": "Contains only minor bug fixes and documentation updates. No configuration changes or API changes."
}
```

```json
{
  "name": "postgresql-13",
  "compatibility_score": "high",
  "compatibility_score_reason": "Introduces new authentication method and changes default configuration. Manual validation recommended."
}
```

---

## ❌ What WatchTower is Not

- Not a CVE scanner or vulnerability manager
- Not a general-purpose package manager
- Not a replacement for unattended-upgrades — it builds on it with AI-based decision making

---

## 🔔 Notifications

If configured, WatchTower will send Telegram messages summarising update activity after each run, including only relevant updates.

---

## 🗃️ History & Logs

All snapshots, risk assessments, and summaries are stored locally to allow for auditing, troubleshooting, or integration into your ops pipeline.

---

## 📬 Questions or Feedback?

File an issue or open a discussion. Contributions are welcome!