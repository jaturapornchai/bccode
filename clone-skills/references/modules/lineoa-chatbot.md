---
name: lineoa-chatbot
description: >
  LINE OA integration, AI Chatbot (Gemini/Ollama/ReAct Agent), document analysis, alert agent,
  knowledge base. Use this skill whenever the user mentions LINE OA, chatbot, AI agent, webhook,
  LINE notification, document analysis, Gemini, Ollama, ReAct, MCP tools for chatbot.
  Trigger keywords: "LINE OA", "chatbot", "AI agent", "webhook", "Gemini", "Ollama", "ReAct",
  "LINE notification", "document analysis", "knowledge base", "alert agent".
user-invocable: true
---

# LINE OA & AI Chatbot

## Overview
LINE Official Account integration + AI Chatbot (Gemini/Agent) + AI document analysis.

## System Structure

### LINE OA
| Section | File | Purpose |
|---------|------|---------|
| Screen | `screens/lineoa/lineoa_config_screen.dart` | LINE OA configuration |
| Config | `screens/config/line_notify_screen.dart` | LINE Notify settings |
| Backend | `backend/internal/line/` | LINE integration |
| GoAPI | `goapi/handlers/lineoa/` | LINE OA webhook + handlers |

### AI Chatbot
| Section | File | Purpose |
|---------|------|---------|
| Screen | `screens/chatbot/chatbot_screen.dart` | Chat with AI |
| Overlay | `screens/chatbot/chatbot_overlay.dart` | Floating chat popup |
| Backend | `goapi/handlers/aichat/` | Agent AI + product/customer search |
| Gemini | `goapi/gemini/` | Google Gemini integration |
| Ollama | `goapi/myollama/` | Ollama AI (local) |

### AI Document Analysis
| Section | File | Purpose |
|---------|------|---------|
| Screen | `screens/transaction/ai_analysis_screen.dart` | Document analysis |
| Backend | `/api/v1/chatbot/analyze-document` | AI reads documents |

### Alert Agent
| Section | File | Purpose |
|---------|------|---------|
| Screen | `screens/alert_agent/` | AI alert settings |

### Knowledge Base
| Section | File | Purpose |
|---------|------|---------|
| Screen | `screens/knowledge_base/` | Knowledge base for AI |

## API Endpoints

### LINE OA
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/lineoa/configs` | View all LINE OA settings |
| POST | `/api/lineoa/config/save` | Save settings |
| POST | `/api/lineoa/test` | Test connection |
| POST | `/api/lineoa/employees` | View LINE-linked employees |
| POST | `/api/lineoa/employee/add` | Add employee |
| POST | `/api/lineoa/employee/link` | Create LINE linking URL |
| POST | `/api/lineoa/webhook` | Receive webhook from LINE |
| POST | `/api/user/lineoa/callback` | Callback after linking |
| POST | `/api/user/lineoa/profile` | View LINE profile |

### AI Chatbot
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/v1/chatbot/chat-gemini` | Chat with Gemini |
| POST | `/api/v1/chatbot/chat-agent` | Chat with Agent (MCP tools) |
| POST | `/api/v1/chatbot/analyze-document` | AI document analysis |

### Unified AI
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/v1/unified/query` | Unified query via AI |
| GET | `/api/v1/unified/health` | AI health status |

## AI Providers
| Provider | Model | Used for |
|----------|-------|----------|
| Groq | Llama 4 Maverick | Chat Agent (good Thai support, limited TPM) |
| OpenRouter | nvidia/nemotron | Fallback |
| Google Gemini | Gemini | Document/image analysis |
| Ollama | local models | Offline testing |

## AI Agent Flow (ReAct Pattern)
```
User asks: "How much did we sell today?"
  |
Agent selects MCP tool: daily-sales
  |
Call MCP → Get data
  |
Agent summarizes → Reply in Thai

(Repeats up to 10 rounds max, 120 second timeout)
```

## MCP Tools for Chatbot (22 readonly tools)
- Sales: daily-sales, sales-by-date-range, top-selling-products, sales-by-seller, monthly-summary
- Products: product search, barcode lookup
- Customers: customer search
- Stock: stock balance
- Dashboard: summary data

## When to Use
- User asks about LINE OA webhook not receiving messages, or Channel Token/Secret config
- Need to add/edit AI provider (Groq, OpenRouter, Gemini, Ollama) or fallback chain
- Need to debug ReAct agent loop or MCP tool call not working in chatbot
- Questions about chatbot UI (overlay, screen) or theme exemption for chatbot
- Need to add knowledge base or alert agent for automatic notifications

## Anti-Patterns
- Never hardcode LINE webhook URL or Channel Token in source code — use config from database
- Never use a single AI provider without a fallback chain — should have at least 2 fallback levels
- Remember that chatbot UI has its own design separate from app theme — never use app theme colors
- Never add more than 22 tool calls in agent without checking provider rate limits
- Always inject system prompt with current date every time AI is called

## Important Notes
- Chatbot UI is **exempt** from theme rules — has its own independent design
- LINE OA requires Channel Access Token + Channel Secret configuration
- LINE LIFF is used for document approval within LINE app
- Agent uses ReAct pattern + automatic fallback across providers
- System prompt must be in English (token efficiency) + injects current date; AI still replies in Thai
- Knowledge Base provides additional context for AI responses

## Related Skills
- [mcp-check](../mcp-check/SKILL.md) — Verify MCP tools used by chatbot (22 readonly tools)
- [dashboard-report](../dashboard-report/SKILL.md) — Reports that chatbot pulls data from to answer
- [procurement](../procurement/SKILL.md) — PR approval via LINE OA, AI chatbot PR workflow
- [restaurant-pos](../restaurant-pos/SKILL.md) — Kitchen alert notifications via LINE OA
