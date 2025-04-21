# Wachturm

**Wachturm** is an intelligent update manager for Ubuntu systems. It uses a Large Language Model (LLM) to evaluate the compatibility risk of applying system package updates — prioritising system stability and operational continuity.

---

## 🔍 What It Does

Wachturm determines whether it's safe to apply system updates automatically based on changelog analysis. It focuses on *compatibility*, not security scanning.

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
| `WACHTURM_TELEGRAM_BOT_TOKEN` | Optional – to enable Telegram messages  |
| `WACHTURM_TELEGRAM_CHAT_ID`   | Optional – chat or channel to notify    |

### 🏗️ Build & Run

```bash
# Build
make build

# Run
./wachturm
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

## ❌ What Wachturm is Not

- Not a CVE scanner or vulnerability manager
- Not a general-purpose package manager
- Not a replacement for unattended-upgrades — it builds on it with AI-based decision making

---

## 🔔 Notifications

If configured, Wachturm will send Telegram messages summarising update activity after each run, including only relevant updates.

---

## 🗃️ History & Logs

All snapshots, risk assessments, and summaries are stored locally to allow for auditing, troubleshooting, or integration into your ops pipeline.

---

## 📬 Questions or Feedback?

File an issue or open a discussion. Contributions are welcome!